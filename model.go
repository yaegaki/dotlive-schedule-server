package main

import (
	"time"
)

// Actor 配信者の情報
type Actor struct {
	// id 配信者ID
	id string
	// TwitterScreenName Twitterのスクリーンネーム
	TwitterScreenName string `firestore:"twitterScreenName"`
	// Name 配信者名
	Name string `firestore:"name"`
	// Emoji 推しアイコン
	Emoji string `firestore:"emoji"`
	// YoutubeChannelID YoutubeのチャンネルID
	YoutubeChannelID string `firestore:"youtubeChannelID"`
	// BilibiliID BilibiliのID
	BilibiliID string `firestore:"bilibiliID"`
	// LastTweetID 最後に取得したTweetのID
	LastTweetID string `firestore:"lastTweetID"`
	// Hashtag 配信者のハッシュタグ
	Hashtag string `firestore:"hashtag"`
}

// Video 動画の情報
type Video struct {
	// id 動画ID
	id string
	// Author 配信者ID
	ActorID string `firestore:"actorID"`
	// Source 動画サイト
	Source string `firestore:"Source"`
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
}

// Plan 配信スケジュール
type Plan struct {
	// Date 配信日
	Date time.Time `firestore:"date"`
	// Entries 配信予定エントリ
	Entries []PlanEntry `firestore:"entries"`
	// SourceID スケジュールのtweetID
	SourceID string `firestore:"sourceID"`
	// Notified 通知を行ったかどうか
	Notified bool `firestore:"notified"`
}

// PlanEntry 配信スケジュールのエントリ
type PlanEntry struct {
	// ActorID 配信者ID
	ActorID string `firestore:"actorID"`
	// StartAt 配信開始時刻
	StartAt time.Time `firestore:"startAt"`
}

// Schedule 1日の配信予定と配信
type Schedule struct {
	// Date 配信日
	Date    time.Time       `firestore:"date"`
	Entries []ScheduleEntry `firestore:"entries"`
}

// ScheduleEntry Scheduleのエントリ
type ScheduleEntry struct {
	// ActorName 配信者名
	ActorName string `firestore:"actorName"`
	// StartAt 配信時刻/予定時刻
	StartAt time.Time `firestore:"startAt"`
	// VideoID 動画ID
	VideoID string `firestore:"videoID"`
	// URL 配信URL
	URL string `firestore:"url"`
	// Planned 計画配信かどうか
	Planned bool `firestore:"planned"`
	// IsLive 生放送かどうか
	IsLive bool `firestore:"isLive"`
	// Text 説明
	Text string `firestore:"text"`
}
