package calendarv2

import "time"

type Event struct {
	StartTime time.Time
	Title     string
}

type Events []Event

func (e Events) GetNextEvent(when time.Time) *Event {
	var (
		closestEvent *Event
	)

	for _, event := range e {
		if event.StartTime.Before(when) {
			continue
		}

		if closestEvent == nil || event.StartTime.Before(closestEvent.StartTime) {
			closestEvent = &event
		}
	}

	return closestEvent
}
