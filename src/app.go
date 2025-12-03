package src

import (
	"io"
	"log"
	"log/slog"
	"os"
)

// GooglePhotosCLI handles event emission and logging for the CLI
type GooglePhotosCLI struct {
	eventCallback func(event string, data any)
	logger        *slog.Logger
}

func NewGooglePhotosCLI(eventCallback func(event string, data any), logLevel slog.Level) *GooglePhotosCLI {
	var logger *slog.Logger

	if logLevel <= slog.LevelInfo {
		// For info level and below, use io.Discard to hide logs
		logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: logLevel,
		}))
		// Disable HTTP client debug logs for info level
		SetHTTPClientLogger(log.New(io.Discard, "", 0))
	} else {
		// For debug level, log to stderr
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		}))
		// Enable HTTP client debug logs for debug level
		SetHTTPClientLogger(nil) // nil will use default retryablehttp logger
	}

	return &GooglePhotosCLI{
		eventCallback: eventCallback,
		logger:        logger,
	}
}

func (c *GooglePhotosCLI) EmitEvent(event string, data any) {
	if c.eventCallback != nil {
		c.eventCallback(event, data)
	}
}

func (c *GooglePhotosCLI) GetLogger() *slog.Logger {
	return c.logger
}
