package model

import "github.com/yaegaki/dotlive-schedule-server/jst"

// Calendar カレンダー
type Calendar struct {
	// BaseDate 何月か
	BaseDate jst.Time `json:"baseDate"`
	// Entries エントリー
	Days CalendarDaySlice `json:"days"`
}

// CalendarDay カレンダーの1日
type CalendarDay struct {
	// Day 何日か
	Day     int                   `json:"day"`
	Entries CalendarDayEntrySlice `json:"entries"`
}

// CalendarDaySlice CalendarDayのスライス
type CalendarDaySlice []CalendarDay

// CalendarDayEntry CalendarDayのエントリー
type CalendarDayEntry struct {
	// ActorIndex 配信者ID
	ActorID string `json:"actorId"`
	// Text テキスト
	Text string `json:"text"`
	// URL 動画URL
	URL string `json:"url"`
}

// CalendarDayEntrySlice CalendarDayEntryのスライス
type CalendarDayEntrySlice []CalendarDayEntry
