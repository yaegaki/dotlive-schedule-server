package notify

import (
	"context"
	"fmt"
	"log"
	"strings"

	"firebase.google.com/go/messaging"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

// PushNotifyPlan 計画をプッシュ通知する
func PushNotifyPlan(ctx context.Context, cli *messaging.Client, p model.Plan, actors model.ActorSlice) error {
	d := p.Date
	_, err := cli.Send(ctx, &messaging.Message{
		Topic: "plan",
		Notification: &messaging.Notification{
			Title: fmt.Sprintf("生放送スケジュール%v月%v日", int(d.Month()), d.Day()),
			Body:  createNotifyPlanBody(p, actors),
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

func createNotifyPlanBody(p model.Plan, actors model.ActorSlice) string {
	var lines []string
	for _, e := range p.Entries {
		a, err := actors.FindActor(e.ActorID)
		if err != nil {
			log.Printf("Unknown Actor: %v", a.ID)
			continue
		}

		lines = append(lines, formatPlanEntry(e, a))
	}

	if len(lines) == 0 {
		return "なし"
	}

	return strings.Join(lines, "\n")
}

func formatPlanEntry(e model.PlanEntry, a model.Actor) string {
	return fmt.Sprintf("%02d:%02d~:%v", e.StartAt.Hour(), e.StartAt.Minute(), a.Hashtag)
}
