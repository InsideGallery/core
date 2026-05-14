# server/sse

Import path: `github.com/InsideGallery/core/server/sse`

`sse` provides basic Server-Sent Events formatting, streaming, and per-user
connection pooling for `net/http` handlers.

## Main APIs

- `Message` and `NewMessage(event, data...)`: event payload model.
- `FormatMSG(event, data...)`: renders SSE bytes using `event:` and `data:`
  lines.
- `Run(ch, w, r)`: streams messages from a channel to an `http.ResponseWriter`.
- `Pool`: tracks user ID to message channel mappings.
- `NewPool(bufferSize)`: creates a pool; non-positive sizes use the default.
- `Pool.Add`, `Remove`, `Send`, `SendToAll`, `StopAll`, `Connections`, and
  `GetAllConnectedUsers`: connection lifecycle and delivery helpers.
- `Pool.Handler`: HTTP handler that reads `ContextUserID` from request context.

## Usage

```go
pool := sse.NewPool(100)

http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), sse.ContextUserID, "user-1")
	pool.Handler(w, r.WithContext(ctx))
})

_ = pool.Send("user-1", sse.NewMessage("notice", "hello"))
```

## Operational Notes

`Run` requires a response writer that implements `http.Flusher`; otherwise it
returns `ErrResponseWriterIsNotFlusher`. It sets keep-alive, `text/event-stream`,
and wildcard CORS headers. `Pool.Handler` logs and returns without a connection
when `ContextUserID` is missing or not a string.
