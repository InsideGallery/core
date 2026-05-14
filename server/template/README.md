# server/template

Import path: `github.com/InsideGallery/core/server/template`

`template` wraps `html/template` parsing and execution for server-side HTML
templates, including embedded filesystems.

## Main APIs

- `Options`: explicit template name, directory, file list, and `fs.FS`.
- `NewTemplateFromOptions(options)`: creates a template from explicit options.
- `NewTemplateWithDir(name, dir, source, files...)`: parses files from an
  explicit base directory.
- `NewTemplateBySource(source, name, pattern)`: parses an embedded template
  pattern.
- `Template`: named template wrapper.
- `Engine`: stores templates by name.
- `Engine.Add`, `Exists`, and `Execute`: template registry and rendering helpers.
- `ErrExecuteTemplate` and `ErrNotFoundTemplate`: execution sentinel errors.

`GetDefaultTemplateDir`, `SetDefaultTemplateDir`, and `NewTemplate` are
deprecated environment-based helpers.

## Usage

```go
tmpl, err := coretemplate.NewTemplateWithDir("page.tmpl", "templates", fsys, "page.tmpl")
if err != nil {
	return err
}

engine := coretemplate.NewEngine()
engine.Add(tmpl)

html, err := engine.Execute("page.tmpl", map[string]string{"Name": "Ada"})
```

## Operational Notes

Templates are configured with `missingkey=error`. Prefer explicit directory
configuration through `Options` or `NewTemplateWithDir` instead of process-wide
environment state.
