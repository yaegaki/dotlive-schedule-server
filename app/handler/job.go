package handler

import (
	"context"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/ChimeraCoder/anaconda"
	"github.com/labstack/echo/v4"
	"github.com/yaegaki/dotlive-schedule-server/app/cache"
	"github.com/yaegaki/dotlive-schedule-server/app/internal"
	"github.com/yaegaki/dotlive-schedule-server/app/service"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/store"
	"github.com/yaegaki/dotlive-schedule-server/tweet"
	"github.com/yaegaki/dotlive-schedule-server/youtube"
)

// appEngineCronHeader
const appEngineCronHeader = "X-Appengine-Cron"

// RouteJob ジョブ関連のルーティングを設定する
func RouteJob(e *echo.Echo) {
	e.GET("/_task/job", jobHandler)
}

// jobHandler 定期実行ジョブ
func jobHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if !internal.IsDevelop && c.Request().Header.Get(appEngineCronHeader) != "true" {
		return c.String(http.StatusBadRequest, "bad request")
	}

	client := store.GetClient()

	videoResolver, err := service.NewVideoResolver(ctx, client)
	if err != nil {
		log.Printf("Can not create VideoResolver: %v", err)
		return c.String(http.StatusInternalServerError, "error2")
	}

	actors, err := store.FindActors(ctx, client)
	if err != nil {
		log.Printf("Can not get actors: %v", err)
		return c.String(http.StatusInternalServerError, "error3")
	}

	api := anaconda.NewTwitterApi("", "")

	// あわい先生のどっとライブスケジュールの情報を更新
	updateAwaiSenseiSchedule(ctx, api, client)

	// プロフィール画像更新
	for _, a := range actors {
		updateProfileImage(ctx, api, client, &a)
	}

	userDotlive, err := store.FindTwitterUser(ctx, client, tweet.ScreenNameDotlive)
	if err != nil {
		log.Printf("Can not get dotlive twitteruser: %v", err)
		return c.String(http.StatusInternalServerError, "error4")
	}

	// ツイートから計画を取得する
	lastTweetID := userDotlive.LastTweetID
	userDotlive, newPlans, err := tweet.FindPlans(api, userDotlive, actors)
	if err != nil {
		log.Printf("Can not get plans: %v", err)
	} else {
		now := jst.Now()
		// 古い計画から順番にセーブしていく
		for i := len(newPlans) - 1; i >= 0; i-- {
			p := newPlans[i]

			// 2日以上前の過去の計画の更新はおかしいので無視する
			if now.AddDay(-2).After(p.Date) {
				log.Printf("Invalid plan date %v", p.Date)
				continue
			}

			err := store.SavePlan(ctx, client, p)
			if err != nil {
				if err == store.ErrFixedPlan {
					log.Printf("Plan is Fixed: %v", p.Date)
				} else {
					log.Printf("Can not save plan %v: %v", p.Date, err)
					return c.String(http.StatusInternalServerError, "error5")
				}
			}
		}

		// TwitterUserの更新
		// 必ず計画を保存した後に更新する
		if lastTweetID != userDotlive.LastTweetID {
			err = store.SaveTwitterUser(ctx, client, userDotlive)
			if err != nil {
				log.Printf("Can not save dotlive twitteruser: %v", err)
				return c.String(http.StatusInternalServerError, "error6")
			}
		}
	}

	// ツイートから動画情報を取得する
	tweet.ResolveVideos(api, actors, videoResolver)

	// 配信者情報をキャッシュ
	cache.SetActors(actors)

	// 開始時間の更新
	updateVideoStartAt(ctx, client, videoResolver, actors)

	// プッシュ通知
	service.PushNotify(ctx, client, actors)

	return c.String(http.StatusOK, "done.")
}

