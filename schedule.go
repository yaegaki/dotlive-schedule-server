package main

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
)

func createSchedule(p Plan, videos []Video, actors []Actor) Schedule {
	s := Schedule{
		Date: p.Date,
	}

	dateJST := p.Date.In(jst)

	for _, v := range videos {
		t := v.StartAt.In(jst)
		// Planと同じ日か？
		if !dateJST.Equal(createJSTTime(t.Year(), t.Month(), t.Day(), 0, 0)) {
			continue
		}

		planned := false
		for _, e := range p.Entries {
			if v.ActorID != e.ActorID {
				continue
			}

			// 計画の+-25分以内なら計画配信とする
			begin := e.StartAt.Add(-25 * time.Minute)
			end := e.StartAt.Add(25 * time.Minute)
			if !between(t, begin, end) {
				continue
			}
			planned = true
			break
		}

		actor, err := findActor(v.ActorID, actors)
		if err != nil {
			continue
		}

		// シロちゃんの動画は常に計画されているとする
		const siroID = "lLhToxu1Kyxuwwygh0FK"
		if v.ActorID == siroID && !v.IsLive {
			planned = true
		}

		s.Entries = append(s.Entries, ScheduleEntry{
			StartAt:   v.StartAt,
			ActorName: actor.Name,
			Planned:   planned,
			IsLive:    v.IsLive,
			URL:       v.URL,
			VideoID:   v.id,
			Text:      v.Text,
		})
	}

	for _, e := range p.Entries {
		begin := e.StartAt.Add(-25 * time.Minute)
		end := e.StartAt.Add(25 * time.Minute)
		added := false

		actor, err := findActor(e.ActorID, actors)
		if err != nil {
			continue
		}

		// TODO: bilibiliの開始時間があれであれ

		for _, se := range s.Entries {
			if !se.Planned {
				continue
			}

			if actor.Name != se.ActorName {
				continue
			}

			if between(se.StartAt, begin, end) {
				added = true
				break
			}
		}

		if added {
			continue
		}

		// まだ動画URLが確定していない場合

		s.Entries = append(s.Entries, ScheduleEntry{
			ActorName: actor.Name,
			StartAt:   e.StartAt.In(jst),
			Planned:   true,
		})
	}

	return s
}

func findSchedule(ctx context.Context, c *firestore.Client, date time.Time) (Schedule, error) {
	plan, err := findPlan(ctx, c, date)
	if err != nil {
		return Schedule{}, err
	}

	actors, err := findActors(ctx, c)
	if err != nil {
		return Schedule{}, err
	}

	videos, err := findVideos(ctx, c, date)
	if err != nil {
		return Schedule{}, err
	}

	return createSchedule(plan, videos, actors), nil
}

func findCachedSchedule(ctx context.Context, c *firestore.Client, date time.Time) (Schedule, error) {
	key := createDayKey(date)
	doc, err := c.Collection("Schedule").Doc(key).Get(ctx)
	if err != nil {
		return Schedule{}, err
	}

	var s Schedule
	doc.DataTo(&s)
	return s, nil
}
