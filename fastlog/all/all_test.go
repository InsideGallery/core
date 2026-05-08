//go:build !fastlog_minimal

package all_test

import (
	"reflect"
	"testing"

	_ "github.com/InsideGallery/core/fastlog/all"

	"github.com/InsideGallery/core/fastlog/handlers"
)

func TestFastlogAllRegistersHandlers(t *testing.T) {
	t.Parallel()

	registered := registeredOutKinds(t, handlers.DefaultRegistry())
	cases := []struct {
		name string
		kind string
	}{
		{name: "stderr", kind: "stderr"},
		{name: "stdout", kind: "stdout"},
		{name: "nop", kind: "nop"},
		{name: "logfile", kind: "file"},
		{name: "logstash", kind: "logstash"},
		{name: "otel", kind: "otel"},
		{name: "datadog", kind: "datadog"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if _, ok := registered[test.kind]; !ok {
				t.Fatalf("registered handlers missing %q", test.kind)
			}
		})
	}
}

func registeredOutKinds(t *testing.T, registry *handlers.Registry) map[string]struct{} {
	t.Helper()

	value := reflect.Indirect(reflect.ValueOf(registry))
	if !value.IsValid() {
		t.Fatal("registry is nil")
	}

	kinds := make(map[string]struct{})
	for _, fieldName := range []string{"writers", "handlers", "handlerFactories"} {
		field := value.FieldByName(fieldName)
		if !field.IsValid() {
			t.Fatalf("registry field %q is missing", fieldName)
		}

		for _, key := range field.MapKeys() {
			kinds[key.String()] = struct{}{}
		}
	}

	return kinds
}
