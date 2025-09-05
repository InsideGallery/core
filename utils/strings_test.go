package utils

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestSplitByChunks(t *testing.T) {
	chunks := SplitByChunks("teststring", 3)
	testutils.Equal(t, chunks, []string{"tes", "tst", "rin", "g"})
}

func TestGetTinyID(t *testing.T) {
	shortID, err := GetTinyID()
	testutils.Equal(t, err, nil)
	shortID2, err := GetTinyID()
	testutils.Equal(t, err, nil)
	testutils.NotEqual(t, shortID, shortID2)
}

func TestRandStringBytes(t *testing.T) {
	val := RandStringBytes(1)
	testutils.Equal(t, len(val), 1)
}

func TestHashName(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		result string
	}{
		{
			name:   "test",
			input:  "test",
			result: "t9b06",
		},
		{
			name:   "weavers",
			input:  "weavers",
			result: "w0709",
		},
		{
			name:   "InsideGallery",
			input:  "InsideGallery",
			result: "i53a5",
		},
	}

	for _, tst := range testcases {
		t.Run(tst.name, func(t *testing.T) {
			val := HashName(tst.input)
			testutils.Equal(t, val, tst.result)
		})
	}
}

func TestByteSliceToString(t *testing.T) {
	testcases := map[string]struct {
		bytes []byte
		out   string
	}{
		"inStrs:empty":     {bytes: []byte{}, out: ""},
		"inStrs:nil":       {bytes: nil, out: ""},
		"inStrs:non_empty": {bytes: []byte{72, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100, 33}, out: "Hello world!"},
	}

	for k, c := range testcases {
		c := c

		t.Run(k, func(t *testing.T) {
			str := ByteSliceToString(c.bytes)
			testutils.Equal(t, str, c.out)
		})
	}
}

func TestByteSliceToStringNative(t *testing.T) {
	testcases := map[string]struct {
		bytes []byte
		out   string
	}{
		"inStrs:empty":     {bytes: []byte{}, out: ""},
		"inStrs:nil":       {bytes: nil, out: ""},
		"inStrs:non_empty": {bytes: []byte{72, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100, 33}, out: "Hello world!"},
	}

	for k, c := range testcases {
		c := c

		t.Run(k, func(t *testing.T) {
			str := ByteSliceToString(c.bytes)
			testutils.Equal(t, str, c.out)
		})
	}
}

func TestBetween(t *testing.T) {
	testcases := map[string]struct {
		data   string
		keys   []string
		result string
	}{
		"empty_data": {
			data:   "",
			keys:   []string{"[RESULT]"},
			result: "",
		},
		"not_key_in_data": {
			data:   "Some text",
			keys:   []string{"[RESULT]"},
			result: "",
		},
		"single_key_in_data": {
			data:   "Some [RESULT]text",
			keys:   []string{"[RESULT]"},
			result: "",
		},
		"key_is_present": {
			data:   "Some [RESULT]text[RESULT]",
			keys:   []string{"[RESULT]"},
			result: "text",
		},
		"trip_space_in_result": {
			data:   "Some [RESULT] text \n[RESULT]",
			keys:   []string{"[RESULT]"},
			result: "text",
		},
		"no_key": {
			data:   "Some [RESULT] text \n[RESULT]",
			keys:   []string{},
			result: "",
		},
		"between_two_keys": {
			data:   "Some < text\n > \n",
			keys:   []string{"<", ">"},
			result: "text",
		},
	}

	for name, test := range testcases {
		test := test

		t.Run(name, func(t *testing.T) {
			result := Between(test.data, test.keys...)
			testutils.Equal(t, test.result, result)
		})
	}
}

func TestMaskField(t *testing.T) {
	testcases := []struct {
		name              string
		str               string
		keepUnmaskedFront int
		keepUnmaskedEnd   int
		expected          string
	}{
		{
			name:              "mask long string",
			str:               "secret-string-here",
			keepUnmaskedFront: 2,
			keepUnmaskedEnd:   3,
			expected:          "se*************ere",
		},
		{
			name:              "mask short string",
			str:               "sec",
			keepUnmaskedFront: 2,
			keepUnmaskedEnd:   3,
			expected:          "***",
		},
		{
			name:              "mask medium string",
			str:               "secret",
			keepUnmaskedFront: 2,
			keepUnmaskedEnd:   3,
			expected:          "******",
		},
		{
			name:              "mask minimum to have show string",
			str:               "secret12345",
			keepUnmaskedFront: 2,
			keepUnmaskedEnd:   3,
			expected:          "se******345",
		},
		{
			name:              "mask without keep",
			str:               "secret12345",
			keepUnmaskedFront: 0,
			keepUnmaskedEnd:   0,
			expected:          "***********",
		},
		{
			name:              "mask front only",
			str:               "sec",
			keepUnmaskedFront: 1,
			keepUnmaskedEnd:   0,
			expected:          "s**",
		},
		{
			name:              "mask back only",
			str:               "sec",
			keepUnmaskedFront: 0,
			keepUnmaskedEnd:   1,
			expected:          "**c",
		},
		{
			name:              "empty str",
			str:               "",
			keepUnmaskedFront: 2,
			keepUnmaskedEnd:   3,
			expected:          "",
		},
	}

	for _, test := range testcases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result := MaskField(test.str, test.keepUnmaskedFront, test.keepUnmaskedEnd)
			testutils.Equal(t, result, test.expected)
		})
	}
}
