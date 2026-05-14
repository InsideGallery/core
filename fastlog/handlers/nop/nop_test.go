package nop

import "testing"

func TestNoopWriterReportsBytesWritten(t *testing.T) {
	w := W{}

	n, err := w.Write([]byte("discarded"))
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if n != len("discarded") {
		t.Fatalf("Write() bytes = %d, want %d", n, len("discarded"))
	}
}

func TestNewReturnsNoopWriter(t *testing.T) {
	writer, opts, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if opts == nil {
		t.Fatal("expected handler options")
	}

	n, err := writer.Write([]byte("discarded"))
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if n != len("discarded") {
		t.Fatalf("Write() bytes = %d, want %d", n, len("discarded"))
	}
}
