package metrics

import (
	"errors"

	"go.opentelemetry.io/otel/attribute"
)

const (
	attributeSubject = "subject"
)

var ErrWrongCountOfArguments = errors.New("error wrong count of arguments")

func PrepareAttributes(keyValues ...string) ([]attribute.KeyValue, error) {
	l := len(keyValues)
	if l == 0 {
		return nil, nil
	}

	if l%2 != 0 {
		return nil, ErrWrongCountOfArguments
	}

	var attributes []attribute.KeyValue

	for i := 0; i < len(keyValues)-1; i += 2 {
		key, value := keyValues[i], keyValues[i+1]
		attributes = append(attributes, attribute.String(key, value))
	}

	return attributes, nil
}
