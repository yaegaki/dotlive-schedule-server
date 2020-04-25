package notify

import (
	"context"
	"testing"

	"github.com/yaegaki/dotlive-schedule-server/model"
)

func TestNotifyVideo(t *testing.T) {
	tests := []struct {
		topic string
		title string
		body  string
		actor model.Actor
	}{
		{
			"'video' in topics && 'siro' in topics",
			"配信:電脳少女シロ",
			"video-text",
			model.Actor{
				Name:              "電脳少女シロ",
				TwitterScreenName: "siro",
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			cli := client{
				t: t,
				m: createMessage(tt.topic, tt.title, tt.body),
			}

			PushNotifyVideo(ctx, cli, model.Video{
				Text: tt.body,
			}, tt.actor)
		})
	}

}
