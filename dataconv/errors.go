package dataconv

import "errors"

// All kind of errors
var (
	ErrWrongEncodeType = errors.New("wrong encode type")
	ErrWrongDecodeType = errors.New("wrong decode type")
)
