//go:generate mockgen -package mock -source=interface.go -destination=mock/client.go
package client

type Logger interface {
	Error(msg string, args ...any)
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
}

type StubLogger struct{}

func (l StubLogger) Error(_ string, _ ...any) {}
func (l StubLogger) Info(_ string, _ ...any)  {}
func (l StubLogger) Debug(_ string, _ ...any) {}
