package main

import (
	calendarv2 "calendar/calendar_v2"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config := calendarv2.NewConfig(
		calendarv2.WithEventResolver(&calendarv2.GoogleCalendar{}),
	)

	calendar := calendarv2.New(config)
	err := calendar.Start()
	if err != nil {
		fmt.Println("Error starting calendar:", err)
		os.Exit(1)
	}
	defer calendar.Stop()

	// Wait for exit signal
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done // Will block here until user hits ctrl+c
}
