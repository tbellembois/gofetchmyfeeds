# GoFetchMyFeeds

```bash
    $ go get -u github.com/tbellembois/gofetchmyfeeds
```

You can also download a binary in the [releases](https://github.com/tbellembois/gofetchmyfeeds/releases) section.

Create a `configuration.toml`:
```
[mail]
    host = "smtp.gmail.com"
    port = 587
    recipient = "foo@gmail.com"
    user = "foo@gmail.com"
    password = "supersecret"

[rss]
        feeds = [ 
                [ "http://www.howtoforge.com/rss/linux/debian.rss", "-HOWTOFORGE-" ],
                [ "http://www.futura-sciences.com/rss/high-tech/dossiers.xml", "-FUTURA-" ],
                [ "http://www.futura-sciences.com/rss/high-tech/actualites.xml", "-FUTURA-" ],
                ['http://downloads.bbc.co.uk/podcasts/worldservice/scia/rss.xml', '-BBC-'],
                ['http://downloads.bbc.co.uk/podcasts/worldservice/docarchive/rss.xml', '-BBC-'],
                ]
```

Run the program:
```
    $ gofetchmyfeeds [-logfile /path/to/gofetchmyfeeds.log] [-debug]
```
