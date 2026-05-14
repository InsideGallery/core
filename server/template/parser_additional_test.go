package template //nolint:revive

import (
	"bytes"
	"embed"
	"errors"
	texttemplate "html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

//go:embed testdata/templates/source.tmpl
var sourceTemplates embed.FS

type leadingSlashFS struct {
	fsys fs.FS
}

const templateNameKey = "Name"

func (f leadingSlashFS) Open(name string) (fs.File, error) {
	return f.fsys.Open(strings.TrimPrefix(name, "/"))
}

func (f leadingSlashFS) Glob(pattern string) ([]string, error) {
	return []string{pattern}, nil
}

func TestTemplateEngine(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "new template executes data",
			run: func(t *testing.T) {
				t.Helper()

				prefix := "UNIT"
				dir := filepath.Join("tmp", "templates")
				SetDefaultTemplateDir(prefix, dir)
				t.Cleanup(func() {
					if err := os.Unsetenv(DefaultTemplateDir + prefix); err != nil {
						t.Fatalf("unset template dir: %v", err)
					}
				})

				absDir := GetDefaultTemplateDir(prefix)
				name := strings.TrimPrefix(filepath.Join(absDir, "page.tmpl"), "/")
				fsys := leadingSlashFS{
					fsys: fstest.MapFS{
						name: {
							Data: []byte("hello {{.Name}}"),
						},
					},
				}

				tpl, err := NewTemplate("page.tmpl", prefix, fsys, "page.tmpl")
				if err != nil {
					t.Fatalf("new template: %v", err)
				}

				engine := NewEngine()
				engine.Add(tpl)

				if !engine.Exists("page.tmpl") {
					t.Fatal("template was not added")
				}

				data, err := engine.Execute("page.tmpl", map[string]string{templateNameKey: "Ada"})
				if err != nil {
					t.Fatalf("execute template: %v", err)
				}

				if string(data) != "hello Ada" {
					t.Fatalf("data = %q, want %q", string(data), "hello Ada")
				}
			},
		},
		{
			name: "new template with explicit directory executes data",
			run: func(t *testing.T) {
				t.Helper()

				dir := filepath.Join("tmp", "explicit")
				absDir, err := filepath.Abs(dir)
				if err != nil {
					t.Fatalf("abs dir: %v", err)
				}

				name := strings.TrimPrefix(filepath.Join(absDir, "page.tmpl"), "/")
				fsys := leadingSlashFS{
					fsys: fstest.MapFS{
						name: {
							Data: []byte("hello {{.Name}}"),
						},
					},
				}

				tpl, err := NewTemplateWithDir("page.tmpl", absDir, fsys, "page.tmpl")
				if err != nil {
					t.Fatalf("new template with dir: %v", err)
				}

				var output bytes.Buffer
				if err := tpl.Execute(&output, map[string]string{templateNameKey: "Lin"}); err != nil {
					t.Fatalf("execute template: %v", err)
				}

				if output.String() != "hello Lin" {
					t.Fatalf("output = %q, want %q", output.String(), "hello Lin")
				}
			},
		},
		{
			name: "execute missing template returns sentinel",
			run: func(t *testing.T) {
				t.Helper()

				data, err := NewEngine().Execute("missing", nil)
				if !errors.Is(err, ErrNotFoundTemplate) {
					t.Fatalf("err = %v, want %v", err, ErrNotFoundTemplate)
				}

				if len(data) != 0 {
					t.Fatalf("data = %q, want empty", string(data))
				}
			},
		},
		{
			name: "execute wraps template error",
			run: func(t *testing.T) {
				t.Helper()

				raw, err := texttemplate.New("broken").Option("missingkey=error").Parse("{{.Name}}")
				if err != nil {
					t.Fatalf("parse template: %v", err)
				}

				engine := NewEngine()
				engine.Add(&Template{Name: "broken", Template: raw})

				_, err = engine.Execute("broken", map[string]string{})
				if !errors.Is(err, ErrExecuteTemplate) {
					t.Fatalf("err = %v, want %v", err, ErrExecuteTemplate)
				}
			},
		},
		{
			name: "new template by source executes embedded file",
			run: func(t *testing.T) {
				t.Helper()

				tpl, err := NewTemplateBySource(sourceTemplates, "source", "testdata/templates/source.tmpl")
				if err != nil {
					t.Fatalf("new template by source: %v", err)
				}

				var output bytes.Buffer
				if err := tpl.Execute(&output, map[string]string{templateNameKey: "Grace"}); err != nil {
					t.Fatalf("execute embedded template: %v", err)
				}

				if output.String() != "welcome Grace\n" {
					t.Fatalf("output = %q, want %q", output.String(), "welcome Grace\n")
				}
			},
		},
		{
			name: "new template by source returns parse error",
			run: func(t *testing.T) {
				t.Helper()

				tpl, err := NewTemplateBySource(sourceTemplates, "missing", "testdata/templates/missing.tmpl")
				if err == nil {
					t.Fatal("expected parse error")
				}

				if tpl != nil {
					t.Fatal("template should be nil")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
