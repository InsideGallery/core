package sse

import (
	"net/http"
	"strings"
)

// Message message
type Message struct {
	Event string
	Text  []string
}

// NewMessage return new message
func NewMessage(event string, data ...string) Message {
	return Message{
		Event: event,
		Text:  data,
	}
}

// FormatMSG prepare bytes to send
func FormatMSG(event string, data ...string) []byte {
	eventPayload := strings.Join([]string{"event: ", event, "\n"}, "")
	for _, line := range data {
		eventPayload = strings.Join([]string{eventPayload, "data: ", line, "\n"}, "")
	}

	return []byte(eventPayload + "\n")
}

// Run listener of channel and prepare messages
func Run(ch chan Message, w http.ResponseWriter, r *http.Request) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return ErrResponseWriterIsNotFlusher
	}

	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				flusher.Flush()
				return nil
			}

			_, err := w.Write(FormatMSG(msg.Event, msg.Text...))
			if err != nil {
				return err
			}

			flusher.Flush()
		case <-r.Context().Done():
			flusher.Flush()

			return nil
		}
	}
}
