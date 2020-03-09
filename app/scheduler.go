package app

import (
	"context"
	"log"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/common"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/store"
)

// createSchedules スケジュールを作成する
func createSchedule(ctx context.Context, c *firestore.Client, date jst.Time, actors []model.Actor) (model.Schedule, error) {
	r := jst.Range{
		Begin: date.AddDay(-2),
		End:   date.AddOneDay(),
	}

	plans, err := store.FindPlans(ctx, c, r)
	// 予定が見つからなかった場合は空で返す
	if err == common.ErrNotFound {
		return createEmptySchedule(date), nil
	}

	if err != nil {
		return model.Schedule{}, err
	}

	videos, err := store.FindVideos(ctx, c, r)
	return createScheduleInternal(date.FloorToDay(), plans, videos, actors)
}

func createEmptySchedule(date jst.Time) model.Schedule {
	return model.Schedule{
		Date:    date.FloorToDay(),
		Entries: []model.ScheduleEntry{},
	}
}

func createScheduleInternal(date jst.Time, plans []model.Plan, videos []model.Video, actors []model.Actor) (model.Schedule, error) {
	var targetPlan model.Plan
	found := false
	for _, p := range plans {
		if date.Equal(p.Date) {
			found = true
			targetPlan = p
			break
		}
	}

	if !found {
		return createEmptySchedule(date), nil
	}

	scheduleRange := jst.Range{
		Begin: targetPlan.Date,
		End:   targetPlan.Date.AddOneDay().Add(-1 * time.Second),
	}

	for _, e := range targetPlan.Entries {
		if e.StartAt.After(scheduleRange.End) {
			scheduleRange.End = e.StartAt
		}
	}

	scheduleRange.End = scheduleRange.End.Add(30 * time.Minute)
	entries := []model.ScheduleEntry{}
	var addedPlanEntries []model.PlanEntry

	for _, v := range videos {
		yesterdayPlanned := false
		// 前日に予定された配信かどうか
		for _, p := range plans {
			if p.Date.Equal(targetPlan.Date) {
				break
			}

			if p.IsPlanned(v) {
				yesterdayPlanned = true
				break
			}
		}

		if yesterdayPlanned {
			continue
		}

		var isPlanned bool
		startAt := v.StartAt
		pe, err := targetPlan.GetEntry(v)
		if err == nil {
			isPlanned = true
			if v.Source == model.VideoSourceBilibili {
				startAt = pe.StartAt
			}
			addedPlanEntries = append(addedPlanEntries, pe)
		} else {
			isPlanned = false

			if !scheduleRange.In(v.StartAt) {
				continue
			}

			// 明日の計画を確認してゲリラ配信かどうか確認する
			tommorowPlanned := false
			after := false
			for _, p := range plans {
				if after {
					if p.IsPlanned(v) {
						tommorowPlanned = true
						break
					}
				} else {
					after = p.Date.Equal(targetPlan.Date)
				}
			}

			if tommorowPlanned {
				continue
			}
		}

		actor, err := findActorByID(actors, v.ActorID)
		if err != nil {
			log.Printf("Unknown actorID: %v", v.ActorID)
			continue
		}

		// シロちゃんの動画は常に計画されているとする
		const siroID = "lLhToxu1Kyxuwwygh0FK"
		if v.ActorID == siroID && !v.IsLive {
			isPlanned = true
		}

		se := model.ScheduleEntry{
			ActorName: actor.Name,
			Icon:      actor.Icon,
			StartAt:   startAt,
			Planned:   isPlanned,
			IsLive:    v.IsLive,
			Text:      v.Text,
			URL:       v.URL,
			VideoID:   v.ID,
		}
		entries = append(entries, se)
	}

	for _, e := range targetPlan.Entries {
		found := false
		for _, added := range addedPlanEntries {
			if e.ActorID == added.ActorID && e.StartAt.Equal(added.StartAt) {
				found = true
				break
			}
		}

		if found {
			continue
		}

		actor, err := findActorByID(actors, e.ActorID)
		if err != nil {
			log.Printf("Unknown actorID: %v", e.ActorID)
			continue
		}

		se := model.ScheduleEntry{
			ActorName: actor.Name,
			Icon:      actor.Icon,
			StartAt:   e.StartAt,
			Planned:   true,
		}
		entries = append(entries, se)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].StartAt.Before(entries[j].StartAt)
	})

	return model.Schedule{
		Date:    targetPlan.Date,
		Entries: entries,
	}, nil
}

func findActorByID(actors []model.Actor, id string) (model.Actor, error) {
	for _, a := range actors {
		if a.ID == id {
			return a, nil
		}
	}

	return model.Actor{}, common.ErrNotFound
}
