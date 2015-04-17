# tw-delete

Deletes all tweets sent after a duration.

Put credentials in a file (or pass on command line, see `--help`) like,

``` toml
consumerKey = "..."
consumerSecret = "..."
accessToken = "..."
accessSecret = "..."
```

Then,

``` bash
$ go get github.com/hawx/tw-delete
$ tw-delete --auth auth.conf --after 72h --save ./someplace
...
```

If `--save` is given the deleted tweets are saved in the folder as `.json`
documents in folders named with the tweetid, any associated media is also saved
in the folder.
