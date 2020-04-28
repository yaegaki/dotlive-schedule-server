package notify

import (
	"context"

	"firebase.google.com/go/messaging"
)

// TestNotifyClient .
type TestNotifyClient struct {
	Messages []*messaging.Message
}

// Send .
func (c *TestNotifyClient) Send(ctx context.Context, message *messaging.Message) (string, error) {
	m := *message
	c.Messages = append(c.Messages, &m)
	return "", nil
}
