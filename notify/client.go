package notify

import (
	"context"

	"firebase.google.com/go/messaging"
)

// Client プッシュ通知クライアント
type Client interface {
	Send(ctx context.Context, message *messaging.Message) (string, error)
}
