package model

import "github.com/yaegaki/dotlive-schedule-server/jst"

// Calendar カレンダー
type Calendar struct {
	// BaseDate 何月か
	BaseDate jst.Time `json:"baseDate"`
	// Entries エントリー
	Days CalendarDaySlice `json:"days"`
	// FixedDay この日以前の日付けは変更されることがない
	// カレンダーをキャッシュする場合、
	// FixedDay以前の情報はキャッシュしても変更されないことが保証されるが、
	// FixedDay以降の情報は更新される可能性がある
	FixedDay int `json:"fixedDay"`
}

// CalendarDay カレンダーの1日
type CalendarDay struct {
	// Day 何日か
	Day      int      `json:"day"`
	ActorIDs []string `json:"actorIds"`
}

// CalendarDaySlice CalendarDayのスライス
type CalendarDaySlice []CalendarDay
