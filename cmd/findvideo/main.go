package main

import (
	"context"
	"log"

	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/youtube"
	"golang.org/x/oauth2/google"
	y "google.golang.org/api/youtube/v3"
)

func main() {
	ctx := context.Background()
	httpClient, err := google.DefaultClient(ctx, y.YoutubeReadonlyScope)
	if err != nil {
		panic(err)
	}

	youtubeService, err := y.New(httpClient)
	if err != nil {
		panic(err)
	}

	url := "https://www.youtube.com/watch?v=l8msnfPoPI8"
	v, err := youtube.FindVideo(ctx, youtubeService, url, model.Actor{}, jst.ShortDate(2020, 1, 1))
	if err != nil {
		panic(err)
	}
	log.Println(v)
}
