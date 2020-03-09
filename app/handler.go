package app

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ChimeraCoder/anaconda"
	"github.com/labstack/echo/v4"
	"github.com/yaegaki/dotlive-schedule-server/common"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/store"
	"github.com/yaegaki/dotlive-schedule-server/tweet"
)

// Route httpのハンドラを設定する
func Route(e *echo.Echo) {
	e.GET("/_task/job", jobHandler)
	e.GET("/", func(c echo.Context) error {
		ctx := c.Request().Context()
		client, err := firestore.NewClient(ctx, "dotlive-schedule")
		if err != nil {
			return c.String(http.StatusInternalServerError, "error1")
		}

		now := jst.Now()
		q := c.Request().URL.Query().Get("q")
		if q != "" {
			xs := strings.Split(q, "-")
			if len(xs) == 3 {
				year, err1 := strconv.Atoi(xs[0])
				month, err2 := strconv.Atoi(xs[1])
				day, err3 := strconv.Atoi(xs[2])
				if err1 == nil && err2 == nil && err3 == nil {
					now = jst.ShortDate(year, time.Month(month), day)
				}
			}
		}

		actors, err := store.FindActors(ctx, client)
		if err != nil {
			return c.String(http.StatusInternalServerError, "error2")
		}

		s, _ := createSchedule(ctx, client, now, actors)
		bytes, _ := json.Marshal(s)
		return c.JSONBlob(http.StatusOK, bytes)
	})
}

// appEngineCronHeader
const appEngineCronHeader = "X-Appengine-Cron"

func jobHandler(c echo.Context) error {
	ctx := c.Request().Context()
	isDevelop := os.Getenv("DEVELOP") == "true"

	if !isDevelop && c.Request().Header.Get(appEngineCronHeader) != "true" {
		return c.String(http.StatusBadRequest, "bad request")
	}

	client, err := firestore.NewClient(ctx, "dotlive-schedule")
	if err != nil {
		log.Printf("Can not create a firestore client: %v", err)
		return c.String(http.StatusInternalServerError, "error1")
	}
	defer client.Close()

	videoResolver, err := newVideoResolver(ctx, client)
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

	// プッシュ通知
	pushNotify(ctx, client, actors)

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
