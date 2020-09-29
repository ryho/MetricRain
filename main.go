package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ChimeraCoder/anaconda"
)

const (
	HisScreenName = "SummerhillRain"
	MyScreenName  = "MetricMartin"
	fetchLimit    = "100"
	cmPerInch     = 2.54

	matchBasicTweetRegexPattern    = "^([\\d]*\\.[\\d]*) inches(.*)"
	matchAdvancedTweetRegexPattern = "^[\\d\\/]+: ([\\d]*\\.[\\d]*) inches$"

	TESTING = false

	TWITTER_CONSUMER_KEY        = "TWITTER_CONSUMER_KEY"
	TWITTER_CONSUMER_SECRET     = "TWITTER_CONSUMER_SECRET"
	TWITTER_ACCESS_TOKEN        = "TWITTER_ACCESS_TOKEN"
	TWITTER_ACCESS_TOKEN_SECRET = "TWITTER_ACCESS_TOKEN_SECRET"
)

var api *anaconda.TwitterApi
var matchBasicTweetRegex, matchAdvancedTweetRegex *regexp.Regexp

func init() {
	anaconda.SetConsumerKey(os.Getenv(TWITTER_CONSUMER_KEY))
	anaconda.SetConsumerSecret(os.Getenv(TWITTER_CONSUMER_SECRET))
	api = anaconda.NewTwitterApi(os.Getenv(TWITTER_ACCESS_TOKEN), os.Getenv(TWITTER_ACCESS_TOKEN_SECRET))
	var err error
	matchBasicTweetRegex, err = regexp.Compile(matchBasicTweetRegexPattern)
	if err != nil {
		panic(err)
	}
	matchAdvancedTweetRegex, err = regexp.Compile(matchAdvancedTweetRegexPattern)
	if err != nil {
		panic(err)
	}
}

func main() {
}

// HandleRequest is called by Google Cloud Functions when a webhook is received
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	err := RunAJob()
	if err != nil {
		fmt.Println("Job Failed with error %v", err)
	}
}

func RunAJob() error {
	repliedTweets := map[string]bool{}

	myTweets, err := GetTheirTweets(MyScreenName)
	if err != nil {
		return err
	}
	for _, myTweet := range myTweets {
		if myTweet.InReplyToScreenName == HisScreenName {
			repliedTweets[myTweet.InReplyToStatusIdStr] = true
		}
	}

	hisTweets, err := GetTheirTweets(HisScreenName)
	if err != nil {
		return err
	}

	for i, hisTweet := range hisTweets {
		//for i := len(hisTweets)-1; i>=0; i-- {
		//	hisTweet := hisTweets[i]
		if repliedTweets[hisTweet.IdStr] == true {
			// If I have replied to the Tweet, stop looking for more Tweets to reply to
			break
		}
		if hisTweet.InReplyToStatusIdStr != "" {
			// Skip tweets where he replies to someone else
			continue
		}
		fmt.Println(i, hisTweet.CreatedAt, hisTweet.FullText)
		inchValue, suffix, ok := parseTweetToInches(hisTweet.FullText)
		if !ok {
			continue
		}
		fmt.Print(inchValue, " in \n")

		cmText := convertInchesToCentimetersText(inchValue)
		message := "@" + HisScreenName + " " + cmText + suffix
		err = PostATweet(hisTweet.IdStr, message)
		if err != nil {
			return err
		}
		break
	}
	return nil
}

func convertInchesToCentimetersText(inches float64) string {
	cm := inches * cmPerInch
	return fmt.Sprintf("%.2f cm \n", cm)
}

func parseTweetToInches(text string) (float64, string, bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		fmt.Printf("Empty tweet\n")
		return 0, "", false
	}
	if text == "Trace" {
		fmt.Printf("Trace tweet\n")
		return 0, "", false
	}
	matches := matchBasicTweetRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		var suffix string
		if len(matches) > 2 {
			suffix = matches[2]
		}
		value, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			fmt.Println(err)
			return 0, "", false
		}
		return value, suffix, true
	}
	matches = matchAdvancedTweetRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		if len(matches) > 2 {
			fmt.Printf("matchAdvancedTweetRegex returned %v matches: %v\n", len(matches), PrettyPrint(matches))
		}
		value, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			fmt.Println(err)
			return 0, "", false
		}
		return value, "", true
	}

	fmt.Printf("Could not parse tweet %v\n", text)
	return 0, "", false
}

func PostATweet(inReplyTo, text string) error {
	req := url.Values{}
	fmt.Println("PostATweet", inReplyTo, text)
	req.Set("in_reply_to_status_id", inReplyTo)
	if !TESTING {
		tweet, err := api.PostTweet(text, req)
		if err != nil {
			return err
		}
		fmt.Println("PostATweet sent tweet ", tweet)
	}
	return nil
}

func PrintTheirTweets(screenName string) {
	tweets, err := GetTheirTweets(screenName)
	if err != nil {
		panic(err)
	}
	for i, tweet := range tweets {
		//fmt.Println(PrettyPrint(tweet))
		fmt.Println(i)
		fmt.Println(tweet.Text)
		fmt.Println(tweet.FullText)
	}
}

func GetTheirTweets(screenName string) ([]anaconda.Tweet, error) {
	req := url.Values{}
	req.Set("screen_name", screenName)
	req.Set("count", fetchLimit)
	return api.GetUserTimeline(req)
}

func PrettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "")
	if err != nil {
		return ""
	}
	return string(b)
}

func DeleteMyTweets() {
	myTweets, err := GetTheirTweets(MyScreenName)
	if err != nil {
		panic(err)
	}
	for _, myTweet := range myTweets {
		_, err = api.DeleteTweet(myTweet.Id, true)
		if err != nil {
			panic(err)
		}
	}
}
