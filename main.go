package main

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/yaegaki/dotlive-schedule-server/app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e := echo.New()
	e.HideBanner = true

	app.Route(e)

	e.Logger.Fatal(e.Start(":" + port))
}
