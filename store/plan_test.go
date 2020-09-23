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

	merged := planA.Merge(planB)

	expectIDs := []string{
		"C", "D", "A", "B",
	}
	if len(merged.Entries) != len(expectIDs) {
		t.Fatal("merge failed")
	}

	for i, e := range merged.Entries {
		if expectIDs[i] != e.ActorID {
			t.Fatalf("id got: %v, expect: %v", expectIDs[i], e.ActorID)
		}
	}

	expectText := `18:00~ C
19:00~ D
20:00~ A
21:00~ B`
	if merged.Text != expectText {
		t.Fatalf("text got:%v, expect:%v", merged.Text, expectText)
	}

	// 同じ計画をマージしても結果が変わらない
	merged2 := merged.Merge(planA).Merge(planB)
	if merged2.Text != expectText {
		t.Fatalf("text got:%v, expect:%v", merged.Text, expectText)
	}
}
