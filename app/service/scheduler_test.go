package service

import (
	"sort"
	"strings"
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
		{
			"2020/4/19",
			jst.Range{
				Begin: jst.ShortDate(2020, 4, 18),
				End:   jst.Date(2020, 4, 19, 23, 59),
			},
			jst.Range{
				Begin: jst.ShortDate(2020, 4, 18),
				End:   jst.Date(2020, 4, 19, 23, 59),
			},
			createScheduleForTest(jst.ShortDate(2020, 4, 19), []scheduleEntryPart{
				createScheduleEntryPartCollabo(Chieri.Name, true, "2020-4-19-20-0-collabo", 20, 0, 1),
				createScheduleEntryPartCollabo(Pino.Name, true, "2020-4-19-20-0-collabo", 20, 0, 1),
				createScheduleEntryPartCollabo(Iroha.Name, true, "2020-4-19-20-0-collabo", 20, 0, 1),
				createScheduleEntryPartCollabo(Mememe.Name, true, "2020-4-19-20-0-collabo", 20, 0, 1),
				createScheduleEntryPart(Chieri.Name, false, "2020-4-19-20-39-collabo", 20, 39),
				// 30分ずれたらゲリラ扱いになるのでコラボ認識されない
				// createScheduleEntryPart(Pino.Name, false, "2020-4-19-20-39-collabo", 20, 39),
				// createScheduleEntryPart(Iroha.Name, false, "2020-4-19-20-39-collabo", 20, 39),
				// createScheduleEntryPart(Mememe.Name, false, "2020-4-19-20-39-collabo", 20, 39),
				createScheduleEntryPart(Suzu.Name, true, "2020-4-19-22-0-suzu", 22, 0),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := createScheduleInternal(tt.schedule.Date, getPlans(tt.planRange), getVideos(tt.videoRange), All)
			compareSchedule(t, s, tt.schedule)
		})
	}

	// 特殊ケースのテスト

	t.Run("over 24h", func(t *testing.T) {
		d := jst.ShortDate(2020, 6, 6)
		p := CreatePlan(d, []EntryPart{
			CreateEntryPart(Suzu, 21, 0),
			CreateEntryPartMildom(Chieri, 22, 0),
		})
		vs := []model.Video{
			{
				ID:      "suzu",
				ActorID: Suzu.ID,
				Source:  model.VideoSourceYoutube,
				StartAt: jst.Date(2020, 6, 6, 21, 0),
			},
			{
				ID:      "chieri",
				ActorID: Chieri.ID,
				Source:  model.VideoSourceMildom,
				StartAt: jst.Date(2020, 6, 6, 22, 00),
			},
			{
				ID:      "natori",
				ActorID: Natori.ID,
				Source:  model.VideoSourceYoutube,
				StartAt: jst.Date(2020, 6, 7, 0, 5),
			},
			{
				ID:      "mememe",
				ActorID: Mememe.ID,
				Source:  model.VideoSourceYoutube,
				StartAt: jst.Date(2020, 6, 7, 0, 15),
			},
		}
		s := createScheduleInternal(d, []model.Plan{p}, vs, All)
		compareSchedule(t, s, createScheduleForTest(jst.ShortDate(2020, 6, 6), []scheduleEntryPart{
			createScheduleEntryPart(Suzu.Name, true, "suzu", 21, 0),
			createScheduleEntryPartMildom(Chieri.Name, true, "chieri", 22, 0),
		}))

		p = CreatePlan(d, []EntryPart{
			CreateEntryPart(Suzu, 21, 0),
			CreateEntryPartMildom(Chieri, 22, 0),
			CreateEntryPart(Mememe, 24, 15),
		})
		s = createScheduleInternal(d, []model.Plan{p}, vs, All)
		compareSchedule(t, s, createScheduleForTest(jst.ShortDate(2020, 6, 6), []scheduleEntryPart{
			createScheduleEntryPart(Suzu.Name, true, "suzu", 21, 0),
			createScheduleEntryPartMildom(Chieri.Name, true, "chieri", 22, 0),
			createScheduleEntryPart(Natori.Name, false, "natori", 24, 5),
			createScheduleEntryPart(Mememe.Name, true, "mememe", 24, 15),
		}))
	})

	// コラボ関連
	t.Run("Collabo", func(t *testing.T) {
		d := jst.ShortDate(2020, 4, 29)
		p := CreatePlan(d, []EntryPart{
			CreateEntryPartCollabo(Iori, 20, 0, 1),
			CreateEntryPartCollabo(Suzu, 20, 0, 1),
			CreateEntryPartCollabo(Pino, 22, 0, 2),
			CreateEntryPartCollabo(Iroha, 22, 0, 2),
		})
		vs := []model.Video{
			{
				ID:      "iosu-1",
				ActorID: Iori.ID,
				Source:  model.VideoSourceYoutube,
				StartAt: jst.Date(2020, 4, 29, 20, 0),
			},
			{
				ID:      "iosu-2",
				ActorID: Iori.ID,
				Source:  model.VideoSourceYoutube,
				StartAt: jst.Date(2020, 4, 29, 20, 15),
			},
			{
				ID:      "pinogon",
				ActorID: Pino.ID,
				Source:  model.VideoSourceYoutube,
				StartAt: jst.Date(2020, 4, 29, 22, 0),
			},
		}
		s := createScheduleInternal(d, []model.Plan{p}, vs, All)
		// 枠取り直した場合はチャンネル主のエントリだけ作られている
		// 2つ以上のコラボがあったときに正しく処理されている
		compareSchedule(t, s, createScheduleForTest(jst.ShortDate(2020, 4, 29), []scheduleEntryPart{
			createScheduleEntryPartCollabo(Iori.Name, true, "iosu-1", 20, 0, 1),
			createScheduleEntryPartCollabo(Suzu.Name, true, "iosu-1", 20, 0, 1),
			createScheduleEntryPartCollabo(Iori.Name, true, "iosu-2", 20, 15, 1),
			createScheduleEntryPartCollabo(Pino.Name, true, "pinogon", 22, 0, 2),
			createScheduleEntryPartCollabo(Iroha.Name, true, "pinogon", 22, 0, 2),
		}))
	})
}

