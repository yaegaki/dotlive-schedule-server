package handler

import (
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo/v4"
	"github.com/yaegaki/dotlive-schedule-server/app/service"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/store"
)

// RouteCalendar カレンダー関連のルーティングを設定する
func RouteCalendar(e *echo.Echo) {
	e.GET("/api/calendar", calendarHandler)
}

func calendarHandler(c echo.Context) error {
	ctx := c.Request().Context()

	query := c.Request().URL.Query()
	now := jst.Now()
	baseDate := jst.ShortDate(now.Year(), now.Month(), 1)
	q := query.Get("q")
	if q != "" {
		temp, err := parseYearMonthDayQuery(q)
		if err == nil {
			baseDate = temp
		}
	}

	actorOnly := query.Get("t") == "actor"

	client, err := firestore.NewClient(ctx, "dotlive-schedule")
	if err != nil {
		return c.String(http.StatusInternalServerError, "error1")
	}

	actors, err := store.FindActors(ctx, client)
	if err != nil {
		return c.String(http.StatusInternalServerError, "error2")
	}

	res := CalendarResponse{}
	for _, a := range actors {
		res.Actors = append(res.Actors, CalendarActor{
			ID:    a.ID,
			Name:  a.Name,
			Icon:  a.Icon,
			Emoji: a.Emoji,
		})
	}

	if actorOnly {
		res.Calendar = model.Calendar{
			BaseDate: baseDate,
			Days:     model.CalendarDaySlice{},
		}
		return c.JSON(http.StatusOK, res)
	}

	calendar, err := service.CreateCalendar(ctx, client, baseDate, now, actors)
	if err != nil {
		return c.String(http.StatusInternalServerError, "error3")
	}

	res.Calendar = calendar

	return c.JSON(http.StatusOK, res)
}

// CalendarResponse カレンダーAPIのレスポンス
type CalendarResponse struct {
	Calendar model.Calendar     `json:"calendar"`
	Actors   CalendarActorSlice `json:"actors"`
}

// CalendarActor Actorのサブセット
type CalendarActor struct {
	// ID ID
	ID string `json:"id"`
	// Name 名前
	Name string `json:"name"`
	// Icon アイコンのURL
	Icon string `json:"icon"`
	// Emoji 絵文字
	Emoji string `json:"emoji"`
}

// CalendarActorSlice CalendarActorのスライス
type CalendarActorSlice []CalendarActor
