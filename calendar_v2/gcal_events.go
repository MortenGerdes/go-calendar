package calendarv2

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type GoogleCalendar struct {
}

func (g *GoogleCalendar) GetEventsToday(ctx context.Context) (Events, error) {
	fromAsDay := time.Now()
	toAsDay := time.Now().Add(24 * time.Hour)

	events, err := getCalendar(fromAsDay, toAsDay)
	if err != nil {
		return nil, err
	}

	return parseEvents(events), nil
}

// gcalcli search "*" 2024-09-09 2024-09-10
func getCalendar(fromAsDay time.Time, toAsDay time.Time) (string, error) {
	format := "2006-01-02"
	cmd := exec.Command("gcalcli", "search", "*", fromAsDay.Format(format), toAsDay.Format(format))
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return removeColor(string(out)), nil
}

func parseEvents(events string) []Event {
	var (
		rawEvents   []string
		result      []Event
		splitEvents = strings.Split(events, "\n")
	)

	for _, event := range splitEvents {
		if event != "" {
			rawEvents = append(rawEvents, event)
		}
	}

	//Replace any number of spaces with a single space
	for i, event := range rawEvents {
		rawEvents[i] = strings.Join(strings.Fields(event), " ")
	}

	for _, event := range rawEvents {
		var (
			splitEvent    = strings.SplitN(event, " ", 3)
			copenhagen, _ = time.LoadLocation("Europe/Copenhagen")
			eventTime, _  = time.ParseInLocation("2006-01-02 15:04", splitEvent[0]+" "+splitEvent[1], copenhagen)
		)

		result = append(result, Event{
			StartTime: eventTime,
			Title:     splitEvent[2],
		})
	}

	return result
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func removeColor(str string) string {
	return re.ReplaceAllString(str, "")
}
