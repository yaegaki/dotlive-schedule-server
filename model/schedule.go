package model

import "github.com/yaegaki/dotlive-schedule-server/jst"

// Schedule スケジュール
type Schedule struct {
	// Date 配信日
	Date jst.Time `json:"date"`
	// TweetID 元ツイートのID
	TweetID string `json:"tweetId"`
	// Entries エントリー
	Entries []ScheduleEntry `json:"entries"`
}

// ScheduleEntry スケジュールのエントリ
type ScheduleEntry struct {
	// ActorName 配信者名
	ActorName string `json:"actorName"`
	// Note 配信の補足情報
	Note string `json:"note"`
	// Icon 配信者アイコン
	Icon string `json:"icon"`
	// StartAt 配信予定/予定時刻
	StartAt jst.Time `json:"startAt"`
	// VideoID 動画ID
	VideoID string `json:"videoId"`
	// URL 配信URL
	URL string `json:"url"`
	// Source 配信サイト
	Source string `json:"source"`
	// Planned 計画配信かどうか
	Planned bool `json:"planned"`
	// IsLive 生放送かどうか
	IsLive bool `json:"isLive"`
	// MemberOnly メンバー限定かどうか
	MemberOnly bool `json:"memberOnly"`
	// Text 説明
	Text string `json:"text"`
	// CollaboID コラボID
	CollaboID int `json:"collaboId"`
}
