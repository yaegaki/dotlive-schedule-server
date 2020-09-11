package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yaegaki/dotlive-schedule-server/store"
)

// RouteWidget ウィジェット用のAPI
func RouteWidget(e *echo.Echo) {
	e.GET("/api/awaisensei", awaiSenseiHandler)
}

func awaiSenseiHandler(c echo.Context) error {
	req := c.Request()
	ctx := req.Context()
	client := store.GetClient()

	s, err := store.FindAwaiSenseiSchedule(ctx, client)
	if err != nil {
		// どのような理由で失敗してもinternal server error
		return c.String(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, s)
}
