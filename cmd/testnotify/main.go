package main

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/notify"
	"github.com/yaegaki/dotlive-schedule-server/store"
)

func main() {
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("Can not create firebase app: %v", err)
	}
	msgCli, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("Can not create firebase messaging client: %v", err)
	}
	log.Print(msgCli)

	storeCli, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Can not create firestore client: %v", err)
	}

	plan, err := store.FindLatestPlan(ctx, storeCli)
	if err != nil {
		log.Fatalf("Can not get latest plan: %v", err)
	}

	actors, err := store.FindActors(ctx, storeCli)
	if err != nil {
		log.Fatalf("Can not get actors: %v", err)
	}

	videos, err := store.FindVideos(ctx, storeCli, jst.Range{
		Begin: jst.ShortDate(2020, 5, 7),
		End:   jst.ShortDate(2020, 5, 8),
	})
	if err != nil {
		log.Fatalf("Can not get videos: %v", err)
	}

	// 計画の通知
	notify.PushNotifyPlan(ctx, msgCli, plan, actors)

	// 動画の通知
	actor, err := actors.FindActor(videos[0].ActorID)
	if err != nil {
		log.Fatalf("Can not get actor")
	}
	log.Print(actor)

	notify.PushNotifyVideo(ctx, msgCli, videos[0].StartAt, videos[0], []model.Actor{actor})
}
