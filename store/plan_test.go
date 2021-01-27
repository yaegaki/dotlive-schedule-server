package store

import (
	"testing"
	"time"

	"github.com/yaegaki/dotlive-schedule-server/jst"
)

func TestPlan(t *testing.T) {
	d := jst.ShortDate(2020, 9, 23)
	planA := plan{
		Entries: planEntrySlice{
			planEntry{
				StartAt: d.Add(20 * time.Hour).Time(),
				PlanTag: "②",
				ActorID: "A",
			},
			planEntry{
				StartAt: d.Add(21 * time.Hour).Time(),
				PlanTag: "②",
				ActorID: "B",
			},
		},
		Texts: planTextSlice{
			{
				Date:    d.Add(20 * time.Hour).Time(),
				PlanTag: "②",
				Text: `20:00~ A
21:00~ B`,
			},
		},
	}

	planB := plan{
		Entries: planEntrySlice{
			planEntry{
				StartAt: d.Add(18 * time.Hour).Time(),
				PlanTag: "①",
				ActorID: "C",
			},
			planEntry{
				StartAt: d.Add(19 * time.Hour).Time(),
				PlanTag: "①",
				ActorID: "D",
			},
		},
		Texts: planTextSlice{
			{
				Date:    d.Add(18 * time.Hour).Time(),
				PlanTag: "①",
				Text: `18:00~ C
19:00~ D`,
			},
		},
	}

	modifiedPlanB := plan{
		Entries: planEntrySlice{
			planEntry{
				StartAt: d.Add(17 * time.Hour).Time(),
				PlanTag: "②",
				ActorID: "C",
			},
		},
		Texts: planTextSlice{
			{
				Date:    d.Add(17 * time.Hour).Time(),
				PlanTag: "①",
				Text:    `17:00~ C`,
			},
		},
	}

	test := func(p plan, ids []string, expectText string) {
		if len(p.Entries) != len(ids) {
			t.Fatal("merge failed")
		}

		for i, e := range p.Entries {
			if ids[i] != e.ActorID {
				t.Fatalf("id got: %v, expect: %v", e.ActorID, ids[i])
			}
		}

		text := p.Plan().Text()
		if text != expectText {
			t.Fatalf("text got:%v, expect:%v", text, expectText)
		}
	}

	expectIDs := []string{
		"C", "D", "A", "B",
	}

	expectText := `18:00~ C
19:00~ D
20:00~ A
21:00~ B`

	test(planA.Merge(planB, "①"), expectIDs, expectText)
	// 逆順にマージしても結果は同じ
	test(planB.Merge(planA, "②"), expectIDs, expectText)
	// 同じ計画をマージしても結果が変わらない
	test(planA.Merge(planB, "①").Merge(planA, "②").Merge(planB, "①"), expectIDs, expectText)

	expectIDs = []string{
		"C", "A", "B",
	}
	expectText = `17:00~ C
20:00~ A
21:00~ B`
	// 修正があった場合
	test(planB.Merge(planA, "②").Merge(modifiedPlanB, "①"), expectIDs, expectText)
}
