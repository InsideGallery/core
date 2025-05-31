package sse

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestListener(t *testing.T) {
	ch := make(chan Message, 100)
	writer := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	var finish sync.WaitGroup
	finish.Add(1)
	go func() {
		defer finish.Done()
		err := Run(ch, writer, req)
		testutils.Equal(t, err, nil)
	}()

	ch <- NewMessage("message", "some text", "some more text")
	close(ch)
	finish.Wait()

	data, err := io.ReadAll(writer.Body)
	testutils.Equal(t, err, nil)
	testutils.Equal(t, string(data), "event: message\ndata: some text\ndata: some more text\n\n")
}
