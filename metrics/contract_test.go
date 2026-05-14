package metrics //nolint:revive // package name matches directory/domain usage

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageStaysTransportAgnostic(t *testing.T) {
	disallowed := []string{
		"http.request.",
		"nats.publisher.",
		"queue.subscriptions.",
		"github.com/gofiber/fiber",
		"github.com/FrogoAI/mq-balancer",
	}

	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		contents := string(data)
		for _, snippet := range disallowed {
			if strings.Contains(contents, snippet) {
				t.Fatalf("pkg/metrics must stay transport-agnostic: %s contains %q", path, snippet)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk pkg/metrics: %v", err)
	}
}
