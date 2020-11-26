package store

import (
	"context"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/common"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"golang.org/x/xerrors"
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
	// Fixed 固定化されているかどうか
	Fixed bool `firestore:"fixed"`
	// Text 計画ツイートの内容部分
	// 計画を通知するときに使用する
	Text string `firestore:"text"`
}

// planEntry 配信スケジュールのエントリ
type planEntry struct {
	// ActorID 配信者ID
	ActorID string `firestore:"actorID"`
	// HashTag コラボハッシュタグ
	HashTag string `firestore:"hashTag"`
	// StartAt 配信開始時刻
	StartAt time.Time `firestore:"startAt"`
	// Source 配信サイト
	Source string `firestore:"source"`
	// MemberOnly メンバー限定かどうか
	MemberOnly bool `firestore:"memberOnly"`
	// CollaboID コラボID
	CollaboID int `firestore:"collaboID"`
}

const collectionNamePlan = "Plan"

// ErrFixedPlan Fixedされた計画をSavePlanWithExplicitID以外で保存しようとしたときに発生する
var ErrFixedPlan = xerrors.Errorf("Plan is fixed")

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
	additional := p.Additional
	fixed := false

	err := c.RunTransaction(ctx, func(ctx context.Context, t *firestore.Transaction) error {
		q := c.Collection(collectionNamePlan).Where("date", "==", temp.Date).Limit(1)
		docs, err := t.Documents(q).GetAll()
		if err != nil {
			return err
		}

		// 既に存在する場合はnotifiedフラグだけ引き継いで上書き
		if len(docs) != 0 {
			var oldPlan plan
			docs[0].DataTo(&oldPlan)
			// fixされている場合は保存しない
			if oldPlan.Fixed {
				fixed = true
				return nil
			}
			temp.Notified = oldPlan.Notified
			if additional {
				temp = oldPlan.Merge(temp)
			}
			return t.Set(docs[0].Ref, temp)
		}

		return t.Set(c.Collection(collectionNamePlan).NewDoc(), temp)
	})

	if err != nil {
		return err
	}

	if fixed {
		return ErrFixedPlan
	}

	return nil
}

// SavePlanWithExplicitID 指定したIDで保存する
func SavePlanWithExplicitID(ctx context.Context, c *firestore.Client, p model.Plan, id string) error {
	temp := fromPlan(p)

	return c.RunTransaction(ctx, func(ctx context.Context, t *firestore.Transaction) error {
		_, err := c.Collection(collectionNamePlan).Doc(id).Set(ctx, temp)
		return err
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
			ActorID:    e.ActorID,
			HashTag:    e.HashTag,
			StartAt:    e.StartAt.Time(),
			Source:     e.Source,
			MemberOnly: e.MemberOnly,
			CollaboID:  e.CollaboID,
		})
	}
	return plan{
		Date:     p.Date.Time(),
		Entries:  entries,
		Notified: p.Notified,
		SourceID: p.SourceID,
		Fixed:    p.Fixed,
		Text:     p.Text,
	}
}

func (p plan) Plan() model.Plan {
	return model.Plan{
		Date:     jst.From(p.Date),
		Entries:  p.Entries.PlanEntries(),
		Notified: p.Notified,
		SourceID: p.SourceID,
		Fixed:    p.Fixed,
		Text:     p.Text,
	}
}

func (p plan) Merge(other plan) plan {
	if len(other.Entries) == 0 {
		return p
	}

	// other.Textがp.Textに完全に含まれている場合は既に追加されている
	if strings.Index(p.Text, other.Text) >= 0 {
		return p
	}

	newPlan := p
	baseCollaboID := 0
	for _, e := range p.Entries {
		if e.CollaboID > baseCollaboID {
			baseCollaboID = e.CollaboID
		}
	}

OUTER:
	for _, e := range other.Entries {
		for _, existsEntries := range p.Entries {
			if e.ActorID == existsEntries.ActorID && e.StartAt.Equal(existsEntries.StartAt) {
				continue OUTER
			}
		}

		// 追加のコラボIDと既存のコラボIDが被らないようにする
		if e.CollaboID > 0 {
			e.CollaboID = e.CollaboID + baseCollaboID
		}

		newPlan.Entries = append(newPlan.Entries, e)
	}

	sort.Slice(newPlan.Entries, func(i, j int) bool {
		l := newPlan.Entries[i]
		r := newPlan.Entries[j]
		if l.StartAt.Equal(r.StartAt) {
			return false
		}

		// 開始時間でソート
		return l.StartAt.Before(r.StartAt)
	})

	if len(p.Entries) == 0 {
		newPlan.Text = other.Text
	} else if p.Entries[0].StartAt.Before(other.Entries[0].StartAt) {
		newPlan.Text = p.Text + "\n" + other.Text
	} else {
		newPlan.Text = other.Text + "\n" + p.Text
	}

	return newPlan
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
		ActorID:    e.ActorID,
		HashTag:    e.HashTag,
		StartAt:    jst.From(e.StartAt),
		Source:     e.Source,
		MemberOnly: e.MemberOnly,
		CollaboID:  e.CollaboID,
	}
}
