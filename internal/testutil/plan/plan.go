package plan

import (
	"github.com/yaegaki/dotlive-schedule-server/jst"
	"github.com/yaegaki/dotlive-schedule-server/model"
)

// CreateEntry .
func CreateEntry(date jst.Time, actor model.Actor, hour, min int) model.PlanEntry {
	return model.PlanEntry{
		ActorID: actor.ID,
		StartAt: jst.Date(date.Year(), date.Month(), date.Day(), hour, min),
		Source:  model.VideoSourceYoutube,
	}
}

// CreateEntryBilibili .
func CreateEntryBilibili(date jst.Time, actor model.Actor, hour, min int) model.PlanEntry {
	return model.PlanEntry{
		ActorID: actor.ID,
		StartAt: jst.Date(date.Year(), date.Month(), date.Day(), hour, min),
		Source:  model.VideoSourceBilibili,
	}
}
