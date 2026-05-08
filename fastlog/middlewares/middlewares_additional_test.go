package middlewares

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

type captureHandler struct {
	records []slog.Record
	attrs   []slog.Attr
	group   string
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *captureHandler) Handle(_ context.Context, record slog.Record) error {
	h.records = append(h.records, record.Clone())

	return nil
}

func (h *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := &captureHandler{
		group: h.group,
		attrs: append([]slog.Attr{}, h.attrs...),
	}
	next.attrs = append(next.attrs, attrs...)

	return next
}

func (h *captureHandler) WithGroup(name string) slog.Handler {
	return &captureHandler{
		attrs: append([]slog.Attr{}, h.attrs...),
		group: name,
	}
}

func TestErrorFormattingMiddleware(t *testing.T) {
	expectedErr := errors.New("boom")

	cases := []struct {
		name           string
		attrs          []slog.Attr
		wantErrorGroup bool
	}{
		{
			name: "formats error attr",
			attrs: []slog.Attr{
				slog.Any("error", expectedErr),
				slog.String("kept", "value"),
			},
			wantErrorGroup: true,
		},
		{
			name: "keeps non-error any attr",
			attrs: []slog.Attr{
				slog.Any("error", struct{ Message string }{Message: "text"}),
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			capture := &captureHandler{}
			record := slog.NewRecord(time.Now(), slog.LevelError, "message", 0)
			record.AddAttrs(test.attrs...)

			err := ErrorFormattingMiddleware(context.Background(), record, capture.Handle)
			if err != nil {
				t.Fatalf("middleware: %v", err)
			}

			attrs := recordAttrs(capture.records[0])
			errorAttr, ok := findAttr(attrs, "error")
			if !test.wantErrorGroup {
				if !ok || errorAttr.Value.Kind() == slog.KindGroup {
					t.Fatalf("error attr = %v, want unformatted attr", errorAttr)
				}

				return
			}

			if !ok {
				t.Fatal("error attr not found")
			}

			if errorAttr.Value.Kind() != slog.KindGroup {
				t.Fatalf("error attr kind = %v, want group", errorAttr.Value.Kind())
			}

			if valueFromGroup(errorAttr, "message") != expectedErr.Error() {
				t.Fatalf("error message = %q, want %q", valueFromGroup(errorAttr, "message"), expectedErr.Error())
			}

			if valueFromGroup(errorAttr, "type") == "" {
				t.Fatal("error type is empty")
			}
		})
	}
}

func TestCallerMiddleware(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "adds caller attr",
			run: func(t *testing.T) {
				t.Helper()

				capture := &captureHandler{}
				record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
				record.AddAttrs(slog.String("kept", "value"))

				err := CallerMiddleware(context.Background(), record, capture.Handle)
				if err != nil {
					t.Fatalf("middleware: %v", err)
				}

				attrs := recordAttrs(capture.records[0])
				caller, ok := findAttr(attrs, "caller")
				if !ok {
					t.Fatal("caller attr not found")
				}
				_ = caller
			},
		},
		{
			name: "caller returns file and line for current stack",
			run: func(t *testing.T) {
				t.Helper()

				if got := Caller(0); !strings.Contains(got, ".go:") {
					t.Fatalf("caller = %q, want file:line", got)
				}
			},
		},
		{
			name: "caller returns empty when runtime cannot resolve depth",
			run: func(t *testing.T) {
				t.Helper()

				if got := Caller(100000); got != "" {
					t.Fatalf("caller = %q, want empty", got)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func TestGDPRMiddleware(t *testing.T) {
	cases := []struct {
		name      string
		attrs     []slog.Attr
		wantEmail string
		wantName  string
	}{
		{
			name: "masks direct pii attr",
			attrs: []slog.Attr{
				slog.String("email", "user@example.com"),
				slog.String("name", "alice"),
			},
			wantEmail: maskValue,
			wantName:  "alice",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			capture := &captureHandler{}
			handler := &gdprMiddleware{next: capture}
			record := slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
			record.AddAttrs(test.attrs...)

			err := handler.Handle(context.Background(), record)
			if err != nil {
				t.Fatalf("handle: %v", err)
			}

			attrs := recordAttrs(capture.records[0])

			email, _ := findAttr(attrs, "email")
			if email.Value.String() != test.wantEmail {
				t.Fatalf("email = %q, want %q", email.Value.String(), test.wantEmail)
			}

			name, _ := findAttr(attrs, "name")
			if name.Value.String() != test.wantName {
				t.Fatalf("name = %q, want %q", name.Value.String(), test.wantName)
			}
		})
	}
}

func TestNewGDPRMiddleware(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "wraps handler"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			base := &captureHandler{}
			handler := NewGDPRMiddleware()(base)

			gdpr, ok := handler.(*gdprMiddleware)
			if !ok {
				t.Fatalf("handler type = %T, want *gdprMiddleware", handler)
			}

			if gdpr.next != base {
				t.Fatal("wrapped handler was not stored")
			}

			if !gdpr.Enabled(context.Background(), slog.LevelInfo) {
				t.Fatal("handler should be enabled")
			}
		})
	}
}

