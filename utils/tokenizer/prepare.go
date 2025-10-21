package tokenizer

import (
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"

	"github.com/InsideGallery/core/multiproc/sync"
)

var (
	tk   *tokenizer.Tokenizer
	once sync.Once
)

func GetTokenizer(fs fs.FS, file string) (*tokenizer.Tokenizer, error) {
	err := once.Do(func() error {
		var err error

		if tk == nil {
			tk, err = FromFile(fs, file)
		}

		return err
	})

	return tk, err
}

func FromFile(fs fs.FS, file string) (*tokenizer.Tokenizer, error) {
	f, err := fs.Open(file)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(f)

	var config *tokenizer.Config

	err = dec.Decode(&config)
	if err != nil {
		return nil, err
	}

	model, err := pretrained.CreateModel(config)
	if err != nil {
		return nil, fmt.Errorf("creating Model failed: %w", err)
	}

	tk := tokenizer.NewTokenizer(model)

	// 2. Normalizer
	n, err := pretrained.CreateNormalizer(config.Normalizer)
	if err != nil {
		return nil, fmt.Errorf("creating normalizer failed: %w", err)
	}

	tk.WithNormalizer(n)

	// 3. PreTokenizer
	preTok, err := pretrained.CreatePreTokenizer(config.PreTokenizer)
	if err != nil {
		return nil, fmt.Errorf("creating pre-tokenizer failed: %w", err)
	}

	tk.WithPreTokenizer(preTok)

	// 4. PostProcessor
	postProcessor, err := pretrained.CreatePostProcessor(config.PostProcessor)
	if err != nil {
		return nil, fmt.Errorf("creating post-processor failed: %w", err)
	}

	tk.WithPostProcessor(postProcessor)

	// 5. Decoder
	decoder, err := pretrained.CreateDecoder(config.Decoder)
	if err != nil {
		return nil, fmt.Errorf("creating decoder failed: %w", err)
	}

	tk.WithDecoder(decoder)

	// 6. AddedVocabulary
	specialAddedTokens, addedTokens := pretrained.CreateAddedTokens(config.AddedTokens)
	if len(specialAddedTokens) > 0 {
		tk.AddSpecialTokens(specialAddedTokens)
	}

	if len(addedTokens) > 0 {
		tk.AddTokens(addedTokens)
	}

	// 7. TruncationParams
	truncParams, err := pretrained.CreateTruncationParams(config.Truncation)
	if err != nil {
		err = fmt.Errorf("creating truncation-params failed: %w", err)
		return nil, err
	}

	tk.WithTruncation(truncParams)

	// 8. PaddingParams
	paddingParams, err := pretrained.CreatePaddingParams(config.Padding)
	if err != nil {
		return nil, fmt.Errorf("creating padding-params failed: %w", err)
	}

	tk.WithPadding(paddingParams)

	return tk, nil
}
