package main

import "errors"

var (
	// ErrNotPlanTweet 生放送スケジュールではない
	ErrNotPlanTweet = errors.New("Not plan")
	// ErrNotFound 見つからなかった
	ErrNotFound = errors.New("Not found")
)
