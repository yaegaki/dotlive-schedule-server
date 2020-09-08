package plan

import (
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

// EntryPart .
type EntryPart struct {
	Actor     model.Actor
	HashTag   string
	Hour      int
	Min       int
	Source    string
	CollaboID int
}

// CreatePlan .
func CreatePlan(d jst.Time, parts []EntryPart) model.Plan {
	var entries []model.PlanEntry
	for _, p := range parts {
		entries = append(entries, model.PlanEntry{
			ActorID:   p.Actor.ID,
			HashTag:   p.HashTag,
			StartAt:   jst.Date(d.Year(), d.Month(), d.Day(), p.Hour, p.Min),
			Source:    p.Source,
			CollaboID: p.CollaboID,
		})
	}

	return model.Plan{
		Date:    d,
		Entries: entries,
	}
}

// CreateEntryPart .
func CreateEntryPart(actor model.Actor, hour, min int) EntryPart {
	return EntryPart{
		Actor:     actor,
		Hour:      hour,
		Min:       min,
		Source:    model.VideoSourceYoutube,
		CollaboID: 0,
	}
}

// CreateEntryPartCollabo .
func CreateEntryPartCollabo(actor model.Actor, hour, min int, collaboID int) EntryPart {
	return EntryPart{
		Actor:     actor,
		Hour:      hour,
		Min:       min,
		Source:    model.VideoSourceYoutube,
		CollaboID: collaboID,
	}
}

// CreateEntryPartCollaboHashTag .
func CreateEntryPartCollaboHashTag(hour, min int, hashTag string) EntryPart {
	return EntryPart{
		Actor: model.Actor{
			ID: model.UnknownActorID,
		},
		HashTag:   hashTag,
		Hour:      hour,
		Min:       min,
		Source:    model.VideoSourceYoutube,
		CollaboID: 0,
	}
}

// CreateEntryPartBilibili .
func CreateEntryPartBilibili(actor model.Actor, hour, min int) EntryPart {
	return EntryPart{
		Actor:     actor,
		Hour:      hour,
		Min:       min,
		Source:    model.VideoSourceBilibili,
		CollaboID: 0,
	}
}

// CreateEntryPartMildom .
func CreateEntryPartMildom(actor model.Actor, hour, min int) EntryPart {
	return EntryPart{
		Actor:     actor,
		Hour:      hour,
		Min:       min,
		Source:    model.VideoSourceMildom,
		CollaboID: 0,
	}
}
