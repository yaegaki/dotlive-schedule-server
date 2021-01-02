package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/store"
)

func main() {
	// 配信者をjsonファイルから登録する

	// args := os.Args
	// if len(args) != 2 {
	// 	log.Fatal("usage: actorregister path/to/json")
	// }

	// a := args[1]
	a := "actors/202101.json"
	bytes, err := ioutil.ReadFile(a)
	if err != nil {
		log.Fatalf("Can not open %v", a)
	}

	var newActors []actor
	err = json.Unmarshal(bytes, &newActors)
	if err != nil {
		log.Fatalf("Can not parse json")
	}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, firestore.DetectProjectID)
	if err != nil {
		log.Fatalf("Can not create a firestore client: %v", err)
	}
	defer client.Close()

	actors, err := store.FindActors(ctx, client)
	if err != nil {
		log.Fatalf("Can not get actors: %v", err)
	}

	for _, a := range newActors {
		_, err := actors.FindActorByName(a.Name)
		if err == nil {
			log.Printf("already exists actor: %v", a.Name)
			continue
		}

		actor := model.Actor{
			Name:              a.Name,
			Hashtag:           a.Hashtag,
			TwitterScreenName: a.TwitterScreenName,
			Emoji:             a.Emoji,
			YoutubeChannelID:  a.YoutubeChannelID,
		}
		err = store.CreateActor(ctx, client, actor)
		if err != nil {
			log.Printf("Can not create actor '%v': %v", actor.Name, err)
			continue
		}

		log.Printf("Create actor: %v", actor.Name)
	}
}

type actor struct {
	Name              string `json:"name"`
	Hashtag           string `json:"hashtag"`
	TwitterScreenName string `json:"twitter_screenname"`
	Emoji             string `json:"emoji"`
	YoutubeChannelID  string `json:"youtube_channelID"`
}
