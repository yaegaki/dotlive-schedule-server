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
	Day      int      `json:"day"`
	ActorIDs []string `json:"actorIDs"`
}

// CalendarDaySlice CalendarDayのスライス
type CalendarDaySlice []CalendarDay
