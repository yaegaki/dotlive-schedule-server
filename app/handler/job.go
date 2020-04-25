package handler

import (
	"context"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/ChimeraCoder/anaconda"
	"github.com/labstack/echo/v4"
	"github.com/yaegaki/dotlive-schedule-server/app/internal"
	"github.com/yaegaki/dotlive-schedule-server/app/service"
	"github.com/yaegaki/dotlive-schedule-server/common"
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

	client, err := firestore.NewClient(ctx, "dotlive-schedule")
	if err != nil {
		log.Printf("Can not create a firestore client: %v", err)
		return c.String(http.StatusInternalServerError, "error1")
	}
	defer client.Close()

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

	// 最新の計画を取得する
	plan, err := store.FindLatestPlan(ctx, client)

	if err != nil && err != common.ErrNotFound {
		log.Printf("Can not get latestplan: %v", err)
		return c.String(http.StatusInternalServerError, "error4")
	}

	lastTweetID := plan.SourceID

	api := anaconda.NewTwitterApi("", "")

	// プロフィール画像更新
	for _, a := range actors {
		updateProfileImage(ctx, api, client, &a)
	}

	// ツイートから計画を取得する
	newPlans, err := tweet.FindPlans(api, lastTweetID, actors)
	if err != nil {
		log.Printf("Can not get plans: %v", err)
	}

	for _, p := range newPlans {
		err := store.SavePlan(ctx, client, p)
		if err != nil {
			log.Printf("Can not save plan %v: %v", p.Date, err)
		}
	}

	// ツイートから動画情報を取得する
	tweet.ResolveVideos(api, actors, videoResolver)

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

		actor, err := actors.FindActor(v.ActorID)
		if err != nil {
			log.Printf("Can not get actor %v", v.ActorID)
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

		err = store.SaveVideo(ctx, c, v)
		if err != nil {
			log.Printf("Can not save video %v: %v", v.ID, err)
		}
	}
}
