package calendarv2

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"
)

type EventResolver interface {
	GetEventsToday(ctx context.Context) ([]Event, error)
}

type Calendar struct {
	lifeCycle  context.Context
	cancelFunc context.CancelFunc

	writeTicker    *time.Ticker
	writeFrequency time.Duration
	writer         io.Writer

	cacheTicker         *time.Ticker
	renewCacheFrequency time.Duration

	eventResolver EventResolver
	eventCache    Events

	lastPopupTime time.Time
}

func New(config Config) *Calendar {
	return &Calendar{
		eventResolver:       config.EventResolver,
		writeFrequency:      config.WriteFequency,
		renewCacheFrequency: config.RenewCacheFrequency,
		writer:              config.Writer,
	}
}

func (c *Calendar) Start() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	c.cancelFunc = cancelFunc
	c.lifeCycle = ctx
	c.writeTicker = time.NewTicker(c.writeFrequency)
	c.cacheTicker = time.NewTicker(c.renewCacheFrequency)

	//Initiate kick-off
	c.onCacheTick()
	c.onWriteTick()

	go func() {
		for {
			select {
			case <-c.lifeCycle.Done():
				return
			case <-c.writeTicker.C:
				c.onWriteTick()
			}
		}
	}()

	go func() {
		for {
			select {
			case <-c.lifeCycle.Done():
				return
			case <-c.cacheTicker.C:
				c.onCacheTick()
			}
		}
	}()

	return nil
}

func (c *Calendar) Stop() error {
	c.writeTicker.Stop()
	c.cancelFunc()
	return nil
}

func (c *Calendar) getNextEventToday(ctx context.Context, after time.Time) *Event {
	return c.eventCache.GetNextEvent(after)
}

func (c *Calendar) onWriteTick() {
	nextEvent := c.getNextEventToday(context.Background(), time.Now())
	if nextEvent == nil {
		fmt.Fprintln(c.writer, "No more events today... Yay! :D")
		return
	}

	fmt.Fprintf(c.writer, "Morten's next event: %q in %s\n", truncateEventTitle(nextEvent.Title, 20), TimeUntilAsString(*nextEvent))
	c.handlePopup(*nextEvent)
}

func (c *Calendar) onCacheTick() {
	fmt.Fprintln(c.writer, "Updating event cache...")
	events, err := c.eventResolver.GetEventsToday(context.Background())
	if err != nil {
		fmt.Fprintf(c.writer, "error updating event cache: %v\n", err)
	}

	c.eventCache = Events(events)
}

func (c *Calendar) handlePopup(nextEvent Event) {
	if time.Until(nextEvent.StartTime) > 1*time.Minute {
		return
	}

	if time.Since(c.lastPopupTime) < 5*time.Minute {
		return
	}

	cmd := exec.Command("tmux", "display-popup", "-S", "fg=#eba0ac", "-w50%", "-h50%", "-d", "#{pane_current_path}", "-T", "You got a meeting!", "echo", fmt.Sprintf("Meeting: %q", nextEvent.Title))
	_ = cmd.Run()

	c.lastPopupTime = time.Now()
}

func truncateEventTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}

	return fmt.Sprintf("%s...", title[:maxLen-3])
}

// TimeUntilAsString returns a human readable string of the time until the event
// in the format "1 hour 30 minutes"
func TimeUntilAsString(event Event) string {
	eventTime := event.StartTime
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
