package service

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/common"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"github.com/yaegaki/dotlive-schedule-server/store"
)

// CreateCalendar カレンダーを作成する
func CreateCalendar(ctx context.Context, client *firestore.Client, baseDate jst.Time, actors model.ActorSlice) (model.Calendar, error) {
	// 次の月の初めの日
	var end jst.Time
	if baseDate.Month() == 12 {
		end = jst.ShortDate(baseDate.Year()+1, 1, 1)
	} else {
		end = jst.ShortDate(baseDate.Year(), baseDate.Month()+1, 1)
	}

	r := jst.Range{
		// スケジュールの作成には前後の情報も必要なので-1/+1する
		Begin: baseDate.AddDay(-1),
		End:   end.AddOneDay(),
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

	for d := baseDate; d.Before(end) && baseDate.Month() == d.Month(); d = d.AddOneDay() {
		s := createScheduleInternal(d, plans, videos, actors)
		actorIDs := []string{}
	OUTER:
		for _, e := range s.Entries {
			a, err := actors.FindActorByName(e.ActorName)
			if err != nil {
				continue
			}

			for _, id := range actorIDs {
				if id == a.ID {
					continue OUTER
				}
			}

			actorIDs = append(actorIDs, a.ID)
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
