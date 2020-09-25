package model

import (
	"strings"
	"time"

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
	// Fixed 固定化されているか
	// 固定化されている場合は定期ジョブによって更新されない
	Fixed bool
	// Text 計画ツイートの内容部分
	// 計画を通知するときに使用する
	Text string
	// Additional 追加の計画かどうか
	// 追加の計画の場合は既存の計画を上書きせずに追加する
	Additional bool
}

// PlanEntry 計画のエントリ
type PlanEntry struct {
	// ActorID 配信者ID
	ActorID string
	// HashTag コラボハッシュタグ
	// 通常の配信の場合は空文字
	// このフィールドが空文字ではない場合は必ずActorID == UnknownActorIDになる
	HashTag string
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
	return p.GetEntryIndex(v) >= 0
}

// GetEntryIndex 指定された動画が計画されたものならエントリーのインデックスを取得する
// 計画されていない場合は-1を返す
func (p Plan) GetEntryIndex(v Video) int {
	videoIsUnknownActor := v.IsUnknownActor()

	for i, e := range p.Entries {
		if !e.within(v.Source, v.StartAt) {
			continue
		}

		entryIsUnknownActor := e.IsUnknownActor()
		// 計画がUnknownの場合はハッシュタグで判定する
		if entryIsUnknownActor {
			for _, hashtag := range v.HashTags {
				if strings.Index(e.HashTag, hashtag) >= 0 {
					return i
				}
			}

			continue
		}

		var videoActorID string
		// 動画がUnknownの場合はRelatedActorIDを配信者IDとして判定する
		if videoIsUnknownActor {
			videoActorID = v.RelatedActorID
		} else {
			videoActorID = v.ActorID
		}

		if videoActorID != e.ActorID {
			continue
		}

		if v.Source != e.Source {
			continue
		}

		return i
	}

	return -1
}

// IsUnknownActor 配信者不明かどうか
func (e PlanEntry) IsUnknownActor() bool {
	return e.ActorID == ActorIDUnknown
}

func (e PlanEntry) within(videoSource string, t jst.Time) bool {
	var planRange jst.Range
	if videoSource == VideoSourceYoutube {
		// 計画から+-30分以内なら計画通りとする
		planRange = jst.Range{
			Begin: e.StartAt.Add(-30 * time.Minute),
			End:   e.StartAt.Add(30 * time.Minute),
		}
	} else {
		// Youtube以外は開始時刻が正確にとれないので計画の時間から-26h~+30minまでは計画通りとする
		// 1日1回、2日連続はないという前提
		beginDate := e.StartAt.Add(-26 * time.Hour)
		endDate := e.StartAt.Add(30 * time.Minute)
		planRange = jst.Range{
			Begin: beginDate,
			End:   endDate,
		}
	}

	return planRange.In(t)
}
