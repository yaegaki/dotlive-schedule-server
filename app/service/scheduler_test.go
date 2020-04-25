package service

import (
	"testing"

	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/actor"
	. "github.com/yaegaki/dotlive-schedule-server/internal/testutil/plan"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

func TestCreateScheduleInternal(t *testing.T) {
	tests := []struct {
		name       string
		planRange  jst.Range
		videoRange jst.Range
		schedule   model.Schedule
	}{
		{
			"2020/4/26 at 2020/4/25",
			jst.Range{
				Begin: jst.ShortDate(2020, 4, 24),
				End:   jst.Date(2020, 4, 26, 23, 59),
			},
			jst.Range{
				Begin: jst.ShortDate(2020, 4, 24),
				End:   jst.Date(2020, 4, 25, 23, 59),
			},
			createScheduleForTest(jst.ShortDate(2020, 4, 26), []scheduleEntryPart{
				createScheduleEntryPart(Pino.Name, true, "", 21, 0),
				createScheduleEntryPart(Suzu.Name, true, "", 22, 0),
			}),
		},
		// 計画は存在しているが動画が存在しないときにVideoIDが設定されていないこと
		{
			"2020/4/25 at 2020/4/24",
			jst.Range{
				Begin: jst.ShortDate(2020, 4, 23),
				End:   jst.Date(2020, 4, 25, 23, 59),
			},
			jst.Range{
				Begin: jst.ShortDate(2020, 4, 23),
				End:   jst.Date(2020, 4, 24, 23, 59),
			},
			createScheduleForTest(jst.ShortDate(2020, 4, 25), []scheduleEntryPart{
				createScheduleEntryPart(Suzu.Name, true, "", 12, 0),
			}),
		},
		{
			"2020/4/25",
			jst.Range{
				Begin: jst.ShortDate(2020, 4, 23),
				End:   jst.Date(2020, 4, 25, 23, 59),
			},
			jst.Range{
				Begin: jst.ShortDate(2020, 4, 23),
				End:   jst.Date(2020, 4, 25, 23, 59),
			},
			createScheduleForTest(jst.ShortDate(2020, 4, 25), []scheduleEntryPart{
				createScheduleEntryPart(Suzu.Name, true, "2020-4-25-12-0-suzu", 12, 0),
			}),
		},
		{
			"2020/4/24(bilibili test)",
			jst.Range{
				Begin: jst.ShortDate(2020, 4, 22),
				End:   jst.Date(2020, 4, 24, 23, 59),
			},
			jst.Range{
				Begin: jst.ShortDate(2020, 4, 22),
				End:   jst.Date(2020, 4, 24, 23, 59),
			},
			createScheduleForTest(jst.ShortDate(2020, 4, 24), []scheduleEntryPart{
				createScheduleEntryPartBilibili(Siro.Name, true, "2020-4-24-19-0-siro", 19, 0),
				createScheduleEntryPart(Suzu.Name, true, "2020-4-24-22-0-suzu", 22, 0),
			}),
		},
		{
			"2020/4/23",
			jst.Range{
				Begin: jst.ShortDate(2020, 4, 21),
				End:   jst.Date(2020, 4, 23, 23, 59),
			},
			jst.Range{
				Begin: jst.ShortDate(2020, 4, 21),
				End:   jst.Date(2020, 4, 23, 23, 59),
			},
			createScheduleForTest(jst.ShortDate(2020, 4, 23), []scheduleEntryPart{
				createScheduleEntryPart(Futaba.Name, false, "2020-4-23-4-5-futaba", 4, 5),
				createScheduleEntryPart(Futaba.Name, false, "2020-4-23-13-0-futaba", 13, 0),
				createScheduleEntryPart(Natori.Name, true, "2020-4-23-18-30-natori", 18, 30),
				createScheduleEntryPart(Siro.Name, true, "2020-4-23-20-0-siro", 20, 0),
				createScheduleEntryPart(Pino.Name, true, "2020-4-23-22-0-pino", 22, 0),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := createScheduleInternal(tt.schedule.Date, getPlans(tt.planRange), getVideos(tt.videoRange), All)
			if err != nil {
				t.Errorf("can not create schedule: %v", err)
				return
			}
			compareSchedule(t, s, tt.schedule)
		})
	}
}

func getPlans(r jst.Range) []model.Plan {
	plans := []model.Plan{
		CreatePlan(jst.ShortDate(2020, 4, 23), []EntryPart{
			CreateEntryPart(Natori, 18, 30),
			CreateEntryPart(Siro, 20, 0),
			CreateEntryPart(Pino, 22, 0),
		}),
		CreatePlan(jst.ShortDate(2020, 4, 24), []EntryPart{
			CreateEntryPart(Siro, 19, 0),
			CreateEntryPart(Suzu, 22, 0),
		}),
		CreatePlan(jst.ShortDate(2020, 4, 25), []EntryPart{
			CreateEntryPart(Suzu, 12, 0),
		}),
		CreatePlan(jst.ShortDate(2020, 4, 26), []EntryPart{
			CreateEntryPart(Pino, 21, 0),
			CreateEntryPart(Suzu, 22, 0),
		}),
	}

	var results []model.Plan
	for _, p := range plans {
		if !r.In(p.Date) {
			continue
		}

		results = append(results, p)
	}

	return results
}

func getVideos(r jst.Range) []model.Video {
	videos := []model.Video{
		{
			ID:      "2020-4-23-4-5-futaba",
			ActorID: Futaba.ID,
			Source:  model.VideoSourceYoutube,
			StartAt: jst.Date(2020, 4, 23, 4, 5),
		},
		{
			ID:      "2020-4-23-13-0-futaba",
			ActorID: Futaba.ID,
			Source:  model.VideoSourceYoutube,
			StartAt: jst.Date(2020, 4, 23, 13, 0),
		},
		{
			ID:      "2020-4-23-18-30-natori",
			ActorID: Natori.ID,
			Source:  model.VideoSourceYoutube,
			StartAt: jst.Date(2020, 4, 23, 18, 30),
		},
		{
			ID:      "2020-4-23-20-0-siro",
			ActorID: Siro.ID,
			Source:  model.VideoSourceYoutube,
			StartAt: jst.Date(2020, 4, 23, 19, 58),
		},
		{
			ID:      "2020-4-23-22-0-pino",
			ActorID: Pino.ID,
			Source:  model.VideoSourceYoutube,
			StartAt: jst.Date(2020, 4, 23, 22, 00),
		},
		{
			ID:      "2020-4-24-19-0-siro",
			ActorID: Siro.ID,
			Source:  model.VideoSourceBilibili,
			// Bilibiliなのでツイート時間が開始時刻になっている
			StartAt: jst.Date(2020, 4, 24, 14, 1),
		},
		{
			ID:      "2020-4-24-22-0-suzu",
			ActorID: Suzu.ID,
			Source:  model.VideoSourceYoutube,
			StartAt: jst.Date(2020, 4, 24, 22, 0),
		},
		{
			ID:      "2020-4-25-12-0-suzu",
			ActorID: Suzu.ID,
			Source:  model.VideoSourceYoutube,
			StartAt: jst.Date(2020, 4, 25, 12, 0),
		},
	}

	var results []model.Video
	for _, v := range videos {
		if !r.In(v.StartAt) {
			continue
		}

		results = append(results, v)
	}

	return results
}

func compareSchedule(t *testing.T, got model.Schedule, expect model.Schedule) {
	if !got.Date.Equal(expect.Date) {
		t.Errorf("date, got: %v expect: %v", got.Date, expect.Date)
	}

	if len(got.Entries) != len(expect.Entries) {
		t.Errorf("len(entries), got: %v expect: %v", len(got.Entries), len(expect.Entries))
		return
	}

	for i, e := range got.Entries {
		expectEntry := expect.Entries[i]
		if e.ActorName != expectEntry.ActorName {
			t.Errorf("ActorName, got: %v expect: %v", e.ActorName, expectEntry.ActorName)
			continue
		}

		if e.Planned != expectEntry.Planned {
			t.Errorf("Planned, got: %v expect: %v", e.Planned, expectEntry.Planned)
		}

		if !e.StartAt.Equal(expectEntry.StartAt) {
			t.Errorf("StartAt, got: %v expect: %v", e.StartAt, expectEntry.StartAt)
		}

		if e.VideoID != expectEntry.VideoID {
			t.Errorf("VideoID, got: %v expect: %v", e.VideoID, expectEntry.VideoID)
		}

		if e.Source != expectEntry.Source {
			t.Errorf("Source, got: %v expect: %v", e.Source, expectEntry.Source)
		}
	}
}

type scheduleEntryPart struct {
	name    string
	planned bool
	videoID string
	source  string
	hour    int
	min     int
}

func createScheduleEntryPart(name string, planned bool, videoID string, hour, min int) scheduleEntryPart {
	return scheduleEntryPart{
		name:    name,
		planned: planned,
		videoID: videoID,
		source:  model.VideoSourceYoutube,
		hour:    hour,
		min:     min,
	}
}

func createScheduleEntryPartBilibili(name string, planned bool, videoID string, hour, min int) scheduleEntryPart {
	return scheduleEntryPart{
		name:    name,
		planned: planned,
		videoID: videoID,
		source:  model.VideoSourceBilibili,
		hour:    hour,
		min:     min,
	}
}

func createScheduleForTest(d jst.Time, parts []scheduleEntryPart) model.Schedule {
	var entries []model.ScheduleEntry
	for _, p := range parts {
		entries = append(entries, model.ScheduleEntry{
			ActorName: p.name,
			Planned:   p.planned,
			StartAt:   jst.Date(d.Year(), d.Month(), d.Day(), p.hour, p.min),
			VideoID:   p.videoID,
			Source:    p.source,
		})
	}

	return model.Schedule{
		Date:    d,
		Entries: entries,
	}
}
