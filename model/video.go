package model

import "github.com/yaegaki/dotlive-schedule-server/jst"

// VideoSource
const (
	// VideoSourceYoutube Youtubeソース
	VideoSourceYoutube = "Youtube"
	// VideoSourceYoutube Bilibiliソース
	VideoSourceBilibili = "Bilibili"
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
	// Notified Push通知送信済みか
	Notified bool
	// StartAt 配信開始時刻
	StartAt jst.Time
}
