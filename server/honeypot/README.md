# server/honeypot

Import path: `github.com/InsideGallery/core/server/honeypot`

`honeypot` provides a small TCP listener that accepts connections, sends a fake
SSH banner, and logs input lines through `slog.Default()`.

## Main APIs

- `Honeypot(listenPort string) error`: listens on `":" + listenPort"` and
  handles accepted TCP connections until the listener fails.

## Usage

```go
if err := honeypot.Honeypot("2222"); err != nil {
	return err
}
```

The handler writes an `SSH-2.0-OpenSSH_7.9p1 Debian-10+deb10u2` banner and then
logs every scanned line with the remote address.

## Operational Notes

This package owns a blocking TCP accept loop and launches one goroutine per
accepted connection. It has no shutdown context; applications that need
coordinated shutdown should run it in a managed process or wrapper that can
close the listener externally.
