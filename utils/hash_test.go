package utils

import (
	"fmt"
	"slices"
	"testing"

	"github.com/dgryski/go-farm"
	"github.com/dgryski/go-minhash"
	"github.com/dgryski/go-spooky"
	"github.com/sugarme/tokenizer/normalizer"

	"github.com/InsideGallery/core/embedded"
	"github.com/InsideGallery/core/testutils"
	"github.com/InsideGallery/core/utils/tokenizer"
)

type Token struct {
	ID     int
	Str    string
	Values []string
}

var storage = map[int]*Token{}

func AddValue(str string) error {
	p, err := tokenizer.GetTokenizer(embedded.GetFS(), "resources/tokenizer.json")
	if err != nil {
		return err
	}

	en, err := p.EncodeSingle(str)
	if err != nil {
		return err
	}

	for i, v := range en.Ids {
		tk, ok := storage[v]
		if !ok {
			tk = &Token{
				ID:     v,
				Str:    en.Tokens[i],
				Values: []string{str},
			}
			storage[v] = tk
		} else {
			tk.Values = append(tk.Values, str)
		}
	}

	return nil
}

func Normalize(str string) (string, error) {
	n := normalizer.NewBertNormalizer(true, true, true, true)

	res, err := n.Normalize(normalizer.NewNormalizedFrom(CommonString(SanitizeEmail(str))))
	if err != nil {
		return "", err
	}

	return res.GetNormalized(), nil
}

func TestMH(t *testing.T) {
	v, err := Normalize("test.èmcop")
	testutils.Equal(t, err, nil)
	fmt.Println(v)

	v2, err := Normalize("testèmCAP")
	testutils.Equal(t, err, nil)
	fmt.Println(v2)

	h1 := spooky.Hash64
	h2 := farm.Hash64
	h := minhash.NewMinWise(h1, h2, 1000)
	h.Push([]byte(v))
	fmt.Println(h.Signature())
	fmt.Println(slices.Min(h.Signature()))

	h = minhash.NewMinWise(h1, h2, 1000)
	h.Push([]byte(v2))
	fmt.Println(h.Signature())
	fmt.Println(slices.Min(h.Signature()))
}

func TestCRC32(t *testing.T) {
	testutils.Equal(t, CRC32("test1"), uint32(1409163093))
	testutils.Equal(t, CRC32("test2"), uint32(1085205665))
	testutils.Equal(t, CRC32("true"), uint32(151551613))
	testutils.Equal(t, CRC32("false"), uint32(118305666))
}

func TestCRC16(t *testing.T) {
	testutils.Equal(t, CRC16("test1"), uint16(4768))
	testutils.Equal(t, CRC16("test2"), uint16(8899))
	testutils.Equal(t, CRC16("true"), uint16(62787))
	testutils.Equal(t, CRC16("false"), uint16(29756))
}
