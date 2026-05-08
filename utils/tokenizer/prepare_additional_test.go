package tokenizer

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestFromFileErrors(t *testing.T) {
	cases := []struct {
		name        string
		fs          fstest.MapFS
		file        string
		wantMessage string
	}{
		{
			name:        "missing file",
			fs:          fstest.MapFS{},
			file:        "missing.json",
			wantMessage: "open missing.json",
		},
		{
			name: "invalid json",
			fs: fstest.MapFS{
				"tokenizer.json": &fstest.MapFile{Data: []byte("{")},
			},
			file:        "tokenizer.json",
			wantMessage: "unexpected EOF",
		},
		{
			name: "unsupported model",
			fs: fstest.MapFS{
				"tokenizer.json": &fstest.MapFile{Data: []byte(`{"model":{"type":"Unknown"}}`)},
			},
			file:        "tokenizer.json",
			wantMessage: "creating Model failed",
		},
		{
			name: "unsupported normalizer",
			fs: fstest.MapFS{
				"tokenizer.json": &fstest.MapFile{Data: []byte(baseTokenizerConfig(`"normalizer":{"type":"Unknown"}`))},
			},
			file:        "tokenizer.json",
			wantMessage: "creating normalizer failed",
		},
		{
			name: "unsupported pre tokenizer",
			fs: fstest.MapFS{
				"tokenizer.json": &fstest.MapFile{Data: []byte(baseTokenizerConfig(`"pre_tokenizer":{"type":"Unknown"}`))},
			},
			file:        "tokenizer.json",
			wantMessage: "creating pre-tokenizer failed",
		},
		{
			name: "unsupported post processor",
			fs: fstest.MapFS{
				"tokenizer.json": &fstest.MapFile{Data: []byte(baseTokenizerConfig(`"post_processor":{"type":"Unknown"}`))},
			},
			file:        "tokenizer.json",
			wantMessage: "creating post-processor failed",
		},
		{
			name: "unsupported decoder",
			fs: fstest.MapFS{
				"tokenizer.json": &fstest.MapFile{Data: []byte(baseTokenizerConfig(`"decoder":{"type":"Unknown"}`))},
			},
			file:        "tokenizer.json",
			wantMessage: "creating decoder failed",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := FromFile(test.fs, test.file)
			if err == nil {
				t.Fatalf("FromFile() err = nil, tokenizer = %v", got)
			}

			if !strings.Contains(err.Error(), test.wantMessage) {
				t.Fatalf("FromFile() err = %q, want containing %q", err.Error(), test.wantMessage)
			}
		})
	}
}

func baseTokenizerConfig(extra string) string {
	const model = `"model":{"type":"BPE","vocab":{},"merges":[]}`

	return "{" + model + "," + extra + "}"
}
