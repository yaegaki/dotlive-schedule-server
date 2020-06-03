package store

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

// actor 配信者
type actor struct {
	// id 配信者ID
	id string
	// Name 名前
	Name string `firestore:"name"`
	// Hashtag ハッシュタグ
	Hashtag string `firestore:"hashTag"`
	// Icon アイコンのURL
	Icon string `firestore:"icon"`
	// TwitterScreenName Twitterのスクリーンネーム
	TwitterScreenName string `firestore:"twitterScreenName"`
	// Emoji 推しアイコン
	Emoji string `firestore:"emoji"`
	// YoutubeChannelID YoutubeのチャンネルID
	YoutubeChannelID string `firestore:"youtubeChannelID"`
	// BilibiliID BilibiliのID
	BilibiliID string `firestore:"bilibiliID"`
	// BilibiliID BilibiliのID
	MildomID string `firestore:"mildomID"`
	// LastTweetID 最後に取得したTweetのID
	LastTweetID string `firestore:"lastTweetID"`
}

const collectionNameActor = "Actor"

// FindActors 配信者を検索する
func FindActors(ctx context.Context, c *firestore.Client) (model.ActorSlice, error) {
	it := c.Collection(collectionNameActor).Documents(ctx)
	docs, err := it.GetAll()
	if err != nil {
		return nil, err
	}
	var actors []model.Actor
	for _, doc := range docs {
		var a actor
		doc.DataTo(&a)
		a.id = doc.Ref.ID
		actors = append(actors, model.Actor{
			ID:                a.id,
			Name:              a.Name,
			Hashtag:           a.Hashtag,
			Icon:              a.Icon,
			TwitterScreenName: a.TwitterScreenName,
			Emoji:             a.Emoji,
			YoutubeChannelID:  a.YoutubeChannelID,
			BilibiliID:        a.BilibiliID,
			MildomID:          a.MildomID,
			LastTweetID:       a.LastTweetID,
		})
	}

	return actors, nil
}

// SaveActor 配信者を保存する
func SaveActor(ctx context.Context, c *firestore.Client, a model.Actor) error {
	// 常に上書きでいいのでトランザクションにしない
	_, err := c.Collection(collectionNameActor).Doc(a.ID).Set(ctx, actor{
		Name:              a.Name,
		Hashtag:           a.Hashtag,
		Icon:              a.Icon,
		TwitterScreenName: a.TwitterScreenName,
		Emoji:             a.Emoji,
		YoutubeChannelID:  a.YoutubeChannelID,
		BilibiliID:        a.BilibiliID,
		MildomID:          a.MildomID,
		LastTweetID:       a.LastTweetID,
	})
	return err
}
