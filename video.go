package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ChimeraCoder/anaconda"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var youtubeURLPrefixes = []string{
	"https://youtu.be/",
	"https://www.youtube.com/watch",
}

var bilibiliURLPrefixes = []string{
	"https://live.bilibili.com/21307497",
}

const (
	VideoSourceYoutube  = "Youtube"
	VideoSourceBilibili = "Bilibili"
)

// ErrInvalidChannel 動画が対象配信者の物じゃない
var ErrInvalidChannel = errors.New("Invalid channel")

func getAndUpdateVideos(ctx context.Context, api *anaconda.TwitterApi, c *firestore.Client, actors []Actor) {
	httpClient, err := google.DefaultClient(ctx, youtube.YoutubeReadonlyScope)
	if err != nil {
		log.Printf("Can not create http client: %v", err)
		return
	}
	youtubeService, err := youtube.New(httpClient)
	if err != nil {
		log.Printf("Can not create youtube service: %v", err)
		return
	}

	videoCollection := c.Collection("Video")

	for _, actor := range actors {
		tl, err := getTimeline(api, actor.TwitterScreenName, actor.LastTweetID)
		if err != nil {
			log.Printf("Can not get tweet for %v: %v", actor.Name, err)
			continue
		}

		hasError := false

		lastTweetID := ""

	FOR_TWEET:
		for _, tweet := range tl {
			if lastTweetID == "" {
				lastTweetID = tweet.IdStr
			}

			for _, urlEntity := range tweet.Entities.Urls {
				var v Video
				if isYoutubeURL(urlEntity.Expanded_url) {
					v, err = getVideoFromYoutube(youtubeService, urlEntity.Expanded_url, actor)
					if err != nil {
						if err == ErrInvalidChannel {
							continue
						}

						log.Printf("Can not get youtube info for %v: %v : %v", actor.Name, urlEntity.Expanded_url, err)
						hasError = true
						break FOR_TWEET
					}
				} else if isBilibiliURL(urlEntity.Expanded_url) {
					v, err = getVideoFromBilibili(urlEntity.Expanded_url, actor, tweet)
					if err != nil {
						if err == ErrInvalidChannel {
							continue
						}

						log.Printf("Can not get bilibili info for %v: %v : %v", actor.Name, urlEntity.Expanded_url, err)
						hasError = true
						break FOR_TWEET
					}
				} else {
					continue
				}
				v.Text = tweet.FullText

				docRef := videoCollection.Doc(v.id)
				_, err := docRef.Get(ctx)
				if err != nil {
					if status.Code(err) != codes.NotFound {
						log.Printf("Can not get old video(%v): %v", v.id, err)
						hasError = true
						break FOR_TWEET
					}

					_, err = docRef.Set(ctx, v)
					if err != nil {
						log.Printf("Can not set video(%v): %v", v.id, err)
						hasError = true
						break FOR_TWEET
					}
				}
			}
		}

		// エラーが出ていた場合はlastTweetIDを更新しない
		if hasError {
			continue
		}

		if lastTweetID == "" {
			continue
		}

		actor.LastTweetID = lastTweetID
		err = actor.update(ctx, c)
		if err != nil {
			log.Printf("Can not update latestTweetID for %v: %v", actor.Name, err)
		}
	}
}

func isYoutubeURL(url string) bool {
	for _, p := range youtubeURLPrefixes {
		if strings.HasPrefix(url, p) {
			return true
		}
	}

	return false
}

func isBilibiliURL(url string) bool {
	for _, p := range bilibiliURLPrefixes {
		if strings.HasPrefix(url, p) {
			return true
		}
	}

	return false
}

