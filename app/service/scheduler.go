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
	date = date.FloorToDay()

	// スケジュールを作成するためには前日の情報、翌日の12時までの情報が必要
	r := jst.Range{
		Begin: date.AddDay(-1),
		End:   date.AddOneDay().Add(time.Hour * 12),
	}

	plans, err := store.FindPlans(ctx, c, r)
	if err != nil && err != common.ErrNotFound {
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

type multiPlan struct {
	prev model.Plan
	cur  model.Plan
	next model.Plan
}

func (p multiPlan) isPrevPlanned(v model.Video) bool {
	return p.prev.IsPlanned(v)
}

func (p multiPlan) GetEntryIndex(v model.Video) int {
	return p.cur.GetEntryIndex(v)
}

func (p multiPlan) isNextPlanned(v model.Video) bool {
	return p.next.IsPlanned(v)
}

func createMultiPlan(date jst.Time, plans []model.Plan) multiPlan {
	result := multiPlan{
		prev: model.Plan{
			Date: jst.ShortDate(date.Year(), date.Month(), date.Day()-1),
		},
		cur: model.Plan{
			Date: date,
		},
		next: model.Plan{
			Date: jst.ShortDate(date.Year(), date.Month(), date.Day()+1),
		},
	}

	for _, p := range plans {
		if p.Date.Before(date) {
			result.prev = p
		} else if p.Date.Equal(date) {
			result.cur = p
		} else {
			result.next = p
			break
		}
	}

	return result
}

func createScheduleInternal(date jst.Time, plans []model.Plan, videos []model.Video, actors []model.Actor) model.Schedule {
	plan := createMultiPlan(date, plans)

	scheduleRange := jst.Range{
		Begin: plan.cur.Date,
		End:   plan.cur.Date.AddOneDay().Add(-1 * time.Second),
	}

	for _, e := range plan.cur.Entries {
		// 25時などが指定されている場合はスケジュールの終了時間を延ばす
		if e.StartAt.After(scheduleRange.End) {
			scheduleRange.End = e.StartAt
		}
	}

	entries := []model.ScheduleEntry{}
	var addedPlanEntries []int

	const dotliveIcon = "https://pbs.twimg.com/profile_images/953977243251822593/tglswtot.jpg"
	addScheduleEntry := func(index int, v model.Video) {
		var actorName string
		var icon string
		var startAt jst.Time
		var collaboID int
		var isPlanned bool

		if index < 0 {
			if v.IsUnknownActor() {
				actorName = v.OwnerName
				icon = dotliveIcon
			} else {
				actor, err := findActorByID(actors, v.ActorID)
				if err != nil {
					log.Printf("Unknown actorID: %v, videoID: %v", v.ActorID, v.ID)
					return
				}
				actorName = actor.Name
				icon = actor.Icon
			}
			startAt = v.StartAt
			collaboID = 0
			isPlanned = false

			// シロちゃんの動画は常に計画されているとする
			const siroID = "lLhToxu1Kyxuwwygh0FK"
			if v.ActorID == siroID && !v.IsLive {
				isPlanned = true
			}
		} else {
			pe := plan.cur.Entries[index]

			alreadyAddedEntry := false
			for _, added := range addedPlanEntries {
				if index == added {
					alreadyAddedEntry = true
				}
			}

			if alreadyAddedEntry {
				startAt = v.StartAt
			} else {
				// 初めて追加する場合は開始時刻を計画の時間に合わせる
				startAt = pe.StartAt
				addedPlanEntries = append(addedPlanEntries, index)
			}

			if pe.IsUnknownActor() {
				actorName = pe.HashTag
				icon = dotliveIcon
			} else if v.IsUnknownActor() {
				relatedActor, err := findActorByID(actors, v.RelatedActorID)
				if err != nil {
					log.Printf("Unknown relatedActorID: %v", v.RelatedActorID)
					return
				}

				actorName = relatedActor.Name + " x " + v.OwnerName
				icon = relatedActor.Icon
			} else {
				actor, err := findActorByID(actors, v.ActorID)
				if err != nil {
					log.Printf("Unknown actorID: %v, videoID: %v", v.ActorID, v.ID)
					return
				}
				actorName = actor.Name
				icon = actor.Icon
			}
			collaboID = pe.CollaboID
			isPlanned = true
		}

		se := model.ScheduleEntry{
			ActorName:  actorName,
			Note:       createNote(isPlanned, v.MemberOnly, v.Source),
			Icon:       icon,
			StartAt:    startAt,
			Planned:    isPlanned,
			IsLive:     v.IsLive,
			Text:       v.Text,
			URL:        v.URL,
			VideoID:    v.ID,
			Source:     v.Source,
			MemberOnly: v.MemberOnly,
			CollaboID:  collaboID,
		}
		entries = append(entries, se)
	}

	// 開始順にソートする
	videos = append([]model.Video{}, videos...)
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].StartAt.Before(videos[j].StartAt)
	})

	for _, v := range videos {
		// 前日に予定された配信かどうか
		if plan.isPrevPlanned(v) {
			continue
		}

		peIndex := plan.GetEntryIndex(v)
		if peIndex < 0 {
			// 今日のものか
			if !scheduleRange.In(v.StartAt) {
				continue
			}

			// 明日に予定されているか
			if plan.isNextPlanned(v) {
				continue
			}

			// Youtube以外のゲリラ配信はない
			if v.Source != model.VideoSourceYoutube {
				continue
			}
		}

		addScheduleEntry(peIndex, v)
	}

LOOP_ENTRIES:
	for i, e := range plan.cur.Entries {
		for _, added := range addedPlanEntries {
			if i == added {
				continue LOOP_ENTRIES
			}
		}

		var actorName string
		var icon string
		if e.IsUnknownActor() {
			actorName = e.HashTag
			icon = dotliveIcon
		} else {
			actor, err := findActorByID(actors, e.ActorID)
			if err != nil {
				log.Printf("Unknown actorID: %v", e.ActorID)
				continue
			}

			actorName = actor.Name
			icon = actor.Icon
		}

		se := model.ScheduleEntry{
			ActorName: actorName,
			Icon:      icon,
			StartAt:   e.StartAt,
			Planned:   true,
			Source:    e.Source,
			Note:      createNote(true, e.MemberOnly, e.Source),
			CollaboID: e.CollaboID,
		}

		entries = append(entries, se)
	}

	// コラボの場合はチャンネル主のエントリで他の人のエントリを上書きする
	// 複数チャンネルで行っているコラボの場合はツイートのタイミング次第で
	// エントリの内容が変わる可能性があるが許容する
	for _, collaboEntry := range entries {
		if collaboEntry.CollaboID <= 0 || collaboEntry.VideoID == "" {
			continue
		}

		for i, target := range entries {
			if target.CollaboID != collaboEntry.CollaboID {
				continue
			}

			if target.VideoID != "" {
				continue
			}

			temp := collaboEntry
			temp.ActorName = target.ActorName
			temp.Icon = target.Icon
			entries[i] = temp
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].StartAt.Before(entries[j].StartAt)
	})

	return model.Schedule{
		Date:    plan.cur.Date,
		TweetID: plan.cur.SourceID,
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

func createNote(isPlanned bool, memberOnly bool, source string) string {
	if source == model.VideoSourceYoutube {
		if memberOnly {
			return " (メン限)"
		}

		if isPlanned {
			return ""
		}

		// ゲリラ配信の表示も一応できるようにしておく
		// 現状は何も表示しない
		return ""
	}

	return " (" + source + ")"
}