func updateProfileImage(ctx context.Context, api *anaconda.TwitterApi, c *firestore.Client, actor *model.Actor) {
	url, err := tweet.GetProfileImageURL(api, *actor)
	if err != nil {
		log.Printf("Can not get profile image for %v: %v", actor.Name, err)
		return
	}

	url = strings.Replace(url, "_normal", "", 1)

	if actor.Icon == url {
		return
	}

	copy := *actor
	copy.Icon = url
	err = store.SaveActor(ctx, c, copy)
	if err != nil {
		log.Printf("Can not save actor %v: %v", actor.Name, err)
		return
	}

	*actor = copy
}

func updateAwaiSenseiSchedule(ctx context.Context, api *anaconda.TwitterApi, client *firestore.Client) {
	userAwaiSensei, err := store.FindTwitterUser(ctx, client, tweet.ScreenNameAwaiSensei)
	if err != nil {
		log.Printf("Can not get user(awaisensei): %v", err)
		return
	}

	tweets, err := tweet.GetTimeline(api, userAwaiSensei.ScreenName, userAwaiSensei.LastTweetID)
	if err != nil {
		log.Printf("Can not get timeline: %v", err)
		return
	}

	if len(tweets) == 0 {
		return
	}

	userAwaiSensei.LastTweetID = tweets[0].ID

	var schedule model.AwaiSenseiSchedule
	for _, t := range tweets {
		if !isAwaiSenseiScheduleTweet(t) {
			continue
		}

		schedule = model.AwaiSenseiSchedule{
			TweetID:  t.ID,
			Title:    strings.Split(t.Text, "\n")[0],
			ImageURL: t.MediaURLs[0],
		}

		break
	}

	if schedule.TweetID != "" {
		err = store.SaveAwaiSenseiSchedule(ctx, client, schedule)
		if err != nil {
			log.Printf("Can not save awaisenseischedule: %v", err)
			return
		}
	}

	err = store.SaveTwitterUser(ctx, client, userAwaiSensei)
	if err != nil {
		log.Printf("Can not save user: %v", err)
	}
}

func isAwaiSenseiScheduleTweet(t tweet.Tweet) bool {
	if len(t.MediaURLs) == 0 {
		return false
	}

	// ハッシュタグがある場合は確実
	for _, ht := range t.HashTags {
		if ht == "どっとライブ予定表" {
			return true
		}
	}

	// ハッシュタグがない場合は以下の条件で判定する
	// 1. 1行目に"どっとライブ予定表"が含まれる
	// 2. 2行目が空白

	lines := strings.Split(t.Text, "\n")
	if len(lines) < 3 {
		return false
	}

	if strings.Index(lines[0], "どっとライブ予定表") < 0 {
		return false
	}

	return strings.TrimSpace(lines[1]) == ""
}

// updateVideoStartAt 開始予定時間より早く始まっている場合に開始時間を修正する
func updateVideoStartAt(ctx context.Context, c *firestore.Client, vr *service.VideoResolver, actors model.ActorSlice) {
	videos, err := store.FindNotNotifiedVideos(ctx, c)
	if err != nil {
		log.Printf("Can not get videos: %v", err)
		return
	}

	now := jst.Now()

	for _, v := range videos {
		if v.StartAt.Before(now) {
			continue
		}

		var relatedActorID string
		if v.IsUnknownActor() {
			relatedActorID = v.RelatedActorID
		} else {
			relatedActorID = v.ActorID
		}
		actor, err := actors.FindActor(relatedActorID)
		if err != nil {
			log.Printf("Can not get actor %v", relatedActorID)
			continue
		}

		newVideo, err := youtube.FindVideo(ctx, vr.YoutubeService(), v.URL, actor)
		if err != nil {
			log.Printf("Can not get video info %v: %v", v.ID, err)
			continue
		}

		if v.StartAt.Equal(newVideo.StartAt) {
			continue
		}
		v.StartAt = newVideo.StartAt

		err = store.SaveVideo(ctx, c, v, nil)
		if err != nil {
			log.Printf("Can not save video %v: %v", v.ID, err)
		}
	}
}
