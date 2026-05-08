//go:generate mockgen -package mock -source=msg.go -destination=mock/msg.go
package interfaces

import "context"

type Msg interface {
	// Deprecated: use Subject.
	GetSubject() string
	IsReply() bool
	ReplyTo() string
	Copy(subject string) Msg
	SetHeader(key, value string)
	Respond([]byte) error
	// Deprecated: use Header.
	GetHeader() map[string][]string
	// Deprecated: use Data.
	GetData() []byte
	RespondMsg(Msg) error
}

type MsgHandler func(ctx context.Context, msg Msg) error

type subjectGetter interface {
	GetSubject() string
}

type subjectValuer interface {
	Subject() string
}

type headerMessage interface {
	Header() map[string][]string
}

type dataMessage interface {
	Data() []byte
}

// Subject returns value subject.
func Subject[T subjectGetter](value T) string {
	if valueSubject, ok := any(value).(subjectValuer); ok {
		return valueSubject.Subject()
	}

	return value.GetSubject()
}

// Header returns message header.
func Header(msg Msg) map[string][]string {
	if valueMsg, ok := msg.(headerMessage); ok {
		return valueMsg.Header()
	}

	return msg.GetHeader()
}

// Data returns message data.
func Data(msg Msg) []byte {
	if valueMsg, ok := msg.(dataMessage); ok {
		return valueMsg.Data()
	}

	return msg.GetData()
}
