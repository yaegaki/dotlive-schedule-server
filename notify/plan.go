package notify

import (
	"context"
	"fmt"

	"github.com/yaegaki/dotlive-schedule-server/model"
)

// PushNotifyPlan 計画をプッシュ通知する
func PushNotifyPlan(ctx context.Context, cli Client, p model.Plan, actors model.ActorSlice) error {
	d := p.Date
	topic := "plan"
	title := fmt.Sprintf("生放送スケジュール%v月%v日", int(d.Month()), d.Day())
	body := p.Text
	msg := createMessage(topic, title, body, map[string]string{
		"date": fmt.Sprintf("%v-%v-%v", d.Year(), int(d.Month()), d.Day()),
	})
	_, err := cli.Send(ctx, msg)

	return err
}
