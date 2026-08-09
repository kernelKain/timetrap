package analyzer

import (
	"sort"

	"github.com/kernelKain/timetrap/backend/internal/domain"
)

// buildTimeline normalizes submitted events into half-open validity intervals.
// It sorts copied event slices and never mutates the scenario.
func buildTimeline(scenario domain.Scenario) []domain.TimelineLane {
	lanes := make([]domain.TimelineLane, 0, len(scenario.Objects))

	for _, object := range scenario.Objects {
		events := append([]domain.Event(nil), object.Events...)
		sort.SliceStable(events, func(left, right int) bool {
			return events[left].AtMinute < events[right].AtMinute
		})

		intervals := make([]domain.ValidityInterval, 0)
		valid := false
		start := 0

		for _, event := range events {
			switch event.Type {
			case domain.EventTypeIssue:
				if !valid {
					valid = true
					start = event.AtMinute
				}
			case domain.EventTypeRefresh:
				if !valid {
					valid = true
					start = event.AtMinute
				}
			case domain.EventTypeRevoke, domain.EventTypeExpire:
				if valid && event.AtMinute > start {
					intervals = append(intervals, domain.ValidityInterval{
						StartMinute: start,
						EndMinute:   event.AtMinute,
					})
				}
				valid = false
			}
		}

		if valid && scenario.HorizonMinutes > start {
			intervals = append(intervals, domain.ValidityInterval{
				StartMinute: start,
				EndMinute:   scenario.HorizonMinutes,
			})
		}

		lanes = append(lanes, domain.TimelineLane{
			ObjectID:  object.ClientID,
			Intervals: intervals,
		})
	}

	return lanes
}