func getVideoFromYoutube(s *youtube.Service, youtubeURL string, actor Actor) (Video, error) {
	u, err := url.Parse(youtubeURL)
	if err != nil {
		return Video{}, err
	}

	var videoID string
	qv, ok := u.Query()["v"]
	if ok && len(qv) > 0 {
		videoID = qv[0]
	} else {
		xs := strings.Split(strings.Trim(u.Path, "/"), "/")
		videoID = xs[len(xs)-1]
	}

	var item *youtube.Video
	retry := 0
	for {
		res, err := s.Videos.List("snippet,contentDetails,liveStreamingDetails").Id(videoID).Do()
		if err != nil {
			return Video{}, err
		}

		if len(res.Items) == 0 {
			retry++
			log.Printf("Can not get video info %v. retry after 5 sec. retry(%v)", videoID, retry)
			if retry >= 5 {
				return Video{}, fmt.Errorf("Can not get video info %v", videoID)
			}
			<-time.After(5 * time.Second)
			continue
		}

		item = res.Items[0]
		if item.Snippet.ChannelId != actor.YoutubeChannelID {
			return Video{}, ErrInvalidChannel
		}

		break
	}

	v := Video{
		id:      videoID + "-Youtube",
		ActorID: actor.id,
		Source:  VideoSourceYoutube,
		URL:     youtubeURL,
	}

	var startAtStr string
	if item.LiveStreamingDetails != nil {
		// プレミア公開の場合もLiveStreamingDetailsが存在する
		v.IsLive = true
		startAtStr = item.LiveStreamingDetails.ScheduledStartTime
	} else {
		startAtStr = item.Snippet.PublishedAt
	}

	startAt, err := time.Parse(time.RFC3339, startAtStr)
	if err != nil {
		return Video{}, err
	}
	v.StartAt = startAt.In(jst)

	return v, nil
}

func getVideoFromBilibili(bilibiliURL string, actor Actor, tweet anaconda.Tweet) (Video, error) {
	u, err := url.Parse(bilibiliURL)
	if err != nil {
		return Video{}, err
	}

	xs := strings.Split(strings.Trim(u.Path, "/"), "/")
	roomID := xs[len(xs)-1]
	res, err := http.Get(fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom?room_id=%v", roomID))
	if err != nil {
		return Video{}, err
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Video{}, err
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
		return Video{}, err
	}

	if strconv.Itoa(roomInfo.Data.RoomInfo.UID) != actor.BilibiliID {
		return Video{}, ErrInvalidChannel
	}

	t, err := tweet.CreatedAtTime()
	if err != nil {
		return Video{}, err
	}
	t = t.In(jst)

	return Video{
		// bilibiliは放送URL固定?なのでツイートの度にVideoエントリを作成する
		// TODO: 連続でツイートすると複数エントリができてしまうのである程度まとめる
		id:      tweet.IdStr + "-bilibili",
		ActorID: actor.id,
		Source:  VideoSourceBilibili,
		URL:     bilibiliURL,
		IsLive:  true,
		StartAt: t,
	}, nil
}

// findVideos 指定された日とその次の日に配信された動画を取得する
func findVideos(ctx context.Context, c *firestore.Client, date time.Time) ([]Video, error) {
	dateJST := date.In(jst)
	begin := createJSTTime(dateJST.Year(), dateJST.Month(), dateJST.Day(), 0, 0)
	dateJST = dateJST.Add(24 * time.Hour)
	end := createJSTTime(dateJST.Year(), dateJST.Month(), dateJST.Day(), 23, 59)
	q := c.Collection("Video").Where("startAt", ">=", begin).Where("startAt", "<=", end)

	it := q.Documents(ctx)
	docs, err := it.GetAll()
	if err != nil {
		return nil, err
	}

	videos := []Video{}
	for _, doc := range docs {
		var v Video
		doc.DataTo(&v)
		v.id = doc.Ref.ID
		v.StartAt = v.StartAt.In(jst)
		videos = append(videos, v)
	}

	return videos, nil
}

// findNotNotifiedVideos プッシュ通知されていない配信を取得する
func findNotNotifiedVideos(ctx context.Context, c *firestore.Client) ([]Video, error) {
	it := c.Collection("Video").Where("notified", "==", false).Documents(ctx)
	docs, err := it.GetAll()
	if err != nil {
		return nil, err
	}

	videos := []Video{}
	for _, doc := range docs {
		var v Video
		doc.DataTo(&v)
		v.id = doc.Ref.ID
		v.StartAt = v.StartAt.In(jst)
		videos = append(videos, v)
	}

	return videos, nil
}

// videosMarkAsNotified 引数で渡した動画のnotifiedをtrueにする
// 戻り値は実際にnotifiedをtrueにした動画
func videosMarkAsNotified(ctx context.Context, c *firestore.Client, videos []Video) ([]Video, error) {
	var updated []Video
	err := c.RunTransaction(ctx, func(ctx context.Context, t *firestore.Transaction) error {
		updated = []Video{}

		for _, v := range videos {
			docRef := c.Collection("Video").Doc(v.id)
			doc, err := t.Get(docRef)
			if err != nil {
				return err
			}

			doc.DataTo(&v)
			if v.Notified {
				continue
			}

			v.Notified = true
			updated = append(updated, v)
		}

		for _, v := range updated {
			_, err := c.Collection("Video").Doc(v.id).Set(ctx, v)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return updated, nil
}
