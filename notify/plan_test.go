package notify

import (
	"context"
	"testing"

	"firebase.google.com/go/messaging"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/actor"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

type client struct {
	t *testing.T
	m *messaging.Message
}

func (c client) Send(ctx context.Context, message *messaging.Message) (string, error) {
	if message.Topic != c.m.Topic {
		c.t.Errorf("topic, got: %v expect: %v", message.Topic, c.m.Topic)
	}

	if message.Notification.Title != c.m.Notification.Title {
		c.t.Errorf("title, got: %v expect: %v", message.Notification.Title, c.m.Notification.Title)
	}

	if message.Notification.Body != c.m.Notification.Body {
		c.t.Errorf("body, got: %v expect: %v", message.Notification.Body, c.m.Notification.Body)
	}

	return "", nil
}

func TestNotifyPlan(t *testing.T) {
	tests := []struct {
		date    jst.Time
		entries []struct {
			actor  model.Actor
			hour   int
			min    int
			source string
		}
		topic string
		title string
		body  string
	}{
		{
			jst.ShortDate(2020, 4, 24),
			[]struct {
				actor  model.Actor
				hour   int
				min    int
				source string
			}{
				{Siro, 19, 0, model.VideoSourceBilibili},
				{Suzu, 22, 0, model.VideoSourceYoutube},
			},
			"plan",
			"生放送スケジュール4月24日",
			"19:00~:#シロ生放送(bilibili)\n22:00~:#神楽すず",
		},
		{
			jst.ShortDate(2099, 4, 1),
			[]struct {
				actor  model.Actor
				hour   int
				min    int
				source string
			}{},
			"plan",
			"生放送スケジュール4月1日",
			"なし",
		},
		{
			jst.ShortDate(2099, 4, 2),
			[]struct {
				actor  model.Actor
				hour   int
				min    int
				source string
			}{
				{Siro, 19, 0, model.VideoSourceYoutube},
				{Suzu, 22, 0, model.VideoSourceYoutube},
			},
			"plan",
			"生放送スケジュール4月2日",
			"19:00~:#シロ生放送\n22:00~:#神楽すず",
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			cli := client{
				t: t,
				m: createMessage(tt.topic, tt.title, tt.body),
			}

			var entries []model.PlanEntry
			for _, e := range tt.entries {
				entries = append(entries, model.PlanEntry{
					ActorID: e.actor.ID,
					StartAt: jst.Date(tt.date.Year(), tt.date.Month(), tt.date.Day(), e.hour, e.min),
					Source:  e.source,
				})
			}

			PushNotifyPlan(ctx, cli, model.Plan{
				Date:    tt.date,
				Entries: entries,
			}, All)
		})
	}

}
