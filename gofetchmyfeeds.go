package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/SlyMarbo/rss"
	"github.com/sgoertzen/html2text"
	"github.com/syndtr/goleveldb/leveldb"
	"gopkg.in/gomail.v2"
)

var (
	err        error
	logf       *os.File          // log file
	conf       config            // TOML config
	mailsender gomail.SendCloser // Mail sender
	db         *leveldb.DB       // database for seen items
	feed       *rss.Feed         // RSS feed
)

// Default configuration parameters.
const (
	dMailHost      = "localhost"
	dMailPort      = 25
	dMailRecipient = "root"
)

// TOML configuration structures.
type mailConfig struct {
	Host      string `toml:"host,omitempty"`
	Port      int    `toml:"port,omitempty"`
	User      string `toml:"user,omitempty"`
	Password  string `toml:"password,omitempty"`
	Recipient string `toml:"recipient,omitempty"`
}
type rssConfig struct {
	Feeds [][]string `toml:"feeds"`
}
type config struct {
	MailConfig mailConfig `toml:"mail"`
	RssConfig  rssConfig  `toml:"rss"`
}

// Mail body builder.
func buildBody(i *rss.Item) string {
	var t string
	b := `
	link: %s

	%s
	`
	// HTML to text conversion
	if t, err = html2text.Textify(i.Summary); err != nil {
		log.Error("can not convert body to html")
		t = i.Summary
	}
	return fmt.Sprintf(b, i.Link, t)
}

// Mail sender.
func sendMail(i *rss.Item, ftag string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "gofetchmyfeeds@nowhere.com")
	m.SetAddressHeader("To", conf.MailConfig.Recipient, conf.MailConfig.Recipient)
	m.SetHeader("Subject", ftag+" "+i.Title)
	m.SetBody("text/plain", buildBody(i))
	if err = gomail.Send(mailsender, m); err != nil {
		return err
	}
	return nil
}

// Store the item to the database.
func markSeen(i *rss.Item) error {
	// We use the item Link as a key.
	// No need for a value.
	return db.Put([]byte(i.Link), []byte(""), nil)
}

func main() {
	// Getting the program parameters.
	debug := flag.Bool("debug", false, "debug (verbose log), default is error")
	logfile := flag.String("logfile", "", "log to the given file")
	flag.Parse()

	// Logging to file if logfile parameter specified.
	if *logfile != "" {
		if logf, err = os.OpenFile(*logfile, os.O_WRONLY|os.O_CREATE, 0755); err != nil {
			log.Panic(err)
		} else {
			log.SetOutput(logf)
		}
	}
	// Setting the log level.
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Info("Opening the configuration file.")
	// Decoding the configuration.
	var md toml.MetaData
	if md, err = toml.DecodeFile("./configuration.toml", &conf); err != nil {
		log.Fatal("error opening the configuration file:", err)
	}
	// Setting default parameters if not defined in the toml file.
	if !md.IsDefined("mail", "host") {
		conf.MailConfig.Host = dMailHost
	}
	if !md.IsDefined("mail", "port") {
		conf.MailConfig.Port = dMailPort
	}
	if !md.IsDefined("mail", "recipient") {
		conf.MailConfig.Recipient = dMailRecipient
	}
	log.WithFields(log.Fields{
		"c.MailConfig.Host":      conf.MailConfig.Host,
		"c.MailConfig.Port":      conf.MailConfig.Port,
		"c.MailConfig.User":      conf.MailConfig.User,
		"c.MailConfig.Password":  conf.MailConfig.Password,
		"c.MailConfig.Recipient": conf.MailConfig.Recipient,
		"c.RssConfig.Feeds":      conf.RssConfig.Feeds,
	}).Debug("main:configuration")

	// Opening the DB file.
	if db, err = leveldb.OpenFile("./seen.db", nil); err != nil {
		log.Fatal("error opening the database file:", err)
	}
	defer func() {
		err = db.Close()
		if err != nil {
			log.Error("error closing the database file:", err)
		}
	}()
	// Connecting to the remote SMTP server.
	dial := gomail.NewDialer(conf.MailConfig.Host, conf.MailConfig.Port, conf.MailConfig.User, conf.MailConfig.Password)
	if mailsender, err = dial.Dial(); err != nil {
		log.Fatal("can not open connection to mail server:", err)
	}
	defer func() {
		err = mailsender.Close()
		if err != nil {
			log.Error("error closing the mailsender connection:", err)
		}
	}()

	// Gathering the feeds from the configuration file.
	for _, fconfig := range conf.RssConfig.Feeds {
		// Fetching the feed.
		furl := fconfig[0]
		ftag := fconfig[1]
		log.Info("Fetching " + furl)
		if feed, err = rss.Fetch(furl); err != nil {
			log.Errorf("error fetching %s: %v", furl, err)
			continue
		}
		log.Debugf("fetched:%s", feed.Title)
		// Parsing the feed items.
		for _, item := range feed.Items {
			log.Infof("- %s", item.Title)
			log.Debugf("item:%s link:", item.Title, item.Link)
			// Looking for the feed into the DB - the link is the key
			if _, err = db.Get([]byte(item.Link), nil); err != nil {
				switch err {
				case leveldb.ErrNotFound:
					log.Debugf("NOT already sent")
					// Sending the item by mail.
					if err = sendMail(item, ftag); err != nil {
						log.Error("error sending mail:", err)
					} else if err = markSeen(item); err != nil {
						log.Error("error marking item as seen:", err)
					}
					// Storing the item into the DB.
				default:
					log.Errorf("error getting key %s from the DB: %v", feed.Link, err)
					continue
				}
			} else {
				log.Debug("already sent")
			}
		}
	}
}
