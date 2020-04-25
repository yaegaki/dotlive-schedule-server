package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo/v4"
	"github.com/yaegaki/dotlive-schedule-server/app/service"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/store"
)

func ScheduleHandler(c echo.Context) error {
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

	s, _ := service.CreateSchedule(ctx, client, now, actors)
	bytes, _ := json.Marshal(s)
	return c.JSONBlob(http.StatusOK, bytes)
}
