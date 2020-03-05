package model

import (
	"time"

	"github.com/yaegaki/dotlive-schedule-server/common"
	"github.com/yaegaki/dotlive-schedule-server/jst"
)

// Plan 計画のツイート
type Plan struct {
	// Date 計画の日付
	Date jst.Time
	// Source ソースとなるツイートID
	SourceID string
	// Entries 計画のエントリ
	Entries []PlanEntry
	// Notified 通知済みか
	Notified bool
}

// PlanEntry 計画のエントリ
type PlanEntry struct {
	// ActorID 配信者ID
	ActorID string
	// StartAt 開始時間
	StartAt jst.Time
}

// IsPlanned 計画配信かどうか
func (p Plan) IsPlanned(v Video) bool {
	_, err := p.GetEntry(v)
	return err == nil
}

// GetEntry 指定された動画が計画されたものならそのエントリを取得する
func (p Plan) GetEntry(v Video) (PlanEntry, error) {
	for _, e := range p.Entries {
		if e.ActorID != v.ActorID {
			continue
		}

		planRange := jst.Range{
			End: e.StartAt.Add(30 * time.Minute),
		}

		if v.Source == VideoSourceYoutube {
			planRange.Begin = e.StartAt.Add(-30 * time.Minute)
		} else {
			// Bilibiliは開始時間があいまいにしか取れないので-90分までは計画配信とする
			planRange.Begin = e.StartAt.Add(-90 * time.Minute)
		}

		if !planRange.In(v.StartAt) {
			continue
		}

		return e, nil
	}

	return PlanEntry{}, common.ErrNotFound
}
