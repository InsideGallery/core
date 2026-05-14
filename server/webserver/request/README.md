# server/webserver/request

Import path: `github.com/InsideGallery/core/server/webserver/request`

`request` contains Fiber request helpers for resolving client IP addresses.

## Main APIs

- `ErrAddressIsNotValid`: sentinel returned when an address cannot be parsed.
- `IsPrivateAddress(address)`: reports whether an IP address is loopback,
  private, or link-local unicast.
- `IPFromRequest(c fiber.Ctx)`: extracts the preferred client IP from
  `X-Forwarded-For`, `X-Real-Ip`, or Fiber's request IP.
- `IPStringFromRequest(c fiber.Ctx)`: returns the parsed IP string or falls back
  to `c.IP()`.

## Usage

```go
app.Get("/", func(c fiber.Ctx) error {
	clientIP := request.IPStringFromRequest(c)

	return c.SendString(clientIP)
})
```

## Operational Notes

`IPFromRequest` selects the first public address in `X-Forwarded-For`. If none is
usable, it falls back to `X-Real-Ip`; without forwarding headers it parses
`c.IP()` and strips a host port when present.
