package store

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

// awaiSenseiSchedule あわい先生のどっとライブスケジュールの情報
type awaiSenseiSchedule struct {
	// TweetID 元ツイートのID
	TweetID string `json:"tweetID"`
	// Title タイトル
	Title string `json:"title"`
	// ImageURL 画像のURL
	ImageURL string `json:"imageURL"`
}

const collectionNameAwaiSenseiSchedule = "AwaiSenseiSchedule"
const docIDAwaiSenseiSchedule = "awaisensei-schedule"

// FindAwaiSenseiSchedule AwaiSenseiScheduleを検索する
func FindAwaiSenseiSchedule(ctx context.Context, c *firestore.Client) (model.AwaiSenseiSchedule, error) {
	docRef := c.Collection(collectionNameAwaiSenseiSchedule).Doc(docIDAwaiSenseiSchedule)
	doc, err := docRef.Get(ctx)
	if err != nil {
		// errがNotFoundでもエラーで返す(最初にデータを入れておくのでNotFoundは基本的にはありえない)
		return model.AwaiSenseiSchedule{}, err
	}

	var s awaiSenseiSchedule
	doc.DataTo(&s)

	return model.AwaiSenseiSchedule{
		TweetID:  s.TweetID,
		Title:    s.Title,
		ImageURL: s.ImageURL,
	}, nil
}

// SaveAwaiSenseiSchedule AwaiSenseiScheduleを保存する
func SaveAwaiSenseiSchedule(ctx context.Context, c *firestore.Client, s model.AwaiSenseiSchedule) error {
	// 常に上書きでいいのでトランザクションにしない
	// (理論上はうまくいかない可能性があるけどほぼ起こりえないので無視)
	_, err := c.Collection(collectionNameAwaiSenseiSchedule).Doc(docIDAwaiSenseiSchedule).Set(ctx, awaiSenseiSchedule{
		TweetID:  s.TweetID,
		Title:    s.Title,
		ImageURL: s.ImageURL,
	})
	return err
}
