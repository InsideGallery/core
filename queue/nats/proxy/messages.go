//go:generate easyjson -all messages.go
package proxy

type Subscribe struct {
	InstanceID string
	Subject    string
}

type Unsubscribe struct {
	InstanceID string
	Subject    string
}
