package template

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
func GetDefaultTemplateDir(prefix string) string {
	dir := os.Getenv(strings.Join([]string{DefaultTemplateDir, prefix}, ""))

	dir, err := filepath.Abs(dir)
	if err != nil {
		dir = ""
	}

	return dir
}

// SetDefaultTemplateDir set default template dir
func SetDefaultTemplateDir(prefix string, d string) {
	err := os.Setenv(strings.Join([]string{DefaultTemplateDir, prefix}, ""), d)
	if err != nil {
		slog.Default().Error("Can not et default template dir")
	}
}

// Template describe template
type Template struct {
	*template.Template
	Name string
}

// NewTemplate return new template
func NewTemplate(name, prefix string, fs fs.FS, files ...string) (*Template, error) {
	dir := GetDefaultTemplateDir(prefix)

	for i, f := range files {
		files[i] = strings.Join([]string{dir, "/", f}, "")
	}

	tmpl := template.Must(template.New(name).ParseFS(fs, files...))

	tmpl.Option("missingkey=error")

	return &Template{
		Name:     name,
		Template: tmpl,
	}, nil
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

// Execute execute current template and return parsed string
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
