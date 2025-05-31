package elasticsearch

import "errors"

var (
	ErrWrongResponse         = errors.New("wrong response")
	ErrWrongCountOfArguments = errors.New("error count of arguments")
)
