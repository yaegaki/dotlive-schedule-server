package notify

import (
	"context"
	"testing"

	"firebase.google.com/go/messaging"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/actor"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/notify"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/plan"
	"github.com/yaegaki/dotlive-schedule-server/jst"
)

func comparePlanMessage(t *testing.T, got *messaging.Message, expect *messaging.Message) (string, error) {
	if got.Topic != expect.Topic {
		t.Errorf("topic, got: %v expect: %v", got.Topic, expect.Topic)
	}

	if got.Notification.Title != expect.Notification.Title {
		t.Errorf("title, got: %v expect: %v", got.Notification.Title, expect.Notification.Title)
	}

	if got.Notification.Body != expect.Notification.Body {
		t.Errorf("body, got: %v expect: %v", got.Notification.Body, expect.Notification.Body)
	}

	return "", nil
}

func TestNotifyPlan(t *testing.T) {
	tests := []struct {
		date  jst.Time
		parts []EntryPart
		title string
		body  string
	}{
		{
			jst.ShortDate(2020, 4, 24),
			[]EntryPart{
				CreateEntryPartBilibili(Siro, 19, 0),
				CreateEntryPart(Suzu, 22, 0),
			},
			"生放送スケジュール4月24日",
			"19:00~:#シロ生放送(bilibili)\n22:00~:#神楽すず",
		},
		{
			jst.ShortDate(2099, 4, 1),
			[]EntryPart{},
			"生放送スケジュール4月1日",
			"なし",
		},
		{
			jst.ShortDate(2099, 4, 2),
			[]EntryPart{
				CreateEntryPart(Siro, 19, 0),
				CreateEntryPart(Suzu, 22, 0),
			},
			"生放送スケジュール4月2日",
			"19:00~:#シロ生放送\n22:00~:#神楽すず",
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			cli := &TestNotifyClient{}

			PushNotifyPlan(ctx, cli, CreatePlan(tt.date, tt.parts), All)
			if len(cli.Messages) != 1 {
				t.Errorf("inavalid len(cli.Messages), got: %v", len(cli.Messages))
				return
			}

			expect := createMessage("plan", tt.title, tt.body)
			comparePlanMessage(t, cli.Messages[0], expect)
		})
	}
}
