package notify

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/yaegaki/dotlive-schedule-server/model"
)

// PushNotifyPlan 計画をプッシュ通知する
func PushNotifyPlan(ctx context.Context, cli Client, p model.Plan, actors model.ActorSlice) error {
	d := p.Date
	topic := "plan"
	title := fmt.Sprintf("生放送スケジュール%v月%v日", int(d.Month()), d.Day())
	body := createNotifyPlanBody(p, actors)
	_, err := cli.Send(ctx, createMessage(topic, title, body))

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
	s := fmt.Sprintf("%02d:%02d~:%v", e.StartAt.Hour(), e.StartAt.Minute(), a.Hashtag)
	if e.Source == model.VideoSourceBilibili {
		s = s + "(bilibili)"
	}
	return s
}
