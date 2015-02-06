package main

import (
	"fmt"
	//"github.com/ChimeraCoder/anaconda"
	//"errors"
	"github.com/corrupt/anaconda"
	rss "github.com/jteeuwen/go-pkg-rss"
	"golang.org/x/net/html/charset"
	"io"
	"log"
	"os"
	"time"
)

var api = anaconda.NewTwitterApi(AccessToken, AccessTokenSecret)

func main() {
	anaconda.SetConsumerKey(ConsumerKey)
	anaconda.SetConsumerSecret(ConsumerSecret)

	go updateConfiguration()
	err := dbInit()
	if err != nil {
		fmt.Println(err)
	}

	PollFeed(RSSFeedLocation, 5)
}

func PollFeed(uri string, timeout int) {
	feed := rss.New(timeout, true, chanHandler, itemHandler)

	for {
		if err := feed.Fetch(uri, getCharsetReader); err != nil {
			fmt.Fprintf(os.Stderr, "[e] %s: %s", uri, err)
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
	//fmt.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
}

func itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	for _, item := range newitems {
		tweet := Tweet{item.Links[0].Href, shortenTweet(item.Title)}
		tweetHandler(tweet)
	}
}

func tweetHandler(tweet Tweet) (err error) {

	twt, err := getTweetByUrl(tweet.url)
	if err != nil {
		fmt.Println(err)
	}
	if twt != nil && twt.url == tweet.url {
		//return errors.New("Tweet '" + tweet.headline + "' is already cached")
		log.Println("Tweet '" + tweet.headline + "' is already cached")
	} else {
		_, err = api.PostTweet(tweet.headline+"\n"+tweet.url, nil)
		if err != nil {
			fmt.Println(err)
		} else {
			err = insertTweet(tweet)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func shortenTweet(tweet string) string {
	txtlen := 140 - linklength - 1 //tweet length - t.co link length - \n
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
			fmt.Println(err)
		} else {
			linklength = config.ShortUrlLengthHttps
		}
		<-time.After(24 * time.Hour)
	}
}
