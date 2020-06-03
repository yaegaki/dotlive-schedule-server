package mildom

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/yaegaki/dotlive-schedule-server/common"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

var mildomURLPrefixes = []string{
	"https://www.mildom.com/",
	"https://mildom.com/",
}

// IsMildomURL URLがMildomのものかどうか
func IsMildomURL(url string) bool {
	for _, p := range mildomURLPrefixes {
		if strings.HasPrefix(url, p) {
			return true
		}
	}

	return false
}

// FindVideo MildomのURLから動画情報を取得する
func FindVideo(mildomURL string, actor model.Actor, tweetDate jst.Time) (model.Video, error) {
	u, err := url.Parse(mildomURL)
	if err != nil {
		return model.Video{}, err
	}

	xs := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(xs) != 1 {
		return model.Video{}, common.ErrInvalidChannel
	}

	mildomID := xs[len(xs)-1]
	if actor.MildomID != mildomID {
		return model.Video{}, common.ErrInvalidChannel
	}

	return model.Video{
		// Mildomは放送URL固定なので1日1回しか配信しない前提でツイート日をIDにする
		ID:      fmt.Sprintf("%v-%v-%v-mildom", tweetDate.Year(), int(tweetDate.Month()), tweetDate.Day()),
		ActorID: actor.ID,
		Source:  model.VideoSourceMildom,
		URL:     mildomURL,
		IsLive:  true,
		StartAt: tweetDate,
	}, nil
}
