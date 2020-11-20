package main

import (
	"context"
	"log"
	"os"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/notify"
	"github.com/yaegaki/dotlive-schedule-server/store"
)

type testClient struct {
	c notify.Client
}

func (c testClient) Send(ctx context.Context, message *messaging.Message) (string, error) {
	message.Topic = "test"
	message.Condition = ""
	return c.c.Send(ctx, message)
}

func createClient(ctx context.Context, app *firebase.App) (notify.Client, error) {
	msgCli, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	tokens := strings.Split(os.Getenv("DEVICE_TOKEN"), ",")

	if len(tokens) > 0 {
		_, err = msgCli.SubscribeToTopic(ctx, tokens, "test")
		if err != nil {
			return nil, err
		}
	}

	return testClient{
		c: msgCli,
	}, nil
}

func main() {
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("Can not create firebase app: %v", err)
	}
	msgCli, err := createClient(ctx, app)
	if err != nil {
		log.Fatalf("Can not create firebase messaging client: %v", err)
	}

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
