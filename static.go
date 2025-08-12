package goqueue

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/bxcodec/goqueue/internal/shared"
)

// SetupLogging configures zerolog with sensible defaults for goqueue.
// This is automatically called when importing consumer or publisher packages,
// but can be called explicitly for custom configuration.
//
// Note: This function is safe to call multiple times.
func SetupLogging() {
	shared.SetupLogging()
}

// SetupLoggingWithDefaults configures zerolog and sets up a default global logger
// with console output and reasonable formatting for development.
// This is useful for development environments or when you want pretty-printed logs.
func SetupLoggingWithDefaults() {
	SetupLogging()

	// Set up a nice console logger for development
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().
		Caller().
		Logger()
}
