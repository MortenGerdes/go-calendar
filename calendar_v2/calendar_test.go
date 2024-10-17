package calendarv2_test

import (
	calendarv2 "calendar/calendar_v2"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalendar(t *testing.T) {
	var (
		newEventResolverMock = func(getEventsFunc func(ctx context.Context) (calendarv2.Events, error)) *EventResolverMock {
			mock := &EventResolverMock{}
			mock.GetEventsTodayFunc = getEventsFunc
			return mock
		}
		newSut = func(mods ...func(*calendarv2.Config)) *calendarv2.Calendar {
			conf := calendarv2.NewConfig(mods...)
			return calendarv2.New(conf)
		}
	)
	t.Run("Start and stop", func(t *testing.T) {
		//Arrange
		var (
			calendar = newSut(calendarv2.WithEventResolver(newEventResolverMock(func(ctx context.Context) (calendarv2.Events, error) {
				return []calendarv2.Event{}, nil
			})))
		)

		//Act and Assert
		err := calendar.Start()
		require.NoError(t, err)

		err = calendar.Stop()
		require.NoError(t, err)
	})

	t.Run("Get next upcoming event", func(t *testing.T) {
		var (
			expectedEvent = calendarv2.Event{
				Title:     "Test Event",
				StartTime: time.Now().Add(10 * time.Minute),
			}
			writer   = &TestWriter{}
			calendar = newSut(
				calendarv2.WithEventResolver(newEventResolverMock(func(ctx context.Context) (calendarv2.Events, error) {
					return []calendarv2.Event{
						expectedEvent,
					}, nil
				})),
				calendarv2.WithWriter(writer),
				calendarv2.WithWriteFrequency(100*time.Millisecond),
			)
		)

		err := calendar.Start()
		require.NoError(t, err)
		defer calendar.Stop()

		time.Sleep(100*time.Millisecond + 20*time.Millisecond) // Wait for a write

		assertWritesContains(t, writer.Writes(), expectedEvent.Title)
	})

	t.Run("Get no upcoming event", func(t *testing.T) {
		var (
			writer   = &TestWriter{}
			calendar = newSut(
				calendarv2.WithEventResolver(newEventResolverMock(func(ctx context.Context) (calendarv2.Events, error) {
					return []calendarv2.Event{}, nil
				})),
				calendarv2.WithWriter(writer),
				calendarv2.WithWriteFrequency(100*time.Millisecond),
			)
		)

		err := calendar.Start()
		require.NoError(t, err)
		defer calendar.Stop()

		time.Sleep(100*time.Millisecond + 20*time.Millisecond) // Wait for a write

		assert.Len(t, writer.Writes(), 1+2) // +2 because of the initial write and the "updating event cache" write
		assertWritesContains(t, writer.Writes(), "No more events today... Yay! :D")
	})

	t.Run("Write to stdout", func(t *testing.T) {
		var (
			expectedAmountOfWrites = 5
			writer                 = &TestWriter{}
			calendar               = newSut(
				calendarv2.WithEventResolver(newEventResolverMock(func(ctx context.Context) (calendarv2.Events, error) {
					return []calendarv2.Event{
						{
							Title:     "Test Event",
							StartTime: time.Now(),
						},
					}, nil
				})),
				calendarv2.WithWriter(writer),
				calendarv2.WithWriteFrequency(100*time.Millisecond),
			)
		)

		err := calendar.Start()
		require.NoError(t, err)

		time.Sleep(time.Duration(expectedAmountOfWrites)*100*time.Millisecond + 20*time.Millisecond) // The 20ms is to make sure the last write has been done. So a small buffer time basically

		err = calendar.Stop()
		require.NoError(t, err)

		assert.Len(t, writer.Writes(), expectedAmountOfWrites+2) // +2 because of the initial write and the "updating event cache" write
	})

	t.Run("Stop writes to stdout", func(t *testing.T) {
		var (
			writer   = &TestWriter{}
			calendar = newSut(
				calendarv2.WithEventResolver(newEventResolverMock(func(ctx context.Context) (calendarv2.Events, error) {
					return []calendarv2.Event{
						{
							Title:     "Test Event",
							StartTime: time.Now(),
						},
					}, nil
				})),
				calendarv2.WithWriter(writer),
				calendarv2.WithWriteFrequency(100*time.Millisecond),
			)
		)

		err := calendar.Start()
		require.NoError(t, err)

		time.Sleep(100*time.Millisecond + 20*time.Millisecond) // Wait for a write
		err = calendar.Stop()
		require.NoError(t, err)

		time.Sleep(500 * time.Millisecond)  // Wait for potentially more writes
		assert.Len(t, writer.Writes(), 1+2) // +2 because of the initial write and the "updating event cache" write
	})

	t.Run("Truncates long event names", func(t *testing.T) {
		var (
			expectedEvent = calendarv2.Event{
				Title:     "This is a very lo...",
				StartTime: time.Now().Add(10 * time.Minute),
			}
			writer   = &TestWriter{}
			calendar = newSut(
				calendarv2.WithEventResolver(newEventResolverMock(func(ctx context.Context) (calendarv2.Events, error) {
					return []calendarv2.Event{
						{
							Title:     "This is a very long event name that should be truncated",
							StartTime: expectedEvent.StartTime,
						},
					}, nil
				})),
				calendarv2.WithWriter(writer),
				calendarv2.WithWriteFrequency(100*time.Millisecond),
			)
		)

		err := calendar.Start()
		require.NoError(t, err)

		time.Sleep(300*time.Millisecond + 20*time.Millisecond) // Wait for a write

		err = calendar.Stop()
		require.NoError(t, err)

		assertWritesContains(t, writer.Writes(), expectedEvent.Title)
	})
}

func assertWritesContains(t *testing.T, writes []string, expected string) {
	for _, write := range writes {
		if strings.Contains(write, expected) {
			return
		}
	}
	assert.Failf(t, "Writes doesn't contain message", "Expected %q to be in writes. Got %q", expected, writes)
}

type TestWriter struct {
	rwMu   sync.RWMutex
	writes []string
}

func (tw *TestWriter) Write(p []byte) (n int, err error) {
	tw.rwMu.Lock()
	defer tw.rwMu.Unlock()
	tw.writes = append(tw.writes, string(p))
	return len(p), nil
}

func (tw *TestWriter) Writes() []string {
	tw.rwMu.RLock()
	defer tw.rwMu.RUnlock()
	return tw.writes
}
