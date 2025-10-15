package ai

import "io"

type Config struct {
	debugOutput io.Writer
}

func defaultConfig() *Config {
	return &Config{
		debugOutput: io.Discard,
	}
}

type OptionFunc func(*Config)

func WithDebugOutput(w io.Writer) func(*Config) {
	return func(c *Config) {
		c.debugOutput = w
	}
}
