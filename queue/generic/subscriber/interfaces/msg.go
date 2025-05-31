//go:generate mockgen -package mock -source=msg.go -destination=mock/msg.go
package interfaces

import "context"

type Msg interface {
	GetSubject() string
	IsReply() bool
	ReplyTo() string
	Copy(subject string) Msg
	SetHeader(key, value string)
	Respond([]byte) error
	GetHeader() map[string][]string
	GetData() []byte
	RespondMsg(Msg) error
}

type MsgHandler func(ctx context.Context, msg Msg) error
