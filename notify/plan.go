package notify

import (
	"context"
	"fmt"
	"strings"

	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

// PushNotifyPlan 計画をプッシュ通知する
func PushNotifyPlan(ctx context.Context, cli Client, p model.Plan, actors model.ActorSlice) error {
	d := p.Date
	topic := "plan"
	title := createTitle(d, p, actors)
	body := createBody(p.Text())
	msg := createMessage(topic, title, body, map[string]string{
		"date": fmt.Sprintf("%v-%v-%v", d.Year(), int(d.Month()), d.Day()),
	})
	_, err := cli.Send(ctx, msg)

	return err
}

func createTitle(d jst.Time, p model.Plan, actors model.ActorSlice) string {
	emojis := []string{}
OUTER:
	for _, e := range p.Entries {
		actor, err := actors.FindActor(e.ActorID)
		if err != nil {
			continue
		}

		for _, emoji := range emojis {
			if emoji == actor.Emoji {
				continue OUTER
			}
		}
		emojis = append(emojis, actor.Emoji)
	}

	var emojiStr string
	if len(emojis) > 0 {
		emojiStr = "(" + strings.Join(emojis, "") + ")"
	} else {
		emojiStr = ""
	}

	return fmt.Sprintf("生放送スケジュール%v月%v日%v", int(d.Month()), d.Day(), emojiStr)
}

func createBody(text string) string {
	if text == "" {
		return "なし"
	}
	return text
}
