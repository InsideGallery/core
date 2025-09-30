package tokenizer

import (
	"encoding/csv"
	"encoding/json"
	"strconv"
	"testing"
	"unicode/utf8"

	"github.com/InsideGallery/core/embedded"
	"github.com/InsideGallery/core/testutils"
	"github.com/InsideGallery/core/utils"
)

func TestGetTokenizer(t *testing.T) {
	p, err := GetTokenizer(embedded.GetFS(), "resources/tokenizer.json")
	testutils.Equal(t, err, nil)

	en, err := p.EncodeSingle("axbgref")
	testutils.Equal(t, err, nil)

	tokens := len(en.Tokens)
	testutils.Equal(t, tokens, 5)

	en, err = p.EncodeSingle("glasses")
	testutils.Equal(t, err, nil)

	tokens = len(en.Tokens)
	testutils.Equal(t, tokens, 2)
}

func TestEmails(t *testing.T) {
	file, err := embedded.GetFS().Open("resources/tests_for_emails.csv")
	testutils.Equal(t, err, nil)

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	testutils.Equal(t, err, nil)

	p, err := GetTokenizer(embedded.GetFS(), "resources/tokenizer.json")
	testutils.Equal(t, err, nil)

	for _, record := range records[1:] {
		record[0] = utils.EmailUserName(record[0])

		en, err := p.EncodeSingle(record[0])
		testutils.Equal(t, err, nil)

		expected, err := strconv.Atoi(record[1])
		testutils.Equal(t, err, nil)

		v, err := json.Marshal(en.Tokens)
		testutils.Equal(t, err, nil)

		count := utf8.RuneCountInString(record[0])

		result := int(float64(len(en.Tokens)) / float64(count) * 100)
		if result != expected {
			t.Fatalf("Go: %3d, Python %3d, Input Len: %d, Input: %s, Tokens: %s\n", result, expected, count, record[0], string(v))
		}
	}
}

func TestNames(t *testing.T) {
	file, err := embedded.GetFS().Open("resources/tests_for_names.csv")
	testutils.Equal(t, err, nil)

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	testutils.Equal(t, err, nil)

	p, err := GetTokenizer(embedded.GetFS(), "resources/tokenizer.json")
	testutils.Equal(t, err, nil)

	for _, record := range records[1:] {
		en, err := p.EncodeSingle(record[0])
		testutils.Equal(t, err, nil)

		expected, err := strconv.Atoi(record[1])
		testutils.Equal(t, err, nil)

		v, err := json.Marshal(en.Tokens)
		testutils.Equal(t, err, nil)

		count := utf8.RuneCountInString(record[0])

		result := int(float64(len(en.Tokens)) / float64(count) * 100)
		if result != expected {
			t.Fatalf("Go: %3d, Python %3d, Input Len: %d, Input: %s, Tokens: %s\n", result, expected, count, record[0], string(v))
		}
	}
}
