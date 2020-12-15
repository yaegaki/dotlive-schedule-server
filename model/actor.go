package model

import "github.com/yaegaki/dotlive-schedule-server/common"

const (
	// ActorIDUnknown コラボの時のID
	ActorIDUnknown = "UNKNOWN"
)

// Actor 配信者
type Actor struct {
	// ID 配信者ID
	ID string
	// Name 名前
	Name string
	// Icon アイコンURL
	Icon string
	// Hashtag ハッシュタグ
	Hashtag string
	// TwitterScreenName Twitterのスクリーンネーム
	TwitterScreenName string
	// Emoji 推しアイコン
	Emoji string
	// YoutubeChannelID YoutubeのチャンネルID
	YoutubeChannelID string
	// YoutubeChannelID YoutubeのチャンネルID
	YoutubeChannelName string
	// BilibiliID BilibiliのID
	BilibiliID string
	// MildomID MildomのID
	MildomID string
	// LastTweetID 最後に取得したTweetのID
	LastTweetID string
}

// ActorSlice Actorのスライス
type ActorSlice []Actor

// FindActor 配信者を探す
func (s ActorSlice) FindActor(id string) (Actor, error) {
	for _, a := range s {
		if a.ID == id {
			return a, nil
		}
	}

	return Actor{}, common.ErrNotFound
}

// FindActorByName 配信者を探す
func (s ActorSlice) FindActorByName(name string) (Actor, error) {
	for _, a := range s {
		if a.Name == name {
			return a, nil
		}
	}

	return Actor{}, common.ErrNotFound
}
