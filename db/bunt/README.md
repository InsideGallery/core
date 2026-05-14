# db/bunt

Import path: `github.com/InsideGallery/core/db/bunt`

Package `bunt` provides a small BuntDB wrapper for JSON-backed key/value storage. It opens a
`github.com/tidwall/buntdb` database, embeds the BuntDB handle in `Wrapper`, and adds `Get` and `Set`
helpers that unmarshal and marshal values as JSON.

## Main APIs

- `ConnectionConfig` configures the BuntDB filename. The default filename is `:memory:`.
- `GetConnectionConfigFromEnv(prefix)` reads `PREFIX_FILENAME`.
- `Open(config)` opens a `Wrapper` from explicit config.
- `OpenFromEnv(prefix)` reads config from an explicit environment prefix and opens the database.
- `Wrapper.Get(name, value)` reads a JSON value by key.
- `Wrapper.Set(name, value)` writes a JSON value by key.
- `GetConnection()` is the legacy default constructor. It uses the `DB` prefix and is deprecated in
  favor of explicit config.

## Usage

```go
package example

import "github.com/InsideGallery/core/db/bunt"

func saveProfile() (err error) {
	store, err := bunt.Open(&bunt.ConnectionConfig{Filename: ":memory:"})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := store.Close(); err == nil {
			err = closeErr
		}
	}()

	if err := store.Set("profile:1", map[string]string{"name": "Ada"}); err != nil {
		return err
	}

	var profile map[string]string
	return store.Get("profile:1", &profile)
}
```

## Configuration And Operations

Use `Open` when the application already owns configuration. Use `OpenFromEnv` when configuration should
come from environment variables. Call `Close` on the returned wrapper during shutdown. Missing keys and
storage errors are returned from BuntDB directly.
