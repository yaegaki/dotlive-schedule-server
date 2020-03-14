package youtube

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/yaegaki/dotlive-schedule-server/common"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	y "google.golang.org/api/youtube/v3"
)

var youtubeURLPrefixes = []string{
	"https://youtu.be/",
	"https://www.youtube.com/watch",
}

// IsYoutubeURL YoutubeのURLかどうか
func IsYoutubeURL(url string) bool {
	for _, p := range youtubeURLPrefixes {
		if strings.HasPrefix(url, p) {
			return true
		}
	}

	return false
}

// FindVideo youtubeのURLから動画情報を取得する
func FindVideo(ctx context.Context, s *y.Service, youtubeURL string, actor model.Actor) (model.Video, error) {
	u, err := url.Parse(youtubeURL)
	if err != nil {
		return model.Video{}, err
	}

	var videoID string
	qv, ok := u.Query()["v"]
	if ok && len(qv) > 0 {
		videoID = qv[0]
	} else {
		xs := strings.Split(strings.Trim(u.Path, "/"), "/")
		videoID = xs[len(xs)-1]
	}

	var item *y.Video
	retry := 0
	for {
		res, err := s.Videos.List("snippet,contentDetails,liveStreamingDetails").Id(videoID).Do()
		if err != nil {
			return model.Video{}, err
		}

		if len(res.Items) == 0 {
			retry++
			log.Printf("Can not get video info %v. retry after 5 sec. retry(%v)", videoID, retry)
			if retry >= 5 {
				return model.Video{}, fmt.Errorf("Can not get video info %v", videoID)
			}
			<-time.After(5 * time.Second)
			continue
		}

		item = res.Items[0]
		if item.Snippet.ChannelId != actor.YoutubeChannelID {
			return model.Video{}, common.ErrInvalidChannel
		}

		break
	}

	v := model.Video{
		ID:      videoID + "-Youtube",
		ActorID: actor.ID,
		Source:  model.VideoSourceYoutube,
		URL:     youtubeURL,
	}

	var startAt time.Time
	if item.LiveStreamingDetails != nil {
		// プレミア公開の場合もLiveStreamingDetailsが存在する
		v.IsLive = true
		startAt, err = time.Parse(time.RFC3339, item.LiveStreamingDetails.ScheduledStartTime)
		if err != nil {
			return model.Video{}, err
		}

		// 既に始まってる場合。
		if item.LiveStreamingDetails.ActualStartTime != "" {
			actualStartAt, err := time.Parse(time.RFC3339, item.LiveStreamingDetails.ActualStartTime)
			if err != nil {
				return model.Video{}, err
			}

			if actualStartAt.Before(startAt) {
				startAt = actualStartAt
			}
		}
	} else {
		startAt, err = time.Parse(time.RFC3339, item.LiveStreamingDetails.ScheduledStartTime)
		if err != nil {
			return model.Video{}, err
		}
	}

	v.StartAt = jst.From(startAt)

	return v, nil
}
