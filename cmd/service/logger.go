package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

////////////////////////////////////////////////////////////////////////////////

// Handler that outputs JSON understood by the structured log agent.
// See https://cloud.google.com/logging/docs/agent/logging/configuration#special-fields
type cloudLogHandler struct{ handler slog.Handler }

const (
	LevelCritical = slog.Level(12)
)

func newCloudLogHandler(severity slog.Level) *cloudLogHandler {

	return &cloudLogHandler{handler: slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: severity,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.MessageKey {
				a.Key = "message"
			} else if a.Key == slog.SourceKey {
				a.Key = "logging.googleapis.com/sourceLocation"
			} else if a.Key == slog.LevelKey {
				a.Key = "severity"
				level := a.Value.Any().(slog.Level)
				if level == LevelCritical {
					a.Value = slog.StringValue("CRITICAL")
				}
			}
			return a
		},
	})}
}

func (h *cloudLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *cloudLogHandler) Handle(ctx context.Context, rec slog.Record) error {
	requestId := requestIdFromContext(ctx)
	if requestId != "" {
		rec = rec.Clone()
		rec.Add(requestIDKey, slog.StringValue(requestId))
	}
	return h.handler.Handle(ctx, rec)
}

func (h *cloudLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &cloudLogHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *cloudLogHandler) WithGroup(name string) slog.Handler {
	return &cloudLogHandler{handler: h.handler.WithGroup(name)}
}

const requestIDKey = "request_id"

func requestIdFromContext(ctx context.Context) string {
	claim := ctx.Value(requestIDKey)
	reqId, _ := claim.(string)

	return reqId
}

func setupLogger(severity, handler string) {
	logLevel := &slog.LevelVar{}
	err := logLevel.UnmarshalText([]byte(severity))
	if err != nil {
		slog.Error(fmt.Sprintf("failed parse severity level: %v", err))
	}

	// set default use cloud log handler
	if handler == "" {
		slog.SetDefault(slog.New(newCloudLogHandler(logLevel.Level())))
	} else {
		slog.Default().Handler().Enabled(context.Background(), logLevel.Level())
	}
}
