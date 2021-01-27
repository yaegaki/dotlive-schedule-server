package store

import (
	"context"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/yaegaki/dotlive-schedule-server/common"
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
	"golang.org/x/xerrors"
	"google.golang.org/api/iterator"
)

type planEntrySlice []planEntry
type planTextSlice []planText

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
	// Texts 計画ツイートの内容部分
	// 計画を通知するときに使用する
	Texts planTextSlice `firestore:"text"`
}

type planText struct {
	// Date テキストの始めの時間
	Date time.Time `firestore:"date"`
	// PlanTag 計画が分割されてるときの識別タグ
	PlanTag string `firestore:"planTag"`
	// Text 計画ツイートの内容部分
	// 計画を通知するときに使用する
	Text string `firestore:"text"`
}

// planEntry 配信スケジュールのエントリ
type planEntry struct {
	// ActorID 配信者ID
	ActorID string `firestore:"actorID"`
	// PlanTag 計画が分割されてるときの識別タグ
	PlanTag string `firestore:"planTag"`
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
	planTag := p.PlanTag
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
			temp = oldPlan.Merge(temp, planTag)
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
			PlanTag:    e.PlanTag,
			HashTag:    e.HashTag,
			StartAt:    e.StartAt.Time(),
			Source:     e.Source,
			MemberOnly: e.MemberOnly,
			CollaboID:  e.CollaboID,
		})
	}
	var texts planTextSlice
	for _, t := range p.Texts {
		texts = append(texts, planText{
			Date:    t.Date.Time(),
			PlanTag: t.PlanTag,
			Text:    t.Text,
		})
	}
	return plan{
		Date:     p.Date.Time(),
		Entries:  entries,
		Notified: p.Notified,
		SourceID: p.SourceID,
		Fixed:    p.Fixed,
		Texts:    texts,
	}
}

func (p plan) Plan() model.Plan {
	return model.Plan{
		Date:     jst.From(p.Date),
		Entries:  p.Entries.PlanEntries(),
		Notified: p.Notified,
		SourceID: p.SourceID,
		Fixed:    p.Fixed,
		Texts:    p.Texts.PlanTexts(),
	}
}

func (p plan) Merge(other plan, planTag string) plan {
	newPlan := p.removeByPlanTag(planTag)
	if len(other.Entries) == 0 {
		return newPlan
	}

	baseCollaboID := 0
	for _, e := range newPlan.Entries {
		if e.CollaboID > baseCollaboID {
			baseCollaboID = e.CollaboID
		}
	}

OUTER:
	for _, e := range other.Entries {
		for _, existsEntries := range newPlan.Entries {
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

	for _, t := range other.Texts {
		newPlan.Texts = append(newPlan.Texts, t)
	}

	sort.Slice(newPlan.Texts, func(i, j int) bool {
		l := newPlan.Texts[i]
		r := newPlan.Texts[j]
		if l.Date.Equal(r.Date) {
			return false
		}

		return l.Date.Before(r.Date)
	})

	return newPlan
}

func (p plan) removeByPlanTag(planTag string) plan {
	var entries []planEntry
	for _, e := range p.Entries {
		if e.PlanTag == planTag {
			continue
		}

		entries = append(entries, e)
	}

	var texts []planText
	for _, t := range p.Texts {
		if t.PlanTag == planTag {
			continue
		}

		texts = append(texts, t)
	}

	p.Entries = entries
	p.Texts = texts
	return p
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
		PlanTag:    e.PlanTag,
		HashTag:    e.HashTag,
		StartAt:    jst.From(e.StartAt),
		Source:     e.Source,
		MemberOnly: e.MemberOnly,
		CollaboID:  e.CollaboID,
	}
}

func (ts planTextSlice) PlanTexts() []model.PlanText {
	var res []model.PlanText
	for _, t := range ts {
		res = append(res, t.PlanText())
	}
	return res
}

func (t planText) PlanText() model.PlanText {
	return model.PlanText{
		Date:    jst.From(t.Date),
		PlanTag: t.PlanTag,
		Text:    t.Text,
	}
}
