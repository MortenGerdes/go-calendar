package calendar

import (
	"fmt"
	"time"
)

type Event struct {
	TimeStamp time.Time
	Title     string
}

// TimeUntil returns a human readable string of the time until the event
// in the format "1 hour 30 minutes"
func (event *Event) TimeUntilAsString() string {
	eventTime := event.TimeStamp
	timeUntil := time.Until(eventTime)

	days := int(timeUntil.Hours()) / 24
	hours := int(timeUntil.Hours())
	minutes := int(timeUntil.Minutes()) % 60

	result := ""
	if days > 0 {
		result += fmt.Sprintf("%d days", days)
	}

	if hours > 0 {
		if days > 0 {
			result += " "
		}
		result += fmt.Sprintf("%d hour", hours)
	}

	if minutes > 0 {
		if hours > 0 {
			result += " "
		}
		result += fmt.Sprintf("%d minutes", minutes)
	}
	return result
}

type Events []Event

func (e Events) GetNextUpcomingEvent() (Event, bool) {
	var (
		currentBestEvent = Event{}
		currentTime      = time.Now()
	)

	for _, event := range e {
		if event.TimeStamp.Before(currentTime) {
			continue
		}

		//If event is closer to current time than currentBestEvent
		if currentBestEvent.TimeStamp.IsZero() || event.TimeStamp.Before(currentBestEvent.TimeStamp) {
			currentBestEvent = event
		}
	}

	return currentBestEvent, !currentBestEvent.TimeStamp.IsZero()
}

//tmux display-popup -S "fg=#eba0ac" -w50% -h50% -d '#{pane_current_path}' -T "You got a meeting" echo "Move fatass"
