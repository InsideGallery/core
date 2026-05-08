package gremlin

import (
	"sync"
)

const (
	SyntaxAerospike   = "aerospike"
	SyntaxNeptun      = "neptun"
	DefaultPropertyID = "id"
)

var (
	syntaxMu sync.RWMutex //nolint:gochecknoglobals // guards legacy syntax globals

	// PropertyID is the legacy package-level Gremlin property identifier.
	//
	// Deprecated: use NewSyntaxState and pass SyntaxState explicitly.
	PropertyID interface{} = DefaultPropertyID

	// Syntax is the legacy package-level Gremlin syntax selector.
	//
	// Deprecated: use NewSyntaxState and pass SyntaxState explicitly.
	Syntax = SyntaxAerospike
)

func init() {
	// Deprecated: call Setup with explicit SyntaxState dependencies.
	Setup(SyntaxStateFromEnv())
}

// SyntaxState is an explicit Gremlin syntax configuration.
type SyntaxState struct {
	Syntax     string
	PropertyID interface{}
}

// NewSyntaxState creates a normalized Gremlin syntax state.
func NewSyntaxState(syntax string) SyntaxState {
	switch syntax {
	case SyntaxAerospike, SyntaxNeptun:
		return SyntaxState{
			Syntax:     syntax,
			PropertyID: DefaultPropertyID,
		}
	default:
		return SyntaxState{
			Syntax:     SyntaxAerospike,
			PropertyID: DefaultPropertyID,
		}
	}
}

// SyntaxStateFromConfig creates a Gremlin syntax state from explicit config.
func SyntaxStateFromConfig(config SyntaxConfig) SyntaxState {
	return NewSyntaxState(config.Syntax)
}

// SyntaxStateFromEnv reads Gremlin syntax config for legacy compatibility.
//
// Deprecated: use GetSyntaxConfigFromEnv and SyntaxStateFromConfig in the application composition root.
func SyntaxStateFromEnv() SyntaxState {
	config, err := GetSyntaxConfigFromEnv()
	if err != nil {
		return NewSyntaxState("")
	}

	return SyntaxStateFromConfig(*config)
}

// Setup applies a Gremlin syntax state to the legacy package-level globals.
func Setup(state SyntaxState) {
	ApplySyntaxState(state)
}

// CurrentSyntaxState returns the current legacy package-level syntax state.
func CurrentSyntaxState() SyntaxState {
	syntaxMu.RLock()
	defer syntaxMu.RUnlock()

	return SyntaxState{
		Syntax:     Syntax,
		PropertyID: PropertyID,
	}
}

// ApplySyntaxState updates the legacy package-level syntax state.
//
// Deprecated: pass SyntaxState explicitly instead of mutating package-level state.
func ApplySyntaxState(state SyntaxState) {
	syntaxMu.Lock()
	defer syntaxMu.Unlock()

	Syntax = state.Syntax
	PropertyID = state.PropertyID
}

// InstallSyntaxState applies a legacy package-level syntax state with a restore path.
//
// Deprecated: pass SyntaxState explicitly instead of mutating package-level state.
func InstallSyntaxState(state SyntaxState) func() {
	previous := CurrentSyntaxState()

	ApplySyntaxState(state)

	return func() {
		ApplySyntaxState(previous)
	}
}
