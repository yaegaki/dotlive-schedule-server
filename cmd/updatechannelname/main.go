package main

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"github.com/yaegaki/dotlive-schedule-server/store"
	"golang.org/x/oauth2/google"
	y "google.golang.org/api/youtube/v3"
)

func main() {
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("Can not create firebase app: %v", err)
	}

	storeCli, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Can not create firestore client: %v", err)
	}

	actors, err := store.FindActors(ctx, storeCli)
	if err != nil {
		log.Fatalf("Can not get actors: %v", err)
	}

	httpClient, err := google.DefaultClient(ctx, y.YoutubeReadonlyScope)
	if err != nil {
		panic(err)
	}

	youtubeService, err := y.New(httpClient)
	if err != nil {
		panic(err)
	}

	for _, actor := range actors {
		res, err := youtubeService.Channels.List("snippet").Id(actor.YoutubeChannelID).Do()
		if err != nil {
			log.Printf("Can not get channel info: %v %v", actor.Name, err)
			continue
		}

		if len(res.Items) == 0 {
			log.Printf("res.Items is empty: %v", actor.Name)
			continue
		}

		actor.YoutubeChannelName = res.Items[0].Snippet.Title
		err = store.SaveActor(ctx, storeCli, actor)
		if err != nil {
			log.Fatalf("Can not save actor: %v %v", actor.Name, err)
		}
	}
}
