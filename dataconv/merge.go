package dataconv

import (
	"dario.cat/mergo"
)

// MergeStruct help to fill all existing fields of record1 by record2
func MergeStruct(record1 interface{}, record2 interface{}) error {
	return mergo.Merge(record1, record2)
}
