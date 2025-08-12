package goqueue

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}
