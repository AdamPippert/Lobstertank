package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/AdamPippert/Lobstertank/internal/config"
)

// Event represents a single auditable action.
type Event struct {
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"`
	Resource  string `json:"resource,omitempty"`
	Subject   string `json:"subject,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

// Logger writes structured audit events.
type Logger struct {
	mu      sync.Mutex
	enabled bool
	output  *os.File
}

// New creates an audit logger from the given configuration.
func New(cfg config.AuditConfig) *Logger {
	l := &Logger{enabled: cfg.Enabled}
	if !cfg.Enabled {
		return l
	}

	switch cfg.Output {
	case "file":
		f, err := os.OpenFile(cfg.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			slog.Error("failed to open audit log file, falling back to stdout", "path", cfg.Path, "error", err)
			l.output = os.Stdout
		} else {
			l.output = f
		}
	default:
		l.output = os.Stdout
	}

	return l
}

// Log records an audit event. It is safe for concurrent use.
func (l *Logger) Log(_ context.Context, evt Event) {
	if !l.enabled {
		return
	}

	evt.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)

	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(evt)
	if err != nil {
		slog.Error("failed to marshal audit event", "error", err)
		return
	}
	data = append(data, '\n')

	if _, err := l.output.Write(data); err != nil {
		slog.Error("failed to write audit event", "error", err)
	}
}

// Close releases any resources held by the logger.
func (l *Logger) Close() error {
	if l.output != nil && l.output != os.Stdout && l.output != os.Stderr {
		return l.output.Close()
	}
	return nil
}
