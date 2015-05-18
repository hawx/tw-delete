# tw-delete

Deletes all tweets sent after a duration.

Put credentials in a file at `~/.config/tw-delete/auth`,

``` toml
consumerKey = "..."
consumerSecret = "..."
accessToken = "..."
accessSecret = "..."
```

Then,

``` bash
$ go get hawx.me/code/tw-delete
$ tw-delete --after 72h --save ./someplace
...
```

If `--save` is given the deleted tweets are saved in the folder as `.json`
documents in folders named with the tweetid, any associated media is also saved
in the folder.
