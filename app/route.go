package app

import (
	"github.com/labstack/echo/v4"
	"github.com/yaegaki/dotlive-schedule-server/app/handler"
)

// Route httpのハンドラを設定する
func Route(e *echo.Echo) {
	handler.RouteJob(e)
	handler.RouteSchedule(e)
	handler.RouteTopic(e)
	handler.RouteCalendar(e)
}
