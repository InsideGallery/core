package utils

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestSplitBetweenTokens(t *testing.T) {
	testcases := []struct {
		name      string
		data      string
		arguments []string
		result    []string
	}{
		{
			name:      "split_between_different_tokens",
			data:      `some_string_which_we_should_split@should_not_be_visible;must_be_present`,
			arguments: []string{"@", ";"},
			result:    []string{"some_string_which_we_should_split", "must_be_present"},
		},
		{
			name:      "split_between_same_token_tokens",
			data:      `some_string_which_we_should_split;should_not_be_visible;must_be_present`,
			arguments: []string{";", ";"},
			result:    []string{"some_string_which_we_should_split", "must_be_present"},
		},
		{
			name:      "split_between_single_token",
			data:      `some_string_which_we_should_split;should_not_be_visible;must_be_present`,
			arguments: []string{";"},
			result:    []string{"some_string_which_we_should_split", "must_be_present"},
		},
		{
			name:      "return_fist_part_for_single_token",
			data:      `some_string_which_we_should_split;should_not_be_visible`,
			arguments: []string{";"},
			result:    []string{"some_string_which_we_should_split"},
		},
		{
			name:      "return_income_string_if_no_arguments",
			data:      `some_string_which_we_should_split;should_be_also_visible`,
			arguments: []string{},
			result:    []string{"some_string_which_we_should_split;should_be_also_visible"},
		},
		{
			name:      "return_income_string_if_no_match",
			data:      `some_string_which_we_should_split;should_be_also_visible`,
			arguments: []string{"@"},
			result:    []string{"some_string_which_we_should_split;should_be_also_visible"},
		},
		{
			name:      "return_empty_for_empty_input",
			data:      ``,
			arguments: []string{"@"},
			result:    []string{},
		},
		{
			name:      "if_both_token_are_empty",
			data:      `some_string_which_we_should_split`,
			arguments: []string{"", ""},
			result:    []string{"some_string_which_we_should_split"},
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			result := SplitBetweenTokens(test.data, test.arguments...)
			testutils.Equal(t, result, test.result)
		})
	}
}

func TestSanitizeEmail(t *testing.T) {
	testcases := []struct {
		name   string
		email  string
		result string
	}{
		{
			name:   "email_with_tag",
			email:  "testemail+example@gmail.com",
			result: `testemail@gmail.com`,
		},
		{
			name:   "email_with_two_tags",
			email:  "testemail+exa+mple@gmail.com",
			result: `testemail@gmail.com`,
		},
		{
			name:   "user_name_with_tag_without_domain",
			email:  "testemail+exa",
			result: `testemail`,
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			result := SanitizeEmail(test.email)
			testutils.Equal(t, result, test.result)
		})
	}
}

func TestEmailDomain(t *testing.T) {
	testcases := []struct {
		name   string
		email  string
		domain string
	}{
		{
			name:   "valid_email",
			email:  "testemail@gmail.com",
			domain: `gmail.com`,
		},
		{
			name:   "empty_email",
			email:  "",
			domain: ``,
		},
		{
			name:   "only_domain",
			email:  "@gmail.com",
			domain: `gmail.com`,
		},
		{
			name:   "only_username",
			email:  "testmail@",
			domain: ``,
		},
		{
			name:   "at_not_present",
			email:  "test;mail.com",
			domain: `test;mail.com`,
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			domain := EmailDomain(test.email)
			testutils.Equal(t, domain, test.domain)
		})
	}
}

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
