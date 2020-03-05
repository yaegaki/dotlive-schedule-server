package common

import "errors"

var (
	// ErrNotFound 見つからなかった
	ErrNotFound = errors.New("Not found")
	// ErrInvalidChannel 動画が対象配信者の物じゃない
	ErrInvalidChannel = errors.New("Invalid channel")
)
