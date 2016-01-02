# ltt - listentothis

Batch download songs posted to [/r/listentothis](https://reddit.com/r/listentothis).

Needs a recent version of youtube-dl in your $PATH. Recommend
`pip install youtube-dl` in a virtualenv. Older versions get a 403 on a lot of
the songs.
 
# Build

`gb build`

Of course, you need [gb](https://getgb.io) for that.

# Run

`bin/ltt` scrapes the feed and downloads songs to `~/Music/listentothis`.
Throw this in a cronjob, and filter feed on music like a sponge as it
drifts by.

