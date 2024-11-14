package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
)

type CustomLogHandler struct {
	slog.Handler
	l *log.Logger
}

func (h *CustomLogHandler) Handle(ctx context.Context, record slog.Record) error {
	level := fmt.Sprintf("[%v]", record.Level.String())
	timeStr := record.Time.Format("[2006-01-02T15:04:05.999]")
	msg := record.Message

	h.l.Println(timeStr, level, "[request_id= ]", "[tenant_id= ]", "[thread= ]", "[class= ]", msg)

	return nil
}

func NewCustomLogHandler(out io.Writer) *CustomLogHandler {
	h := &CustomLogHandler{
		Handler: slog.NewTextHandler(out, nil),
		l:       log.New(out, "", 0),
	}
	return h
}
