package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yaegaki/dotlive-schedule-server/app/cache"
	"github.com/yaegaki/dotlive-schedule-server/app/service"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/store"
)

// RouteSchedule スケジュール関連のルーティングを設定する
func RouteSchedule(e *echo.Echo) {
	e.GET("/api/schedule", scheduleHandler)
}

func scheduleHandler(c echo.Context) error {
	ctx := c.Request().Context()
	client := store.GetClient()

	now := jst.Now()
	q := c.Request().URL.Query().Get("q")
	if q != "" {
		temp, err := parseYearMonthDayQuery(q)
		if err == nil {
			now = temp
		}
	}

	actors, err := cache.FindActorsWithCache(ctx, client)
	if err != nil {
		return c.String(http.StatusInternalServerError, "error2")
	}

	s, err := service.CreateSchedule(ctx, client, now, actors)
	if err != nil {
		log.Printf("can not create schedule: %v", err)
		return c.String(http.StatusInternalServerError, "error3")
	}
	bytes, _ := json.Marshal(s)
	return c.JSONBlob(http.StatusOK, bytes)
}
