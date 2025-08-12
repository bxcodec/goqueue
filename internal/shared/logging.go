package shared

import (
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var loggingSetupOnce sync.Once

// SetupLogging configures zerolog with sensible defaults for goqueue.
// This function is safe to call multiple times - it will only execute once.
//
// It sets:
// - TimeFieldFormat to Unix timestamp format
// - ErrorStackMarshaler to include stack traces in error logs
func SetupLogging() {
	loggingSetupOnce.Do(func() {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	})
}

//nolint:gochecknoinits // Required for auto-setup of logging when any goqueue package is imported
func init() {
	// Automatically setup logging when any goqueue package is imported
	SetupLogging()
}
