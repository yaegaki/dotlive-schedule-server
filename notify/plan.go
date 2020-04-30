package notify

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/yaegaki/dotlive-schedule-server/jst"
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
	lastCollaboID := 0
	var collaboActors []string
	var lastCollaboEntry model.PlanEntry
	addCollaboToLines := func() {
		t := strings.Join(collaboActors, " x ")
		l := formatPlanEntry(lastCollaboEntry.StartAt, t, lastCollaboEntry.Source)
		lines = append(lines, l)
		lastCollaboID = 0
		collaboActors = nil
		lastCollaboEntry = model.PlanEntry{}
	}

	for _, e := range p.Entries {
		a, err := actors.FindActor(e.ActorID)
		if err != nil {
			log.Printf("Unknown Actor: %v", a.ID)
			continue
		}

		if lastCollaboID > 0 {
			if e.CollaboID != lastCollaboID {
				addCollaboToLines()
			}
		}

		if e.CollaboID > 0 {
			lastCollaboID = e.CollaboID
			lastCollaboEntry = e
			collaboActors = append(collaboActors, a.Hashtag)
			continue
		}

		lines = append(lines, formatPlanEntry(e.StartAt, a.Hashtag, e.Source))
	}

	if lastCollaboID > 0 {
		addCollaboToLines()
	}

	if len(lines) == 0 {
		return "なし"
	}

	return strings.Join(lines, "\n")
}

func formatPlanEntry(startAt jst.Time, hashtag, source string) string {
	s := fmt.Sprintf("%02d:%02d~:%v", startAt.Hour(), startAt.Minute(), hashtag)
	if source == model.VideoSourceBilibili {
		s = s + "(bilibili)"
	}
	return s
}
