package notify

import (
	"context"
	"fmt"

	"firebase.google.com/go/messaging"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

// PushNotifyVideo 配信をプッシュ通知する
func PushNotifyVideo(ctx context.Context, cli *messaging.Client, v model.Video, actor model.Actor) error {
	topic := "video"
	title := fmt.Sprintf("配信:%v", actor.Name)
	body := v.Text
	_, err := cli.Send(ctx, createMessage(topic, title, body))

	return err
}
