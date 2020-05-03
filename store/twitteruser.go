package store

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// actor 配信者
type twitterUser struct {
	// id 配信者ID
	id string
	// LastTweetID 最後に取得したTweetのID
	LastTweetID string `firestore:"lastTweetID"`
}

const collectionNameTwitterUser = "TwitterUser"

// FindTwitterUser TwitterUserを検索する
func FindTwitterUser(ctx context.Context, c *firestore.Client, screenName string) (model.TwitterUser, error) {
	docRef := c.Collection(collectionNameTwitterUser).Doc(screenName)
	doc, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return model.TwitterUser{
				ScreenName:  screenName,
				LastTweetID: "",
			}, nil
		}

		return model.TwitterUser{}, err
	}

	var user model.TwitterUser
	doc.DataTo(&user)
	user.ScreenName = screenName

	return user, nil
}

// SaveTwitterUser 配信者を保存する
func SaveTwitterUser(ctx context.Context, c *firestore.Client, u model.TwitterUser) error {
	// 常に上書きでいいのでトランザクションにしない
	_, err := c.Collection(collectionNameTwitterUser).Doc(u.ScreenName).Set(ctx, twitterUser{
		LastTweetID: u.LastTweetID,
	})
	return err
}
