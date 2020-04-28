package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yaegaki/dotlive-schedule-server/notify"
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
	req.ParseForm()
	token := req.Form.Get("t")
	if token == "" {
		return c.String(http.StatusBadRequest, "bad request")
	}

	topics, err := notify.GetTopics(token)
	if err != nil {
		// どのような理由で失敗してもbad request
		return c.String(http.StatusBadRequest, "bad request...")
	}

	return c.JSON(http.StatusOK, topics)
}
