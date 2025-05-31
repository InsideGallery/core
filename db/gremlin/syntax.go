package gremlin

import (
	"os"
)

const (
	SyntaxAerospike   = "aerospike"
	SyntaxNeptun      = "neptun"
	DefaultPropertyID = "id"
)

var (
	PropertyID interface{} = DefaultPropertyID
	Syntax                 = SyntaxAerospike
)

func init() {
	syntax := os.Getenv("GREMLIN_SYNTAX")
	if syntax != "" {
		Syntax = syntax
		switch Syntax {
		case SyntaxAerospike:
			PropertyID = DefaultPropertyID
		case SyntaxNeptun:
			PropertyID = DefaultPropertyID
		default:
			panic("syntax does not available: " + syntax)
		}
	}
}
