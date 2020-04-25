package store

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/common"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"google.golang.org/api/iterator"
)

type planEntrySlice []planEntry

// plan 配信スケジュール
type plan struct {
	// Date 配信日
	Date time.Time `firestore:"date"`
	// Entries 配信予定エントリ
	Entries planEntrySlice `firestore:"entries"`
	// SourceID スケジュールのtweetID
	SourceID string `firestore:"sourceID"`
	// Notified 通知を行ったかどうか
	Notified bool `firestore:"notified"`
}

// planEntry 配信スケジュールのエントリ
type planEntry struct {
	// ActorID 配信者ID
	ActorID string `firestore:"actorID"`
	// StartAt 配信開始時刻
	StartAt time.Time `firestore:"startAt"`
	// Source 配信サイト
	Source string `firestore:"source"`
}

const collectionNamePlan = "Plan"

// FindPlans 開始時刻と終了時刻を指定して計画を検索する
func FindPlans(ctx context.Context, c *firestore.Client, r jst.Range) ([]model.Plan, error) {
	it := c.Collection(collectionNamePlan).Where("date", ">=", r.Begin.Time()).Where("date", "<=", r.End.Time()).Documents(ctx)
	var plans []model.Plan
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var p plan
		doc.DataTo(&p)
		plans = append(plans, p.Plan())
	}

	if plans == nil {
		return nil, common.ErrNotFound
	}

	return plans, nil
}

// FindLatestPlan 最新の計画を取得する
func FindLatestPlan(ctx context.Context, c *firestore.Client) (model.Plan, error) {
	it := c.Collection(collectionNamePlan).OrderBy("date", firestore.Desc).Limit(1).Documents(ctx)
	docs, err := it.GetAll()
	if err != nil {
		return model.Plan{}, err
	}

	if len(docs) == 0 {
		return model.Plan{}, common.ErrNotFound
	}

	var p plan
	docs[0].DataTo(&p)
	return p.Plan(), nil
}

// SavePlan 計画を保存する
// Notifiedを更新する場合はMarkPlanAsNotifiedを使用する
func SavePlan(ctx context.Context, c *firestore.Client, p model.Plan) error {
	temp := fromPlan(p)

	return c.RunTransaction(ctx, func(ctx context.Context, t *firestore.Transaction) error {
		q := c.Collection(collectionNamePlan).Where("date", "==", temp.Date).Limit(1)
		docs, err := t.Documents(q).GetAll()
		if err != nil {
			return err
		}

		// 既に存在する場合はnotifiedフラグだけ引き継いで上書き
		if len(docs) != 0 {
			var oldPlan plan
			docs[0].DataTo(&oldPlan)
			temp.Notified = oldPlan.Notified
			return t.Set(docs[0].Ref, temp)
		}

		return t.Set(c.Collection(collectionNamePlan).NewDoc(), temp)
	})
}

// MarkPlanAsNotified 計画を通知済みとする
// すでに通知済みな場合はなにもしない
// 更新された場合はtrue、されなかった場合はfalse
func MarkPlanAsNotified(ctx context.Context, c *firestore.Client, p model.Plan) (model.Plan, bool, error) {
	updated := false
	var temp model.Plan

	err := c.RunTransaction(ctx, func(ctx context.Context, t *firestore.Transaction) error {
		updated = false
		q := c.Collection(collectionNamePlan).Where("date", "==", p.Date.Time()).Limit(1)
		docs, err := t.Documents(q).GetAll()
		if err != nil {
			return err
		}

		// 保存されていない物は更新できない
		if len(docs) == 0 {
			return nil
		}

		doc := docs[0]
		// 既に存在する場合はnotifiedフラグをチェック
		var oldPlan plan
		doc.DataTo(&oldPlan)
		// 既にNotifiedの場合は何もしない
		if oldPlan.Notified {
			return nil
		}

		updated = true
		oldPlan.Notified = true
		temp = oldPlan.Plan()
		return t.Set(doc.Ref, oldPlan)
	})

	if err != nil {
		return model.Plan{}, false, err
	}

	if !updated {
		return p, false, nil
	}

	return temp, true, nil
}

func fromPlan(p model.Plan) plan {
	var entries planEntrySlice
	for _, e := range p.Entries {
		entries = append(entries, planEntry{
			ActorID: e.ActorID,
			StartAt: e.StartAt.Time(),
			Source:  e.Source,
		})
	}
	return plan{
		Date:     p.Date.Time(),
		Entries:  entries,
		Notified: p.Notified,
		SourceID: p.SourceID,
	}
}

func (p plan) Plan() model.Plan {
	return model.Plan{
		Date:     jst.From(p.Date),
		Entries:  p.Entries.PlanEntries(),
		Notified: p.Notified,
		SourceID: p.SourceID,
	}
}

func (es planEntrySlice) PlanEntries() []model.PlanEntry {
	var res []model.PlanEntry
	for _, e := range es {
		res = append(res, e.PlanEntry())
	}
	return res
}

func (e planEntry) PlanEntry() model.PlanEntry {
	return model.PlanEntry{
		ActorID: e.ActorID,
		StartAt: jst.From(e.StartAt),
		Source:  e.Source,
	}
}
