package tweet

import (
	"net/url"

	"github.com/ChimeraCoder/anaconda"
	"github.com/yaegaki/dotlive-schedule-server/jst"
)

func getTimeline(api *anaconda.TwitterApi, screenName, lastTweetID string) ([]Tweet, error) {
	param := url.Values{
		"screen_name":     []string{screenName},
		"exclude_replies": []string{"true"},
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

		result = append(result, Tweet{
			ID:   t.IdStr,
			Date: jst.From(ti),
			Text: t.FullText,
			URLs: urls,
		})
	}

	return result, nil
}
