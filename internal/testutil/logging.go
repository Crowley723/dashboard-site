package testutil

import (
	"context"
	"log/slog"
	"sync"
)

type TestLogHandler struct {
	mu      sync.Mutex
	records []TestLogRecord
}

type TestLogRecord struct {
	Level   slog.Level
	Message string
	Attrs   map[string]any
}

func NewTestLogHandler() *TestLogHandler {
	return &TestLogHandler{
		records: make([]TestLogRecord, 0),
	}
}

func (h *TestLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *TestLogHandler) Handle(ctx context.Context, record slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	attrs := make(map[string]any)
	record.Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value.Any()
		return true
	})
	h.records = append(h.records, TestLogRecord{
		Level:   record.Level,
		Message: record.Message,
		Attrs:   attrs,
	})

	return nil
}

func (h *TestLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *TestLogHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *TestLogHandler) GetRecords() []TestLogRecord {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]TestLogRecord(nil), h.records...)
}

func (h *TestLogHandler) GetRecordsByLevel(level slog.Level) []TestLogRecord {
	h.mu.Lock()
	defer h.mu.Unlock()

	var filtered []TestLogRecord
	for _, record := range h.records {
		if record.Level == level {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

func (h *TestLogHandler) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = h.records[:0]
}

func (h *TestLogHandler) ContainsMessage(level slog.Level, message string) bool {
	records := h.GetRecordsByLevel(level)
	for _, record := range records {
		if record.Message == message {
			return true
		}
	}
	return false
}

func (h *TestLogHandler) CountByLevel(level slog.Level) int {
	return len(h.GetRecordsByLevel(level))
}
