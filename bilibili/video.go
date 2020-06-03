package bilibili

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/yaegaki/dotlive-schedule-server/common"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

var bilibiliURLPrefixes = []string{
	"https://live.bilibili.com/",
}

// IsBilibiliURL URLがBilibiliのものかどうか
func IsBilibiliURL(url string) bool {
	for _, p := range bilibiliURLPrefixes {
		if strings.HasPrefix(url, p) {
			return true
		}
	}

	return false
}

// FindVideo BilibiliのURLから動画情報を取得する
func FindVideo(bilibiliURL string, actor model.Actor, tweetDate jst.Time) (model.Video, error) {
	u, err := url.Parse(bilibiliURL)
	if err != nil {
		return model.Video{}, err
	}

	xs := strings.Split(strings.Trim(u.Path, "/"), "/")
	roomID := xs[len(xs)-1]
	res, err := http.Get(fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom?room_id=%v", roomID))
	if err != nil {
		return model.Video{}, err
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return model.Video{}, err
	}

	var roomInfo struct {
		Data struct {
			RoomInfo struct {
				UID int `json:"uid"`
			} `json:"room_info"`
		} `json:"data"`
	}
	err = json.Unmarshal(bytes, &roomInfo)
	if err != nil {
		return model.Video{}, err
	}

	if strconv.Itoa(roomInfo.Data.RoomInfo.UID) != actor.BilibiliID {
		return model.Video{}, common.ErrInvalidChannel
	}

	return model.Video{
		// bilibiliは放送URL固定なので1日1回しか配信しない前提でツイート日をIDにする
		ID:      fmt.Sprintf("%v-%v-%v-biibili", tweetDate.Year(), int(tweetDate.Month()), tweetDate.Day()),
		ActorID: actor.ID,
		Source:  model.VideoSourceBilibili,
		URL:     bilibiliURL,
		IsLive:  true,
		StartAt: tweetDate,
	}, nil
}
