package model

import (
	"github.com/yaegaki/dotlive-schedule-server/jst"
)

// VideoSource
const (
	// VideoSourceYoutube Youtubeソース
	VideoSourceYoutube = "Youtube"
	// VideoSourceBilibili Bilibiliソース
	VideoSourceBilibili = "Bilibili"
	// VideoSourceMildom Mildomソース
	VideoSourceMildom = "Mildom"
)

// Video 動画の情報
type Video struct {
	// id 動画ID
	ID string
	// Author 配信者ID
	ActorID string
	// Source 動画サイト
	Source string
	// URL 動画のURL
	URL string
	// Text 動画の説明
	Text string
	// IsLive 生放送かどうか
	// プレミア公開もTrue
	IsLive bool
	// MemberOnly メンバー限定配信かどうか
	MemberOnly bool
	// Notified Push通知送信済みか
	Notified bool
	// StartAt 配信開始時刻
	StartAt jst.Time
	// RelatedActorID 関連する配信者のID
	RelatedActorID string
	// RelatedActorIDs 関連する配信者のIDの配列
	RelatedActorIDs []string
	// OwnerName 動画配信者の名前
	// ほとんどの場合はActorの名前と同じ
	// コラボ時などActorIDから名前が特定できない時に使用する
	OwnerName string
	// HashTags 配信の関連するハッシュタグ
	HashTags []string
}

// IsUnknownActor 配信者不明かどうか
func (v Video) IsUnknownActor() bool {
	return v.ActorID == ActorIDUnknown
}
