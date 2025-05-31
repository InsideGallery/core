package dataconv

import "errors"

// All kind of errors
var (
	ErrWrongEncodeType = errors.New("error wrong encode type")
	ErrWrongDecodeType = errors.New("error wrong decode type")
)
