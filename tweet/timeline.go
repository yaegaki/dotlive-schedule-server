package tweet

import (
	"net/url"

	"github.com/ChimeraCoder/anaconda"
	"github.com/yaegaki/dotlive-schedule-server/jst"
)

// GetTimeline タイムラインを取得する
func GetTimeline(api *anaconda.TwitterApi, screenName, lastTweetID string) ([]Tweet, error) {
	return getTimeline(api, screenName, lastTweetID)
}

func getTimeline(api *anaconda.TwitterApi, screenName, lastTweetID string) ([]Tweet, error) {
	param := url.Values{
		"screen_name":     []string{screenName},
		"exclude_replies": []string{"false"},
		"include_rts":     []string{"false"},
	}
	if lastTweetID != "" {
		param["since_id"] = []string{lastTweetID}
	}

	timeline, err := api.GetUserTimeline(param)
	if err != nil {
		return nil, err
	}

	var result []Tweet
	for _, t := range timeline {
		ti, err := t.CreatedAtTime()
		if err != nil {
			return nil, err
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

		result = append(result, Tweet{
			ID:        t.IdStr,
			Date:      jst.From(ti),
			Text:      t.FullText,
			URLs:      urls,
			MediaURLs: mediaURLs,
			HashTags:  hashTags,
		})
	}

	return result, nil
}