func TestGDPRMiddlewareWithAttrs(t *testing.T) {
	cases := []struct {
		name      string
		anonymize bool
		attrs     []slog.Attr
		wantName  string
		wantPhone string
	}{
		{
			name: "masks only pii attrs by default",
			attrs: []slog.Attr{
				slog.String("name", "alice"),
				slog.String("phone", "123"),
			},
			wantName:  "alice",
			wantPhone: maskValue,
		},
		{
			name:      "masks all attrs inside pii group",
			anonymize: true,
			attrs: []slog.Attr{
				slog.String("name", "alice"),
				slog.String("phone", "123"),
			},
			wantName:  maskValue,
			wantPhone: maskValue,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			handler := &gdprMiddleware{
				next:      &captureHandler{},
				anonymize: test.anonymize,
			}

			next := handler.WithAttrs(test.attrs).(*gdprMiddleware)
			capture := next.next.(*captureHandler)

			name, _ := findAttr(capture.attrs, "name")
			if name.Value.String() != test.wantName {
				t.Fatalf("name = %q, want %q", name.Value.String(), test.wantName)
			}

			phone, _ := findAttr(capture.attrs, "phone")
			if phone.Value.String() != test.wantPhone {
				t.Fatalf("phone = %q, want %q", phone.Value.String(), test.wantPhone)
			}
		})
	}
}

func TestGDPRMiddlewareWithGroup(t *testing.T) {
	cases := []struct {
		name          string
		group         string
		wantAnonymize bool
	}{
		{
			name:          "pii group enables anonymize",
			group:         "email",
			wantAnonymize: true,
		},
		{
			name:  "non-pii group preserves anonymize flag",
			group: "service",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			handler := &gdprMiddleware{next: &captureHandler{}}
			next := handler.WithGroup(test.group).(*gdprMiddleware)

			if next.anonymize != test.wantAnonymize {
				t.Fatalf("anonymize = %v, want %v", next.anonymize, test.wantAnonymize)
			}
		})
	}
}

func TestAnonymizeGroup(t *testing.T) {
	cases := []struct {
		name string
		attr slog.Attr
	}{
		{
			name: "group attr masks children",
			attr: slog.Group("user",
				slog.String("name", "Ada"),
				slog.String("email", "ada@example.test"),
			),
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			attr := anonymize(test.attr)
			for _, child := range attr.Value.Group() {
				if child.Value.String() != maskValue {
					t.Fatalf("child %s = %q, want mask", child.Key, child.Value.String())
				}
			}
		})
	}
}

func recordAttrs(record slog.Record) []slog.Attr {
	var attrs []slog.Attr
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)

		return true
	})

	return attrs
}

func findAttr(attrs []slog.Attr, key string) (slog.Attr, bool) {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr, true
		}
	}

	return slog.Attr{}, false
}

func valueFromGroup(attr slog.Attr, key string) string {
	for _, groupAttr := range attr.Value.Group() {
		if groupAttr.Key == key {
			return groupAttr.Value.String()
		}
	}

	return ""
}
