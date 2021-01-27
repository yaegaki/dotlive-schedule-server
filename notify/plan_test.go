package notify

import (
	"context"
	"fmt"
	"testing"

	"firebase.google.com/go/messaging"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/actor"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/notify"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/plan"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
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

	for key, s := range expect.Data {
		temp, ok := got.Data[key]
		if !ok {
			t.Errorf("data, missing key: %v", key)
			continue
		}

		if temp != s {
			t.Errorf("data[%v], got: %v expect: %v", key, temp, s)
		}
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
			jst.ShortDate(2020, 4, 19),
			[]EntryPart{
				CreateEntryPartCollabo(Chieri, 20, 0, 1),
				CreateEntryPartCollabo(Pino, 20, 0, 1),
				CreateEntryPartCollabo(Iroha, 20, 0, 1),
				CreateEntryPartCollabo(Mememe, 20, 0, 1),
				CreateEntryPart(Suzu, 22, 0),
			},
			"生放送スケジュール4月19日(🍒🐜💎🐏🍋)",
			"20:00~:#花京院ちえり x #カルロピノ x #金剛いろは x #もこ田めめめ\n22:00~:#神楽すず",
		},
		{
			jst.ShortDate(2099, 4, 19),
			[]EntryPart{
				CreateEntryPartCollabo(Chieri, 20, 0, 1),
				CreateEntryPartCollabo(Pino, 20, 0, 1),
				CreateEntryPartCollabo(Iori, 20, 0, 2),
				CreateEntryPartCollabo(Suzu, 20, 0, 2),
				CreateEntryPartCollabo(Mememe, 21, 0, 3),
				CreateEntryPartCollabo(Iroha, 21, 0, 3),
				CreateEntryPart(Suzu, 22, 0),
				CreateEntryPartCollabo(Mememe, 23, 0, 4),
				CreateEntryPartCollabo(Iroha, 23, 0, 4),
			},
			"生放送スケジュール4月19日(🍒🐜🍄🍋🐏💎)",
			"20:00~:#花京院ちえり x #カルロピノ\n20:00~:#ヤマトイオリ x #神楽すず\n21:00~:#もこ田めめめ x #金剛いろは\n22:00~:#神楽すず\n23:00~:#もこ田めめめ x #金剛いろは",
		},
		{
			jst.ShortDate(2020, 4, 24),
			[]EntryPart{
				CreateEntryPartBilibili(Siro, 19, 0),
				CreateEntryPart(Suzu, 22, 0),
			},
			"生放送スケジュール4月24日(🐬🍋)",
			"19:00~:#シロ生放送(bilibili)\n22:00~:#神楽すず",
		},
		{
			jst.ShortDate(2099, 4, 1),
			[]EntryPart{},
			"生放送スケジュール4月1日",
			"",
		},
		{
			jst.ShortDate(2099, 4, 2),
			[]EntryPart{
				CreateEntryPart(Siro, 19, 0),
				CreateEntryPart(Suzu, 22, 0),
			},
			"生放送スケジュール4月2日(🐬🍋)",
			"19:00~:#シロ生放送\n22:00~:#神楽すず",
		},
		{
			jst.ShortDate(2099, 4, 2),
			[]EntryPart{
				CreateEntryPart(Siro, 19, 0),
				CreateEntryPartMildom(Suzu, 22, 0),
			},
			"生放送スケジュール4月2日(🐬🍋)",
			"19:00~:#シロ生放送\n22:00~:#神楽すず(Mildom)",
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			cli := &TestNotifyClient{}

			p := CreatePlan(tt.date, tt.parts)
			p.Texts = []model.PlanText{
				{
					Date:    jst.ShortDate(2021, 1, 1),
					PlanTag: "",
					Text:    tt.body,
				},
			}
			PushNotifyPlan(ctx, cli, p, All)
			if len(cli.Messages) != 1 {
				t.Errorf("inavalid len(cli.Messages), got: %v", len(cli.Messages))
				return
			}

			body := tt.body
			if body == "" {
				body = "なし"
			}

			expect := createMessage("plan", tt.title, body, map[string]string{
				"date": fmt.Sprintf("%v-%v-%v", tt.date.Year(), int(tt.date.Month()), tt.date.Day()),
			})
			comparePlanMessage(t, cli.Messages[0], expect)
		})
	}
}