func getPlans(r jst.Range) []model.Plan {
	plans := []model.Plan{
		CreatePlan(jst.ShortDate(2020, 4, 19), []EntryPart{
			CreateEntryPartCollabo(Chieri, 20, 0, 1),
			CreateEntryPartCollabo(Pino, 20, 0, 1),
			CreateEntryPartCollabo(Iroha, 20, 0, 1),
			CreateEntryPartCollabo(Mememe, 20, 0, 1),
			CreateEntryPart(Suzu, 22, 0),
		}),
		CreatePlan(jst.ShortDate(2020, 4, 23), []EntryPart{
			CreateEntryPart(Natori, 18, 30),
			CreateEntryPart(Siro, 20, 0),
			CreateEntryPart(Pino, 22, 0),
		}),
		CreatePlan(jst.ShortDate(2020, 4, 24), []EntryPart{
			CreateEntryPartBilibili(Siro, 19, 0),
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
			ID:      "2020-4-19-20-0-collabo",
			ActorID: Chieri.ID,
			Source:  model.VideoSourceYoutube,
			StartAt: jst.Date(2020, 4, 19, 20, 0),
		},
		{
			ID:      "2020-4-19-20-39-collabo",
			ActorID: Chieri.ID,
			Source:  model.VideoSourceYoutube,
			StartAt: jst.Date(2020, 4, 19, 20, 39),
		},
		{
			ID:      "2020-4-19-22-0-suzu",
			ActorID: Suzu.ID,
			Source:  model.VideoSourceYoutube,
			StartAt: jst.Date(2020, 4, 19, 22, 0),
		},
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

	createSortedEntries := func(entries []model.ScheduleEntry) []model.ScheduleEntry {
		c := append([]model.ScheduleEntry{}, entries...)
		sort.Slice(c, func(i, j int) bool {
			if c[i].StartAt.Equal(c[j].StartAt) {
				// 開始時間が同じ場合は名前順
				return strings.Compare(c[i].ActorName, c[j].ActorName) < 0
			}

			// 開始時間でソート
			return c[i].StartAt.Before(c[j].StartAt)
		})
		return c
	}

	expectEntries := createSortedEntries(expect.Entries)

	for i, e := range createSortedEntries(got.Entries) {
		expectEntry := expectEntries[i]
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

		if e.CollaboID != expectEntry.CollaboID {
			t.Errorf("CollaboID, got: %v expect: %v", e.CollaboID, expectEntry.CollaboID)
		}
	}
}

type scheduleEntryPart struct {
	name      string
	planned   bool
	videoID   string
	source    string
	hour      int
	min       int
	collaboID int
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

func createScheduleEntryPartCollabo(name string, planned bool, videoID string, hour, min int, collaboID int) scheduleEntryPart {
	return scheduleEntryPart{
		name:      name,
		planned:   planned,
		videoID:   videoID,
		source:    model.VideoSourceYoutube,
		hour:      hour,
		min:       min,
		collaboID: collaboID,
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

func createScheduleEntryPartMildom(name string, planned bool, videoID string, hour, min int) scheduleEntryPart {
	return scheduleEntryPart{
		name:    name,
		planned: planned,
		videoID: videoID,
		source:  model.VideoSourceMildom,
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
			CollaboID: p.collaboID,
		})
	}

	return model.Schedule{
		Date:    d,
		Entries: entries,
	}
}
