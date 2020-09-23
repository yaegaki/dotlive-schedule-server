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
				ActorID: "A",
			},
			planEntry{
				StartAt: d.Add(21 * time.Hour).Time(),
				ActorID: "B",
			},
		},
		Text: `20:00~ A
21:00~ B`,
	}

	planB := plan{
		Entries: planEntrySlice{
			planEntry{
				StartAt: d.Add(18 * time.Hour).Time(),
				ActorID: "C",
			},
			planEntry{
				StartAt: d.Add(19 * time.Hour).Time(),
				ActorID: "D",
			},
		},
		Text: `18:00~ C
19:00~ D`,
	}

	test := func(p plan) {
		expectIDs := []string{
			"C", "D", "A", "B",
		}
		if len(p.Entries) != len(expectIDs) {
			t.Fatal("merge failed")
		}

		for i, e := range p.Entries {
			if expectIDs[i] != e.ActorID {
				t.Fatalf("id got: %v, expect: %v", e.ActorID, expectIDs[i])
			}
		}

		expectText := `18:00~ C
19:00~ D
20:00~ A
21:00~ B`
		if p.Text != expectText {
			t.Fatalf("text got:%v, expect:%v", p.Text, expectText)
		}
	}

	test(planA.Merge(planB))
	// 逆順にマージしても結果は同じ
	test(planB.Merge(planA))
	// 同じ計画をマージしても結果が変わらない
	test(planA.Merge(planB).Merge(planA).Merge(planB))
}
