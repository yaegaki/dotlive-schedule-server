package main

import (
	"time"
)

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

func createJSTTime(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, jst)
}

func getDay(t time.Time) time.Time {
	return createJSTTime(t.Year(), t.Month(), t.Day(), 0, 0)
}

func between(t time.Time, begin time.Time, end time.Time) bool {
	if t.Equal(begin) || t.Equal(end) {
		return true
	}

	if t.After(begin) && t.Before(end) {
		return true
	}

	return false
}
