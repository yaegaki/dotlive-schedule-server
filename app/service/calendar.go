package service

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/common"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/store"
)

// CreateCalendar カレンダーを作成する
func CreateCalendar(ctx context.Context, client *firestore.Client, baseDate jst.Time, now jst.Time, actors model.ActorSlice) (model.Calendar, error) {
	// 次の月の初めの日
	var end jst.Time
	if baseDate.Month() == 12 {
		end = jst.ShortDate(baseDate.Year()+1, 1, 1)
	} else {
		end = jst.ShortDate(baseDate.Year(), baseDate.Month()+1, 1)
	}

	// スケジュールを作成するためには開始日の前日の情報、終了日の翌日の12時までの情報が必要
	r := jst.Range{
		Begin: baseDate.AddDay(-1),
		End:   end.AddOneDay().Add(time.Hour * 12),
	}

	calendar := model.Calendar{
		BaseDate: baseDate,
		Days:     []model.CalendarDay{},
	}

	plans, err := store.FindPlans(ctx, client, r)
	if err != nil && err != common.ErrNotFound {
		return model.Calendar{}, err
	}

	videos, err := store.FindVideos(ctx, client, r)
	if err != nil && err != common.ErrNotFound {
		return model.Calendar{}, err
	}
	videoMap := map[string]model.Video{}
	for _, v := range videos {
		videoMap[v.ID] = v
	}

	// 2日以上前は確実にFixされている
	fixedDayLimit := now.AddDay(-2)
	monthStart := jst.ShortDate(baseDate.Year(), baseDate.Month(), 1)
	if monthStart.Before(fixedDayLimit) && fixedDayLimit.Before(baseDate) {
		calendar.FixedDay = fixedDayLimit.Day()
	}

	for d := baseDate; baseDate.Month() == d.Month(); d = d.AddOneDay() {
		if d.Before(fixedDayLimit) {
			calendar.FixedDay = d.Day()
		}

		s := createScheduleInternal(d, plans, videos, actors)
		actorIDs := []string{}
		for _, e := range s.Entries {
			relatedActors := findActorsByScheduleEntry(e, videoMap, actors)

		OUTER:
			for _, relatedActor := range relatedActors {
				for _, id := range actorIDs {
					if relatedActor.ID == id {
						continue OUTER
					}
				}

				actorIDs = append(actorIDs, relatedActor.ID)
			}
		}

		if len(actorIDs) == 0 {
			continue
		}

		calendar.Days = append(calendar.Days, model.CalendarDay{
			Day:      d.Day(),
			ActorIDs: actorIDs,
		})
	}

	return calendar, nil
}

func findActorsByScheduleEntry(se model.ScheduleEntry, videoMap map[string]model.Video, actors model.ActorSlice) model.ActorSlice {
	if se.VideoID != "" {
		return findActorsByVideoID(se.VideoID, videoMap, actors)
	}

	actor, err := actors.FindActorByName(se.ActorName)
	if err != nil {
		return nil
	}

	return model.ActorSlice{actor}
}

func findActorsByVideoID(videoID string, videoMap map[string]model.Video, actors model.ActorSlice) model.ActorSlice {
	var result model.ActorSlice
	v, ok := videoMap[videoID]
	if !ok {
		return result
	}

	if v.IsUnknownActor() {
		actor, err := actors.FindActor(v.RelatedActorID)
		if err == nil {
			result = append(result, actor)
		} else {
			log.Printf("Unknown actor: %v", v.ActorID)
		}

		// 関連する配信者が二人以上いる場合
		// RelatedActorIDsにはRelatedActorIDが含まれる場合とそうでない場合がある
	RELATED_ACTORID_LOOP:
		for _, relatedActorID := range v.RelatedActorIDs {
			for _, temp := range result {
				if temp.ID == relatedActorID {
					continue RELATED_ACTORID_LOOP
				}
			}

			actor, err = actors.FindActor(relatedActorID)
			if err != nil {
				log.Printf("Unknown actor: %v", relatedActorID)
				continue
			}
			result = append(result, actor)
		}
	} else {
		actor, err := actors.FindActor(v.ActorID)
		if err != nil {
			log.Printf("Unknown actor: %v", v.ActorID)
		}
		result = append(result, actor)
	}

	return result
}
