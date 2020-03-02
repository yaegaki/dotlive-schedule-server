package main

import (
	"net/url"

	"github.com/ChimeraCoder/anaconda"
)

func getTimeline(api *anaconda.TwitterApi, screenName, lastTweetID string) ([]anaconda.Tweet, error) {
	param := url.Values{
		"screen_name":     []string{screenName},
		"exclude_replies": []string{"true"},
		"include_rts":     []string{"false"},
	}
	if lastTweetID != "" {
		param["since_id"] = []string{lastTweetID}
	}

	return api.GetUserTimeline(param)
}
