package proxy

import (
	"strings"
)

const (
	proxyQueue = "proxy"

	suffixSubscribe   = "subscribe"
	suffixUnsubscribe = "unsubscribe"

	HeaderID = "id"

	formatIntBase = 10
	separator     = "."

	PingMsg = "PING"
	PongMsg = "PONG"
)

func GetSubscribeSubject(subject string) string {
	return strings.Join([]string{subject, suffixSubscribe}, separator)
}

func GetUnsubscribeSubject(subject string) string {
	return strings.Join([]string{subject, suffixUnsubscribe}, separator)
}

func GetPodSubject(subject string, instance string) string {
	return strings.Join([]string{subject, instance}, separator)
}
