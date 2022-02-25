package utils

import (
	"context"
	"io"

	"github.com/rs/zerolog/log"
)

// Close closes the io.Closer interface. Logs failures with warning level and the given context on failure.
// Supposed usage as simple drop-in replacement for defer closer.Close() usage where errors could go unnoticed.
func Close(ctx context.Context, closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to close")
	}
}
