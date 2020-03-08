package notify

import (
	"context"
	"fmt"

	"firebase.google.com/go/messaging"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

// PushNotifyVideo 配信をプッシュ通知する
func PushNotifyVideo(ctx context.Context, cli *messaging.Client, v model.Video, actor model.Actor) error {
	_, err := cli.Send(ctx, &messaging.Message{
		Topic: "video",
		Notification: &messaging.Notification{
			Title: fmt.Sprintf("配信:%v", actor.Name),
			Body:  v.Text,
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	})

	return err
}
