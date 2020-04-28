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
	// Source 配信サイト
	Source string
	// CollaboID コラボの場合に識別するためのID
	//           1以上の場合が有効な値
	CollaboID int
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

		var planRange jst.Range
		if v.Source == VideoSourceBilibili {
			// bilibiliは開始時刻が正確にとれないので同じ日であれば計画通りとする
			// 1日1回という前提
			d := e.StartAt.FloorToDay()
			planRange = jst.Range{
				Begin: d,
				End:   d.AddOneDay(),
			}
		} else {
			// 計画から+-30分以内なら計画通りとする
			planRange = jst.Range{
				Begin: e.StartAt.Add(-30 * time.Minute),
				End:   e.StartAt.Add(30 * time.Minute),
			}
		}

		if !planRange.In(v.StartAt) {
			continue
		}

		return e, nil
	}

	return PlanEntry{}, common.ErrNotFound
}
