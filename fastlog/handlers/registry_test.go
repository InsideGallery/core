package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestGetBuildsJSONHandlerFromRegisteredWriter(t *testing.T) {
	var buf bytes.Buffer

	RegisterWriter("unit-json-writer", func() (io.Writer, *slog.HandlerOptions, error) {
		return &buf, nil, nil
	})

	handler, err := Get("unit-json-writer", FormatJSON, slog.LevelDebug)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "hello", 0)
	record.AddAttrs(slog.String("key", "value"))

	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	body := buf.String()
	for _, want := range []string{`"msg":"hello"`, `"key":"value"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("JSON log missing %q in %s", want, body)
		}
	}
}

func TestGetBuildsTextHandlerFromRegisteredWriter(t *testing.T) {
	var buf bytes.Buffer

	RegisterWriter("unit-text-writer", func() (io.Writer, *slog.HandlerOptions, error) {
		return &buf, &slog.HandlerOptions{Level: slog.LevelWarn}, nil
	})

	handler, err := Get("unit-text-writer", FormatText, slog.LevelDebug)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	record := slog.NewRecord(time.Now(), slog.LevelWarn, "hello", 0)
	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if !strings.Contains(buf.String(), `msg=hello`) {
		t.Fatalf("text log missing message in %s", buf.String())
	}
}

func TestGetUsesRegisteredHandlerFunc(t *testing.T) {
	var buf bytes.Buffer

	RegisterHandlerFunc("unit-handler-func", func() (slog.Handler, error) {
		return slog.NewTextHandler(&buf, nil), nil
	})

	handler, err := Get("unit-handler-func", "ignored", slog.LevelInfo)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "direct", 0)
	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if !strings.Contains(buf.String(), `msg=direct`) {
		t.Fatalf("handler log missing message in %s", buf.String())
	}
}

func TestGetReturnsWriterError(t *testing.T) {
	wantErr := errors.New("writer failed")

	RegisterWriter("unit-writer-error", func() (io.Writer, *slog.HandlerOptions, error) {
		return nil, nil, wantErr
	})

	if _, err := Get("unit-writer-error", FormatJSON, slog.LevelInfo); !errors.Is(err, wantErr) {
		t.Fatalf("Get() error = %v, want %v", err, wantErr)
	}
}

func TestGetReturnsNotFoundForUnknownKind(t *testing.T) {
	if _, err := Get("unit-missing", FormatJSON, slog.LevelInfo); !errors.Is(err, ErrNotFoundHandler) {
		t.Fatalf("Get() error = %v, want ErrNotFoundHandler", err)
	}
}
