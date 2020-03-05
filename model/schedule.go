package model

import "github.com/yaegaki/dotlive-schedule-server/jst"

// Schedule スケジュール
type Schedule struct {
	// Date 配信日
	Date    jst.Time
	Entries []ScheduleEntry
}

// ScheduleEntry スケジュールのエントリ
type ScheduleEntry struct {
	// ActorName 配信者名
	ActorName string
	// StartAt 配信予定/予定時刻
	StartAt jst.Time
	// VideoID 動画ID
	VideoID string
	// URL 配信URL
	URL string
	// Planned 計画配信かどうか
	Planned bool
	// IsLive 生放送かどうか
	IsLive bool
	// Text 説明
	Text string
}
