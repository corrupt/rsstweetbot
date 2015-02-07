package main

import (
	"github.com/corrupt/anaconda"
	rss "github.com/jteeuwen/go-pkg-rss"
	"golang.org/x/net/html/charset"
	"io"
	"log"
	"time"
)

var api = anaconda.NewTwitterApi(AccessToken, AccessTokenSecret)

func main() {
	anaconda.SetConsumerKey(ConsumerKey)
	anaconda.SetConsumerSecret(ConsumerSecret)

	go updateConfiguration()
	err := dbInit()
	if err != nil {
		log.Println(err)
	}

	PollFeed(RSSFeedLocation, 5)
}

func PollFeed(uri string, timeout int) {
	feed := rss.New(timeout, true, chanHandler, itemHandler)

	for {
		log.Println("Fetching RSS Feed")
		if err := feed.Fetch(uri, getCharsetReader); err != nil {
			log.Printf("[e] %s: %s", uri, err)
			return
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
		log.Println("\tReceived '" + item.Title + "'")
		tweet := Tweet{item.Links[0].Href, shortenTweet(item.Title)}
		tweetHandler(tweet)
	}
}

func tweetHandler(tweet Tweet) (err error) {

	twt, err := getTweetByUrl(tweet.url)
	if err != nil {
		log.Println(err)
	}
	if twt != nil && twt.url == tweet.url {
		log.Println("Tweet '" + tweet.headline + "' is already cached")
	} else {
		log.Println("Tweeting '" + tweet.headline + "'")
		_, err = api.PostTweet(tweet.headline+"\n"+tweet.url, nil)
		if err != nil {
			log.Println(err)
		} else {
			err = insertTweet(tweet)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return nil
}

func shortenTweet(tweet string) string {
	txtlen := 140 - LinkLength - 1 //tweet length - t.co link length - \n
	if len(tweet) > txtlen {
		return tweet[:txtlen-3] + "..."
	}
	return tweet
}

func updateConfiguration() {
	for {
		log.Println("Updating Twitter configuration cache")
		config, err := api.GetConfiguration(nil)
		if err != nil {
			log.Println(err)
		} else {
			LinkLength = config.ShortUrlLengthHttps
		}
		<-time.After(24 * time.Hour)
	}
}
