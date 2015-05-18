package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ChimeraCoder/anaconda"
	"hawx.me/code/xdg"
)

var (
	auth     = flag.String("auth", "", "")
	after    = flag.String("after", "120h", "")
	save     = flag.String("save", "", "")
	noDelete = flag.Bool("no-delete", false, "")
	help     = flag.Bool("help", false, "")
)

const HELP = `Usage: tw-delete [options]

  Deletes old tweets. Note: If --save is not given data is not saved!

    --auth PATH         # Path to file with auth details
    --after DUR         # Duration to delete after (default: '120h')
    --save DIR          # Directory to save tweets to
    --no-delete         # Don't delete tweets
    --help              # Display this help message
`

type Saver interface {
	Save(tweet anaconda.Tweet) error
}

type Deleter interface {
	DeleteTweet(id int64, trimUser bool) (tweet anaconda.Tweet, err error)
}

type emptyDeleter struct{}

func (_ *emptyDeleter) DeleteTweet(_ int64, _ bool) (anaconda.Tweet, error) {
	return anaconda.Tweet{}, nil
}

type emptySaver struct{}

func (_ *emptySaver) Save(_ anaconda.Tweet) error {
	return nil
}

type fileSaver struct {
	loc string
}

func (s *fileSaver) Save(tweet anaconda.Tweet) (err error) {
	tweetLoc := filepath.Join(s.loc, tweet.IdStr)

	err = os.Mkdir(tweetLoc, 0755)
	if err != nil {
		return
	}

	data, err := json.Marshal(tweet)
	if err != nil {
		return
	}

	log.Println("writing:", tweetLoc)
	err = ioutil.WriteFile(tweetLoc+"/data.json", data, 0644)
	if err != nil {
		return
	}

	for _, media := range tweet.Entities.Media {
		resp, err := http.Get(media.Media_url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		mediaLoc := filepath.Join(tweetLoc, media.Id_str+filepath.Ext(media.Media_url))
		log.Println("writing:", mediaLoc)

		file, err := os.OpenFile(mediaLoc, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		io.Copy(file, resp.Body)
	}

	return
}

func main() {
	flag.Parse()

	if *help {
		fmt.Println(HELP)
		return
	}

	authPath := xdg.Config("tw-delete/auth")
	if *auth != "" {
		authPath = *auth
	}

	var conf struct {
		ConsumerKey, ConsumerSecret, AccessToken, AccessSecret string
	}

	if _, err := toml.DecodeFile(authPath, &conf); err != nil {
		log.Fatal(err)
	}

	anaconda.SetConsumerKey(conf.ConsumerKey)
	anaconda.SetConsumerSecret(conf.ConsumerSecret)
	api := anaconda.NewTwitterApi(conf.AccessToken, conf.AccessSecret)

	duration, err := time.ParseDuration(*after)
	if err != nil {
		log.Fatal(err)
	}

	var saver Saver = &emptySaver{}
	if *save != "" {
		saver = &fileSaver{loc: *save}
	}

	var deleter Deleter = api
	if *noDelete {
		deleter = &emptyDeleter{}
	}

	var maxId int64 = -1

	for {
		v := url.Values{}
		v.Add("count", "200")
		if maxId > 0 {
			v.Add("max_id", strconv.FormatInt(maxId, 10))
		}

		log.Println("getting", maxId)
		timeline, err := api.GetUserTimeline(v)
		if err != nil {
			log.Fatal(err)
		}

		if len(timeline) == 0 {
			break
		}

		for _, tweet := range timeline {
			maxId = tweet.Id
			t, _ := tweet.CreatedAtTime()

			if t.Add(duration).Before(time.Now()) {
				deleteTweet(tweet, deleter, saver)
			}
		}
	}
}

func deleteTweet(tweet anaconda.Tweet, deleter Deleter, saver Saver) {
	if err := saver.Save(tweet); err != nil {
		// if we wanted to save but couldn't don't delete
		return
	}

	if _, err := deleter.DeleteTweet(tweet.Id, false); err != nil {
		log.Fatal(err)
	}
}
