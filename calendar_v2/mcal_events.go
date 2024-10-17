package calendarv2

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type McalCalendar struct{}

func (m *McalCalendar) GetEventsToday(_ context.Context) (Events, error) {
	eventsAsString, err := getMcalEvents()
	if err != nil {
		return nil, err
	}

	return parseToEvents(eventsAsString), nil
}

func getMcalEvents() (string, error) {
	cmd := exec.Command("mcal", "list")
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))

	return string(out), nil
}

func parseToEvents(eventsAsString string) Events {
	var (
		lines       = strings.Split(eventsAsString, "\n")
		location, _ = time.LoadLocation("Europe/Copenhagen")
		events      = make(Events, 0)
	)

	for _, line := range lines {
		if line == "" {
			continue
		}

		var (
			parts       = strings.Split(line, " ")
			timeStr     = parts[0]
			hour        = strings.Split(timeStr, ":")[0]
			minute      = strings.Split(timeStr, ":")[1]
			eventStr    = strings.Join(parts[1:], " ")
			midnight, _ = time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02"), location)
			eventTime   = midnight.Add(time.Hour*time.Duration(atoi(hour)) + time.Minute*time.Duration(atoi(minute)))
		)

		events = append(events, Event{
			StartTime: eventTime,
			Title:     eventStr,
		})
	}

	return events
}

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return n
}
