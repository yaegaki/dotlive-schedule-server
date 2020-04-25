package plan

import (
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

// EntryPart .
type EntryPart struct {
	Actor  model.Actor
	Hour   int
	Min    int
	Source string
}

// CreatePlan .
func CreatePlan(d jst.Time, parts []EntryPart) model.Plan {
	var entries []model.PlanEntry
	for _, p := range parts {
		entries = append(entries, model.PlanEntry{
			ActorID: p.Actor.ID,
			StartAt: jst.Date(d.Year(), d.Month(), d.Day(), p.Hour, p.Min),
			Source:  p.Source,
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
		Actor:  actor,
		Hour:   hour,
		Min:    min,
		Source: model.VideoSourceYoutube,
	}
}

// CreateEntryPartBilibili .
func CreateEntryPartBilibili(actor model.Actor, hour, min int) EntryPart {
	return EntryPart{
		Actor:  actor,
		Hour:   hour,
		Min:    min,
		Source: model.VideoSourceBilibili,
	}
}
