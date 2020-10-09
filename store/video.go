package store

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// video 動画の情報
type video struct {
	// id 動画ID
	id string
	// Author 配信者ID
	ActorID string `firestore:"actorID"`
	// Source 動画サイト
	Source string `firestore:"source"`
	// URL 動画のURL
	URL string `firestore:"url"`
	// Text 動画の説明
	Text string `firestore:"text"`
	// IsLive 生放送かどうか
	// プレミア公開もTrue
	IsLive bool `firestore:"isLive"`
	// Notified Push通知送信済みか
	Notified bool `firestore:"notified"`
	// StartAt 配信開始時刻
	StartAt time.Time `firestore:"startAt"`
	// RelatedActorID 関連する配信者ID
	RelatedActorID string `firestore:"relatedActorID"`
	// RelatedActorIDs 関連する配信者IDの配列
	RelatedActorIDs []string `firestore:"relatedActorIDs"`
	// OwnerName 動画配信者の名前
	OwnerName string `firestore:"ownerName"`
	// HashTags ハッシュタグ
	HashTags []string `firestore:"hashTags"`
}

const collectionNameVideo = "Video"

// FindVideos 開始時刻と終了時刻を指定して動画を検索する
func FindVideos(ctx context.Context, c *firestore.Client, r jst.Range) ([]model.Video, error) {
	it := c.Collection(collectionNameVideo).Where("startAt", ">=", r.Begin.Time()).Where("startAt", "<=", r.End.Time()).Documents(ctx)
	return getVideos(it)
}

// FindNotNotifiedVideos 通知していない動画を取得する
func FindNotNotifiedVideos(ctx context.Context, c *firestore.Client) ([]model.Video, error) {
	it := c.Collection(collectionNameVideo).Where("notified", "==", false).Documents(ctx)
	return getVideos(it)
}

func getVideos(it *firestore.DocumentIterator) ([]model.Video, error) {
	var videos []model.Video
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var v video
		doc.DataTo(&v)
		v.id = doc.Ref.ID
		videos = append(videos, v.Video())
	}

	return videos, nil
}

// SaveVideo 動画を保存する
// 既に存在している場合は通知設定は更新されない
func SaveVideo(ctx context.Context, c *firestore.Client, v model.Video, overrideOldVideoHandler func(v model.Video) bool) error {
	temp := fromVideo(v)

	return c.RunTransaction(ctx, func(ctx context.Context, t *firestore.Transaction) error {
		docRef := c.Collection(collectionNameVideo).Doc(v.ID)
		doc, err := t.Get(docRef)

		// 既に存在する場合は通知設定を引き継ぐ
		if err == nil {
			var oldVideo video
			doc.DataTo(&oldVideo)
			oldVideo.id = doc.Ref.ID

			if overrideOldVideoHandler != nil && !overrideOldVideoHandler(oldVideo.Video()) {
				return nil
			}

			// 古い動画の配信者IDが分かっている場合かつ新しい動画の配信者IDが分からない場合は更新しない
			// コラボ配信などで一人のチャンネルでしか配信しない場合、
			// チャンネル主のツイートの動画を保存した方がいいため
			if oldVideo.ActorID != model.ActorIDUnknown && v.ActorID == model.ActorIDUnknown {
				return nil
			}

			temp.Notified = oldVideo.Notified
			temp.RelatedActorIDs = createRelatedActorIDs(temp, oldVideo)
		} else if status.Code(err) != codes.NotFound {
			return err
		}

		return t.Set(c.Collection(collectionNameVideo).Doc(v.ID), temp)
	})
}

func createRelatedActorIDs(v1 video, v2 video) []string {
	var result []string
	add := func(id string) {
		if id == "" || id == model.ActorIDUnknown {
			return
		}
		for _, temp := range result {
			if temp == id {
				return
			}
		}
		result = append(result, id)
	}

	add(v1.ActorID)
	add(v1.RelatedActorID)

	for _, actorID := range v1.RelatedActorIDs {
		add(actorID)
	}

	add(v2.ActorID)
	add(v2.RelatedActorID)

	for _, actorID := range v2.RelatedActorIDs {
		add(actorID)
	}
	return result
}

// MarkVideoAsNotified 計画を通知済みとする
// すでに通知済みな場合はなにもしない
// 更新された場合はtrue、されなかった場合はfalse
func MarkVideoAsNotified(ctx context.Context, c *firestore.Client, v model.Video) (model.Video, bool, error) {
	updated := false
	var temp model.Video

	err := c.RunTransaction(ctx, func(ctx context.Context, t *firestore.Transaction) error {
		updated = false
		docRef := c.Collection(collectionNameVideo).Doc(v.ID)
		doc, err := t.Get(docRef)

		if err != nil {
			if status.Code(err) != codes.NotFound {
				return err
			}

			// 存在しない場合は何もしない
			return nil
		}

		updated = true
		var oldVideo video
		doc.DataTo(&oldVideo)
		oldVideo.Notified = true
		temp = oldVideo.Video()

		return t.Set(doc.Ref, oldVideo)
	})

	if err != nil {
		return model.Video{}, false, err
	}

	if !updated {
		return v, false, nil
	}

	return temp, true, nil
}

func fromVideo(v model.Video) video {
	return video{
		id:              v.ID,
		ActorID:         v.ActorID,
		Source:          v.Source,
		URL:             v.URL,
		Text:            v.Text,
		IsLive:          v.IsLive,
		Notified:        v.Notified,
		StartAt:         v.StartAt.Time(),
		RelatedActorID:  v.RelatedActorID,
		RelatedActorIDs: v.RelatedActorIDs,
		OwnerName:       v.OwnerName,
		HashTags:        v.HashTags,
	}
}

func (v video) Video() model.Video {
	return model.Video{
		ID:              v.id,
		ActorID:         v.ActorID,
		Source:          v.Source,
		URL:             v.URL,
		Text:            v.Text,
		IsLive:          v.IsLive,
		Notified:        v.Notified,
		StartAt:         jst.From(v.StartAt),
		RelatedActorID:  v.RelatedActorID,
		RelatedActorIDs: v.RelatedActorIDs,
		OwnerName:       v.OwnerName,
		HashTags:        v.HashTags,
	}
}
