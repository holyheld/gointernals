package closeutil

import (
	"context"
	"io"
	"log/slog"

	slogutil "github.com/holyheld/gointernals/slogutil"
)

// CloseOrLog is an util function that calls [io.Closer.Close] method on provided closer
// and logs the error with provided logger if Close fails.
func CloseOrLog(
	ctx context.Context,
	logger *slog.Logger,
	closer io.Closer,
	resource string,
) {
	err := closer.Close()
	if err != nil && logger != nil {
		logger.WarnContext(
			ctx,
			"Failed to close resource",
			slog.String("resource", resource),
			slogutil.Error(err),
		)
	}
}

// CloseOrSuppress is an util function that calls [io.Closer.Close] method on provided closer
// and ignores the output.
func CloseOrSuppress(c io.Closer) {
	_ = c.Close()
}
