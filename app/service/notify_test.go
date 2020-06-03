package service

import (
	"context"
	"testing"

	"firebase.google.com/go/messaging"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/actor"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/notify"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/plan"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

type notifyVideoTestClient struct {
	m *messaging.Message
}

func TestPushNotifyVideoInternal(t *testing.T) {
	tests := []struct {
		d      jst.Time
		title  string
		body   string
		videos []model.Video
	}{
		{
			jst.Date(2020, 4, 29, 20, 0),
			"コラボ配信:🍄🍋",
			"collabo",
			[]model.Video{
				{
					ID:      "aaa",
					ActorID: Suzu.ID,
					StartAt: jst.Date(2020, 4, 29, 0, 0),
					Text:    "past",
					URL:     "https://past",
					Source:  model.VideoSourceYoutube,
				},
				{
					ID:      "video-id-1",
					ActorID: Iori.ID,
					StartAt: jst.Date(2020, 4, 29, 20, 0),
					Text:    "collabo",
					URL:     "https://1",
					Source:  model.VideoSourceYoutube,
				},
			},
		},
		{
			jst.Date(2020, 4, 29, 22, 0),
			"配信:カルロピノ",
			"solo",
			[]model.Video{
				{
					ID:      "aaa",
					ActorID: Suzu.ID,
					StartAt: jst.Date(2020, 4, 29, 0, 0),
					Text:    "past",
					URL:     "https://past",
					Source:  model.VideoSourceYoutube,
				},
				{
					ID:      "video-id-2",
					ActorID: Pino.ID,
					StartAt: jst.Date(2020, 4, 29, 22, 0),
					Text:    "solo",
					URL:     "https://2",
					Source:  model.VideoSourceYoutube,
				},
			},
		},
	}

	baseDate := jst.ShortDate(2020, 4, 29)
	plans := []model.Plan{
		CreatePlan(baseDate, []EntryPart{
			CreateEntryPartCollabo(Iori, 20, 0, 1),
			CreateEntryPartCollabo(Suzu, 20, 0, 1),
			CreateEntryPart(Pino, 22, 0),
		}),
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			cli := &TestNotifyClient{}
			pushNotifyVideoInternal(ctx, cli, plans, tt.videos, All, tt.d, func(ctx context.Context, v model.Video) (model.Video, bool, error) {
				return v, true, nil
			})

			if len(cli.Messages) != 1 {
				t.Errorf("len(messages), got: %v", len(cli.Messages))
				return
			}

			m := cli.Messages[0]
			if m.Notification.Title != tt.title {
				t.Errorf("title, got: %v expect: %v", m.Notification.Title, tt.title)
			}

			if m.Notification.Body != tt.body {
				t.Errorf("body, got: %v expect: %v", m.Notification.Body, tt.body)
			}
		})
	}
}
