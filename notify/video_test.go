package notify

import (
	"context"
	"strings"
	"testing"

	"firebase.google.com/go/messaging"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/actor"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/notify"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

func testNotifyVideoMessages(t *testing.T, got []*messaging.Message, expectTitle, expectBody string, expectConditions []string, date string) {
	matchConditions := map[string]bool{}

	for _, m := range got {
		if m.Notification.Title != expectTitle {
			t.Errorf("title, got: %v expect: %v", m.Notification.Title, expectTitle)
		}

		if m.Notification.Body != expectBody {
			t.Errorf("body, got: %v expect: %v", m.Notification.Body, expectBody)
		}

		if d, ok := m.Data["date"]; !ok || d != date {
			t.Errorf("data, got: %v, expect: %v", d, date)
		}

		if m.Condition == "" {
			t.Errorf("condition is empty")
			return
		}

		conditions := strings.Split(m.Condition, " || ")
		if len(conditions) > maxTopicCount {
			t.Errorf("invalid condition(condition must be less than maxTopicCount): %v", m.Condition)
			return
		}

		for _, cond := range conditions {
			if _, ok := matchConditions[cond]; ok {
				t.Errorf("invalid condition(duplicate): %v", cond)
				return
			}

			found := false
			for _, expectCond := range expectConditions {
				if cond == expectCond {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("invalid condition(not expected): %v", cond)
				return
			}

			matchConditions[cond] = true
		}
	}

	if len(matchConditions) != len(expectConditions) {
		t.Errorf("invalid condition(not enough), got: %v expect: %v", len(matchConditions), len(expectConditions))
	}
}

func TestNotifyVideo(t *testing.T) {
	tests := []struct {
		conditions []string
		title      string
		body       string
		actors     []model.Actor
	}{
		{
			[]string{
				"'test-siro' in topics",
			},
			"配信:電脳少女シロ",
			"video-text",
			[]model.Actor{
				Siro,
			},
		},
		{
			[]string{
				"'test-iori' in topics",
				"'test-suzu' in topics",
			},
			"コラボ配信:🍄🍋",
			"video-text",
			[]model.Actor{
				Iori,
				Suzu,
			},
		},
		{
			[]string{
				"'test-iori' in topics",
				"'test-pino' in topics",
				"'test-suzu' in topics",
				"'test-chieri' in topics",
				"'test-iroha' in topics",
				"'test-futaba' in topics",
			},
			"コラボ配信:🍄🐜🍋🍒💎🌱",
			"video-text",
			[]model.Actor{
				Iori,
				Pino,
				Suzu,
				Chieri,
				Iroha,
				Futaba,
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			cli := &TestNotifyClient{}

			PushNotifyVideo(ctx, cli, jst.ShortDate(2020, 5, 11), model.Video{
				Text: tt.body,
			}, tt.actors)

			testNotifyVideoMessages(t, cli.Messages, tt.title, tt.body, tt.conditions, "2020-5-11")
		})
	}
}
