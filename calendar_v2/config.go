package calendarv2

import (
	"io"
	"os"
	"time"
)

type Config struct {
	WriteFequency       time.Duration
	RenewCacheFrequency time.Duration
	Writer              io.Writer
	EventResolver       EventResolver
}

func NewConfig(mods ...func(*Config)) Config {
	c := Config{
		WriteFequency:       5 * time.Second,
		RenewCacheFrequency: 10 * time.Minute,
		Writer:              os.Stdout,
	}

	for _, mod := range mods {
		mod(&c)
	}

	return c
}

func WithWriteFrequency(frequency time.Duration) func(*Config) {
	return func(c *Config) {
		c.WriteFequency = frequency
	}
}

func WithRenewCacheFrequency(frequency time.Duration) func(*Config) {
	return func(c *Config) {
		c.RenewCacheFrequency = frequency
	}
}

func WithWriter(writer io.Writer) func(*Config) {
	return func(c *Config) {
		c.Writer = writer
	}
}

func WithEventResolver(eventResolver EventResolver) func(*Config) {
	return func(c *Config) {
		c.EventResolver = eventResolver
	}
}
