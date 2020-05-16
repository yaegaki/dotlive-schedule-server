package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yaegaki/dotlive-schedule-server/app/cache"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/notify"
	"github.com/yaegaki/dotlive-schedule-server/store"
)

// RouteTopic プッシュ通知のトピック関連のルーティングを設定する
func RouteTopic(e *echo.Echo) {
	e.POST("/api/topic", topicHandler)
	/*
		e.GET("/debug/topic", func(c echo.Context) error {
			return c.HTML(http.StatusOK, `<html><body>
			<form action="/api/topic" method="post">
				token:<input type="text" name="t">
				<input type="submit">
			</form>
			</body></html>`)
		})
	*/
}

func topicHandler(c echo.Context) error {
	req := c.Request()
	ctx := req.Context()
	client := store.GetClient()

	actors, err := cache.FindActorsWithCache(ctx, client)
	if err != nil {
		return c.String(http.StatusInternalServerError, "error2")
	}

	req.ParseForm()
	token := req.Form.Get("t")
	if token == "" {
		return c.String(http.StatusBadRequest, "bad request")
	}

	subscribedTopics, err := notify.GetTopics(token)
	if err != nil {
		// どのような理由で失敗してもbad request
		return c.String(http.StatusBadRequest, "bad request...")
	}

	result := []model.Topic{
		model.Topic{
			Name:        "plan",
			DisplayName: "計画",
		},
	}

	for _, a := range actors {
		result = append(result, model.Topic{
			// Twitterのスクリーンネームをトピック名に使用する
			Name:        a.TwitterScreenName,
			DisplayName: a.Name,
		})
	}

	for _, t := range subscribedTopics {
		for i := range result {
			temp := result[i]
			if temp.Name == t {
				temp.Subscribed = true
				result[i] = temp
				break
			}
		}
	}

	return c.JSON(http.StatusOK, result)
}
