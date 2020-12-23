package tweet

import (
	"errors"
	"net/url"

	"github.com/ChimeraCoder/anaconda"
	"github.com/yaegaki/dotlive-schedule-server/jst"
)

// GetTimeline タイムラインを取得する
func GetTimeline(api *anaconda.TwitterApi, screenName, lastTweetID string) ([]Tweet, error) {
	return getTimeline(api, screenName, lastTweetID, "")
}

// GetTimelineWithMaxID タイムラインを取得する(MaxID以下のIDを持つツイートのみを対象にする)
func GetTimelineWithMaxID(api *anaconda.TwitterApi, screenName, lastTweetID string, maxID string) ([]Tweet, error) {
	return getTimeline(api, screenName, lastTweetID, maxID)
}

func getTimeline(api *anaconda.TwitterApi, screenName, lastTweetID string, maxID string) ([]Tweet, error) {
	param := url.Values{
		"screen_name":     []string{screenName},
		"exclude_replies": []string{"false"},
		"include_rts":     []string{"false"},
	}
	if lastTweetID != "" {
		param["since_id"] = []string{lastTweetID}
	}

	if maxID != "" {
		param["max_id"] = []string{maxID}
	}

	timeline, err := api.GetUserTimeline(param)
	if err != nil {
		return nil, err
	}

	var result []Tweet
	for _, t := range timeline {
		tweet, err := tweetToTweet(t)
		if err != nil {
			return nil, err
		}

		result = append(result, tweet)
	}

	return result, nil
}

func tweetToTweet(t anaconda.Tweet) (Tweet, error) {
	return tweetToTweetCore(t, 0)
}

func tweetToTweetCore(t anaconda.Tweet, depth int) (Tweet, error) {
	ti, err := t.CreatedAtTime()
	if err != nil {
		return Tweet{}, err
	}

	var userName = t.User.Name

	var quotedTweet *Tweet
	if t.QuotedStatus != nil {
		// 無限ループ防止
		if depth > 100 {
			return Tweet{}, errors.New("recursive references")
		}

		q, err := tweetToTweetCore(*t.QuotedStatus, depth+1)
		if err == nil {
			quotedTweet = &q
		}
	}

	var urls []string
	for _, e := range t.Entities.Urls {
		urls = append(urls, e.Expanded_url)
	}

	var mediaURLs []string
	for _, m := range t.Entities.Media {
		mediaURLs = append(mediaURLs, m.Media_url_https)
	}

	var hashTags []string
	for _, t := range t.Entities.Hashtags {
		hashTags = append(hashTags, t.Text)
	}

	return Tweet{
		ID:          t.IdStr,
		UserName:    userName,
		Date:        jst.From(ti),
		Text:        t.FullText,
		QuotedTweet: quotedTweet,
		URLs:        urls,
		MediaURLs:   mediaURLs,
		HashTags:    hashTags,
	}, nil
}
