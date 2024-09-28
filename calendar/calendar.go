package calendar

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Calendar struct {
	lifeCycle  context.Context
	cancelFunc context.CancelFunc

	updateTextTicker     *time.Ticker
	pullFromGoogleTicker *time.Ticker

	cachedTodayEvents Events // So we don't have to spam the calendar API
	whenLastPopup     time.Time
}

func NewCalendar() *Calendar {
	return &Calendar{
		updateTextTicker:     &time.Ticker{},
		pullFromGoogleTicker: &time.Ticker{},
	}
}

func (c *Calendar) Start() error {
	c.lifeCycle, c.cancelFunc = context.WithCancel(context.Background())
	c.updateTextTicker = time.NewTicker(10 * time.Second)
	c.pullFromGoogleTicker = time.NewTicker(5 * time.Minute)
	go c.keepEventsUpdated()
	go c.keepPrintingNextEvent()

	err := c.updateCachedEvents() // Manually update the events once on start
	if err != nil {
		return fmt.Errorf("updating events: %w", err)
	}
	c.printNextEvent() // Manually print first time

	return nil
}

func (c *Calendar) Stop() error {
	c.updateTextTicker.Stop()
	c.pullFromGoogleTicker.Stop()
	c.cancelFunc()
	return nil
}

func (c *Calendar) GetNextUpcomingEvent() (Event, bool) {
	return c.cachedTodayEvents.GetNextUpcomingEvent()
}

// printNextEvent prints the next event in the calendar to STDOUT
// This will this be picked up by Tmux and displayed in the status bar
func (c *Calendar) keepPrintingNextEvent() {
	for {
		select {
		case <-c.updateTextTicker.C:
			c.printNextEvent()
		case <-c.lifeCycle.Done():
			return
		}
	}
}

func (c *Calendar) printNextEvent() {
	event, ok := c.GetNextUpcomingEvent()
	if ok {
		fmt.Printf("Morten's next event: %q [%s]\n", event.Title, event.TimeUntilAsString())
	} else {
		fmt.Printf("Morten has no more meetings today! Boogie time!")
		return
	}

	// Print popup if time.Until(event) is less or equal to 1 minute
	if time.Until(event.TimeStamp) <= 1*time.Minute {
		if time.Since(c.whenLastPopup) < 10*time.Minute {
			return
		}

		c.whenLastPopup = time.Now()
		cmd := exec.Command("tmux", "display-popup", "-S", "fg=#eba0ac", "-w50%", "-h50%", "-d", "#{pane_current_path}", "-T", "You got a meeting!", "echo", fmt.Sprintf("Meeting: %q", event.Title))
		_ = cmd.Run()
	}
}

func (c *Calendar) keepEventsUpdated() {
	for {
		select {
		case <-c.pullFromGoogleTicker.C:
			fmt.Printf("Updating events\n")
			err := c.updateCachedEvents()
			if err != nil {
				fmt.Println(err)
			}
		case <-c.lifeCycle.Done():
			return
		}
	}
}

func (c *Calendar) updateCachedEvents() error {
	toDay := time.Now()
	tomorrow := toDay.AddDate(0, 0, 1)

	eventsAsString, err := getCalendar(toDay, tomorrow)
	if err != nil {
		return fmt.Errorf("getting calendar: %w", err)
	}

	c.cachedTodayEvents = parseEvents(eventsAsString)

	return nil
}

// gcalcli search "*" 2024-09-09 2024-09-10
func getCalendar(fromDay time.Time, toDay time.Time) (string, error) {
	format := "2006-01-02"
	cmd := exec.Command("gcalcli", "search", "*", fromDay.Format(format), toDay.Format(format))
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
			TimeStamp: eventTime,
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
