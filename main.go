package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ChimeraCoder/anaconda"
	"github.com/labstack/echo/v4"
)

func jobHandler(c echo.Context) error {
	ctx := c.Request().Context()
	isDevelop := os.Getenv("DEVELOP") == "true"

	if !isDevelop && c.Request().Header.Get("X-Appengine-Cron") != "true" {
		return c.String(http.StatusBadRequest, "bad request")
	}

	api := anaconda.NewTwitterApi("", "")
	defer api.Close()

	client, err := firestore.NewClient(ctx, "dotlive-schedule")
	if err != nil {
		log.Printf("Can not create firestore client: %v", err)
		return c.String(http.StatusInternalServerError, "error1")
	}
	defer client.Close()

	actors, err := findActors(ctx, client)
	if err != nil {
		log.Printf("Can not get actors: %v", err)
		return c.String(http.StatusInternalServerError, "error2")
	}

	// どっとライブのツイートからスケジュール取得
	err = storePlanToStore(ctx, api, client, actors)
	if err != nil {
		log.Printf("Can not create plan: %v", err)
		// ここでエラーが出ても他のタスクは実行する
	}

	// ツイートから動画情報取得
	getAndUpdateVideos(ctx, api, client, actors)

	// プッシュ通知
	pushNotify(ctx, client)

	return c.String(http.StatusOK, "done.")
}

func main() {
	anaconda.SetConsumerKey(os.Getenv("TWITTER_CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("TWITTER_CONSUMER_SECRET"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e := echo.New()
	e.GET("/_task/job", jobHandler)
	e.GET("/", func(c echo.Context) error {
		ctx := c.Request().Context()
		client, err := firestore.NewClient(ctx, "dotlive-schedule")
		if err != nil {
			return c.String(http.StatusInternalServerError, "error1")
		}

		nowJST := time.Now().In(jst)
		q := c.Request().URL.Query().Get("q")
		if q != "" {
			d, err := strconv.Atoi(q)
			if err == nil {
				nowJST = nowJST.Add(time.Duration(d) * 24 * time.Hour)
			}
		}

		s, _ := findSchedule(ctx, client, nowJST)
		bytes, _ := json.Marshal(s)
		return c.JSONBlob(http.StatusOK, bytes)
	})

	e.Logger.Fatal(e.Start(":" + port))
}
