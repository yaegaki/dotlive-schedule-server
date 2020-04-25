package notify

import (
	"context"
	"fmt"

	"github.com/yaegaki/dotlive-schedule-server/model"
)

// PushNotifyVideo 配信をプッシュ通知する
func PushNotifyVideo(ctx context.Context, cli Client, v model.Video, actor model.Actor) error {
	condition := fmt.Sprintf("'video' in topics && '%v' in topics", actor.TwitterScreenName)
	title := fmt.Sprintf("配信:%v", actor.Name)
	body := v.Text
	_, err := cli.Send(ctx, createMessageWithCondition(condition, title, body))

	return err
}
