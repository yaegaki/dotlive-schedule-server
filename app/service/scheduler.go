package service

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

// CreateSchedule スケジュールを作成する
func CreateSchedule(ctx context.Context, c *firestore.Client, date jst.Time, actors []model.Actor) (model.Schedule, error) {
	// スケジュールを作成するためには前日の情報、翌日の12時までの情報が必要
	r := jst.Range{
		Begin: date.AddDay(-1),
		End:   date.AddOneDay().Add(time.Hour * 12),
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
	if err != nil && err != common.ErrNotFound {
		return model.Schedule{}, err
	}

	s := createScheduleInternal(date.FloorToDay(), plans, videos, actors)
	return s, nil
}

func createEmptySchedule(date jst.Time) model.Schedule {
	return model.Schedule{
		Date:    date.FloorToDay(),
		Entries: []model.ScheduleEntry{},
	}
}

func createScheduleInternal(date jst.Time, plans []model.Plan, videos []model.Video, actors []model.Actor) model.Schedule {
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
		return createEmptySchedule(date)
	}

	scheduleRange := jst.Range{
		Begin: targetPlan.Date,
		End:   targetPlan.Date.AddOneDay().Add(-1 * time.Second),
	}

	for _, e := range targetPlan.Entries {
		// 25時などが指定されている場合はスケジュールの終了時間を延ばす
		if e.StartAt.After(scheduleRange.End) {
			scheduleRange.End = e.StartAt
		}
	}

	entries := []model.ScheduleEntry{}
	var addedPlanEntries []model.PlanEntry
	var addedCollaboEntries []model.ScheduleEntry

	// 開始順にソートする
	videos = append([]model.Video{}, videos...)
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].StartAt.Before(videos[j].StartAt)
	})

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
			found := false
			for _, temp := range addedPlanEntries {
				if temp.ActorID == v.ActorID && pe.StartAt.Equal(temp.StartAt) {
					found = true
					break
				}
			}

			// 計画されている場合はその時間に合わせる
			// ただし、既にその計画に対してエントリが追加されている場合は時間を補正しない
			if !found {
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
			Source:    v.Source,
			CollaboID: pe.CollaboID,
		}
		entries = append(entries, se)

		if se.CollaboID > 0 {
			addedCollaboEntries = append(addedCollaboEntries, se)
		}
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
			Source:    e.Source,
			CollaboID: e.CollaboID,
		}
		entries = append(entries, se)
	}

	// コラボの場合はチャンネル主のエントリで他の人のエントリを上書きする
	// 複数チャンネルで行っているコラボの場合はツイートのタイミング次第で
	// エントリの内容が変わる可能性があるが許容する
	for _, se := range addedCollaboEntries {
		for i, target := range entries {
			if target.CollaboID != se.CollaboID {
				continue
			}

			if target.VideoID != "" {
				continue
			}

			temp := se
			temp.ActorName = target.ActorName
			temp.Icon = target.Icon
			entries[i] = temp
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].StartAt.Before(entries[j].StartAt)
	})

	return model.Schedule{
		Date:    targetPlan.Date,
		Entries: entries,
	}
}

func findActorByID(actors []model.Actor, id string) (model.Actor, error) {
	for _, a := range actors {
		if a.ID == id {
			return a, nil
		}
	}

	return model.Actor{}, common.ErrNotFound
}
