package main

import (
	"github.com/corrupt/anaconda"
	rss "github.com/jteeuwen/go-pkg-rss"
	"golang.org/x/net/html/charset"
	"html"
	"io"
	"log"
	"os"
	"time"
)

var (
	api    = anaconda.NewTwitterApi(AccessToken, AccessTokenSecret)
	logger *log.Logger
)

func main() {
	anaconda.SetConsumerKey(ConsumerKey)
	anaconda.SetConsumerSecret(ConsumerSecret)

	logFile, err := os.Create(LogFile)
	if err != nil {
		log.Println("Could not open log file")
		os.Exit(1)
	}
	logger = log.New(io.MultiWriter(logFile, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)

	err = dbInit()
	if err != nil {
		logger.Println(err)
		os.Exit(1)
	}

	go updateConfiguration()
	PollFeed(RSSFeedLocation, 5)
}

func PollFeed(uri string, timeout int) {
	feed := rss.New(timeout, true, chanHandler, itemHandler)

	for {
		logger.Println("Fetching RSS Feed")
		if err := feed.Fetch(uri, getCharsetReader); err != nil {
			logger.Printf("[e] %s: %s", uri, err)
		}

		<-time.After(time.Duration(feed.SecondsTillUpdate() * 1e9))
	}
}

func getCharsetReader(contentType string, input io.Reader) (io.Reader, error) {
	reader, err := charset.NewReader(input, contentType)
	return reader, err
}

func chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	//No need for a channel handler
}

func itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	for _, item := range newitems {
		logger.Println("Received '" + item.Title + "'")
		unescapedTitle := html.UnescapeString(item.Title)
		tweet := Tweet{item.Links[0].Href, shortenTweet(unescapedTitle), *(item.Guid)}
		tweetHandler(tweet)
	}
	//databaseCleanup(
}

func tweetHandler(tweet Tweet) (err error) {

	twt, err := getTweetByGuid(tweet.guid)
	if err != nil {
		logger.Println(err)
	}
	if twt != nil && twt.guid == tweet.guid {
		logger.Println("\tTweet '" + tweet.headline + "' is already cached")
	} else {
		logger.Println("\tTweeting '" + tweet.headline + "'")
		_, err = api.PostTweet(tweet.headline+"\n"+tweet.url, nil)
		if err != nil {
			logger.Println(err)
		} else {
			err = insertTweet(tweet)
			if err != nil {
				logger.Println(err)
			}
		}
	}
	return nil
}

func shortenTweet(tweet string) string {
	txtlen := 280 - LinkLength - 1 //tweet length - t.co link length - \n
	if len(tweet) > txtlen {
		return tweet[:txtlen-3] + "..."
	}
	return tweet
}

func updateConfiguration() {
	for {
		logger.Println("Updating Twitter configuration cache")
		config, err := api.GetConfiguration(nil)
		if err != nil {
			logger.Println(err)
		} else {
			LinkLength = config.ShortUrlLengthHttps
		}
		<-time.After(24 * time.Hour)
	}
}
