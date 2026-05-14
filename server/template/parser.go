// Package template provides server-side HTML template rendering helpers.
//
// Prefer NewTemplateWithDir or NewTemplateFromOptions so template directory
// configuration is explicit.
package template //nolint:revive

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// DefaultTemplateDir have constant for template
const DefaultTemplateDir = "DEFAULT_TEMPLATE_DIR"

// templates related errors
var (
	ErrExecuteTemplate  = errors.New("error execute template")
	ErrNotFoundTemplate = errors.New("error not found template")
)

// GetDefaultTemplateDir return default template dir
//
// Deprecated: pass the directory to NewTemplateWithDir or NewTemplateFromOptions.
func GetDefaultTemplateDir(prefix string) string {
	dir := os.Getenv(strings.Join([]string{DefaultTemplateDir, prefix}, ""))

	dir, err := filepath.Abs(dir)
	if err != nil {
		dir = ""
	}

	return dir
}

// SetDefaultTemplateDir set default template dir
//
// Deprecated: pass the directory to NewTemplateWithDir or NewTemplateFromOptions.
func SetDefaultTemplateDir(prefix string, d string) {
	err := os.Setenv(strings.Join([]string{DefaultTemplateDir, prefix}, ""), d)
	if err != nil {
		slog.Default().Error("can not set default template dir", "err", err)
	}
}

// Template describe template
type Template struct {
	*template.Template
	Name string
}

// Options configures template parsing without environment state.
type Options struct {
	Name  string
	Dir   string
	Files []string
	FS    fs.FS
}

// NewTemplateFromOptions returns a new template from explicit options.
func NewTemplateFromOptions(options Options) (*Template, error) {
	return NewTemplateWithDir(options.Name, options.Dir, options.FS, options.Files...)
}

// NewTemplateWithDir returns a new template using an explicit base directory.
func NewTemplateWithDir(name, dir string, source fs.FS, files ...string) (*Template, error) {
	tmpl, err := template.New(name).ParseFS(source, templateFiles(dir, files)...)
	if err != nil {
		return nil, err
	}

	tmpl.Option("missingkey=error")

	return &Template{
		Name:     name,
		Template: tmpl,
	}, nil
}

// NewTemplate return new template
//
// Deprecated: use NewTemplateWithDir or NewTemplateFromOptions with explicit directory config.
func NewTemplate(name, prefix string, source fs.FS, files ...string) (*Template, error) {
	return NewTemplateWithDir(name, GetDefaultTemplateDir(prefix), source, files...)
}

// NewTemplateBySource return new template bu source
func NewTemplateBySource(source embed.FS, name, pattern string) (*Template, error) {
	tmpl, err := template.ParseFS(source, pattern)
	if err != nil {
		return nil, err
	}

	tmpl.Option("missingkey=error")

	return &Template{
		Name:     name,
		Template: tmpl,
	}, nil
}

// Engine type template engine
type Engine struct {
	pages map[string]*Template
}

// NewEngine return engine
func NewEngine() *Engine {
	return &Engine{
		pages: make(map[string]*Template),
	}
}

// Add add template
func (e *Engine) Add(t *Template) {
	e.pages[t.Name] = t
}

// Exists return true if page exists
func (e *Engine) Exists(name string) bool {
	_, exists := e.pages[name]
	return exists
}

// Execute execute current template and return parsed strings
func (e *Engine) Execute(name string, data interface{}) ([]byte, error) {
	var (
		err error
		tpl bytes.Buffer
	)

	if t, ok := e.pages[name]; ok {
		err = t.Execute(&tpl, data)
		if err != nil {
			err = errors.Wrap(ErrExecuteTemplate, err.Error())
		}

		return tpl.Bytes(), err
	}

	return []byte{}, ErrNotFoundTemplate
}

func templateFiles(dir string, files []string) []string {
	resolved := append([]string(nil), files...)
	if dir == "" {
		return resolved
	}

	for i, file := range resolved {
		resolved[i] = strings.Join([]string{dir, "/", file}, "")
	}

	return resolved
}
