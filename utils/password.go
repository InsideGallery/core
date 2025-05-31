package utils

import (
	"encoding/json"
)

// Password describe masked field
type Password string

// MarshalJSON block to show value in json
func (p Password) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// String return hide password value
func (p Password) String() string {
	return "********"
}

// Value return exactly password value
func (p Password) Value() string {
	return string(p)
}

type Str struct {
	Pass Password
}
