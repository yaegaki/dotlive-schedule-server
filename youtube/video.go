package youtube

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/yaegaki/dotlive-schedule-server/common"
	"github.com/yaegaki/dotlive-schedule-server/internal/video"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	y "google.golang.org/api/youtube/v3"
)

var youtubeURLPrefixes = []string{
	"https://youtu.be/",
	"https://www.youtube.com/watch",
}

// IsYoutubeChannelURL YoutubeのチャンネルのURLかどうか
func IsYoutubeChannelURL(u string) bool {
	return strings.HasPrefix(u, "https://www.youtube.com/channel/")
}

// IsYoutubeURL YoutubeのURLかどうか
func IsYoutubeURL(url string) bool {
	return video.IsTargetVideoSource(youtubeURLPrefixes, url)
}

// FindVideo youtubeのURLから動画情報を取得する
func FindVideo(ctx context.Context, s *y.Service, youtubeURL string, relatedActor model.Actor, tweetDate jst.Time) (model.Video, error) {
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

	// コラボで他の配信者の枠の場合
	isCollaboVideo := false

	var item *y.Video
	retry := 0
	videoOwnerName := relatedActor.Name
	isDotLiveChannel := false
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
		if item.Snippet.ChannelId != relatedActor.YoutubeChannelID {
			if item.Snippet.ChannelId == ChannelIDDotLive {
				isDotLiveChannel = true
				videoOwnerName = ChannelNameDotLive
			} else if hasYoutubeChannelLink(item.Snippet.Description, relatedActor.YoutubeChannelID) {
				isCollaboVideo = true
				videoOwnerName = item.Snippet.ChannelTitle
			} else {
				return model.Video{}, common.ErrInvalidChannel
			}
		}

		break
	}

	v := model.Video{
		ID:        videoID + "-Youtube",
		Source:    model.VideoSourceYoutube,
		URL:       youtubeURL,
		OwnerName: videoOwnerName,
		// TODO: 動画からメン限かどうか取得する
		//       (無理そう, status.privacyStatusがunlistedだったら大体メン限だが限定公開の可能性もある)
		MemberOnly: false,
	}

	if isDotLiveChannel {
		v.ActorID = model.ActorIDUnknown
	} else if isCollaboVideo {
		v.ActorID = model.ActorIDUnknown
		v.RelatedActorID = relatedActor.ID
	} else {
		v.ActorID = relatedActor.ID
	}

	var startAt time.Time
	if item.LiveStreamingDetails != nil {
		// プレミア公開の場合もLiveStreamingDetailsが存在する
		v.IsLive = true
		hasScheduledStartTime := false
		// ScheduledStartTimeが存在しない場合があるのでチェックする
		if item.LiveStreamingDetails.ScheduledStartTime != "" {
			startAt, err = time.Parse(time.RFC3339, item.LiveStreamingDetails.ScheduledStartTime)
			if err != nil {
				return model.Video{}, err
			}
			hasScheduledStartTime = true
		}

		// 既に始まってる場合。
		hasActualStartTime := false
		if item.LiveStreamingDetails.ActualStartTime != "" {
			actualStartAt, err := time.Parse(time.RFC3339, item.LiveStreamingDetails.ActualStartTime)
			if err != nil {
				return model.Video{}, err
			}

			if !hasScheduledStartTime || actualStartAt.Before(startAt) {
				startAt = actualStartAt
			}
			hasActualStartTime = true
		}

		// どっちも見つからない場合
		// 存在するかわからないが一応対策しておく
		if !hasScheduledStartTime && !hasActualStartTime {
			startAt = tweetDate.Time()
			log.Printf("warning: not found startTime for id: %v", videoID)
		}
	} else {
		startAt, err = time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			return model.Video{}, err
		}
	}

	v.StartAt = jst.From(startAt)

	return v, nil
}

// hasYoutubeChannelLink 文字列中にyoutubeのチャンネルIDへのリンクが含まれているかどうか
func hasYoutubeChannelLink(text string, channelID string) bool {
	link := "https://www.youtube.com/channel/" + channelID
	return strings.Index(text, link) >= 0
}
