package gremlin

import "testing"

func TestSyntaxState(t *testing.T) {
	cases := []struct {
		name         string
		syntax       string
		wantSyntax   string
		wantProperty interface{}
	}{
		{
			name:         "aerospike",
			syntax:       SyntaxAerospike,
			wantSyntax:   SyntaxAerospike,
			wantProperty: DefaultPropertyID,
		},
		{
			name:         "neptun",
			syntax:       SyntaxNeptun,
			wantSyntax:   SyntaxNeptun,
			wantProperty: DefaultPropertyID,
		},
		{
			name:         "unknown falls back",
			syntax:       "unknown",
			wantSyntax:   SyntaxAerospike,
			wantProperty: DefaultPropertyID,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := NewSyntaxState(test.syntax)

			if got.Syntax != test.wantSyntax {
				t.Fatalf("Syntax = %q, want %q", got.Syntax, test.wantSyntax)
			}

			if got.PropertyID != test.wantProperty {
				t.Fatalf("PropertyID = %v, want %v", got.PropertyID, test.wantProperty)
			}
		})
	}
}

func TestSyntaxStateFromConfig(t *testing.T) {
	cases := []struct {
		name       string
		config     SyntaxConfig
		wantSyntax string
	}{
		{
			name: "config syntax",
			config: SyntaxConfig{
				Syntax: SyntaxNeptun,
			},
			wantSyntax: SyntaxNeptun,
		},
		{
			name:       "empty config falls back",
			wantSyntax: SyntaxAerospike,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := SyntaxStateFromConfig(test.config)
			if got.Syntax != test.wantSyntax {
				t.Fatalf("Syntax = %q, want %q", got.Syntax, test.wantSyntax)
			}
		})
	}
}

func TestSetup(t *testing.T) {
	cases := []struct {
		name  string
		state SyntaxState
	}{
		{
			name:  "applies explicit syntax state",
			state: NewSyntaxState(SyntaxNeptun),
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			restoreOriginal := InstallSyntaxState(NewSyntaxState(SyntaxAerospike))
			defer restoreOriginal()

			Setup(test.state)
			Setup(test.state)

			got := CurrentSyntaxState()
			if got.Syntax != test.state.Syntax {
				t.Fatalf("Syntax = %q, want %q", got.Syntax, test.state.Syntax)
			}

			if got.PropertyID != test.state.PropertyID {
				t.Fatalf("PropertyID = %v, want %v", got.PropertyID, test.state.PropertyID)
			}
		})
	}
}

func TestInstallSyntaxState(t *testing.T) {
	cases := []struct {
		name  string
		state SyntaxState
	}{
		{
			name:  "installs and restores legacy state",
			state: NewSyntaxState(SyntaxNeptun),
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			restoreOriginal := InstallSyntaxState(NewSyntaxState(SyntaxAerospike))
			defer restoreOriginal()

			restore := InstallSyntaxState(test.state)
			if got := CurrentSyntaxState(); got.Syntax != test.state.Syntax {
				t.Fatalf("Syntax = %q, want %q", got.Syntax, test.state.Syntax)
			}

			restore()

			if got := CurrentSyntaxState(); got.Syntax != SyntaxAerospike {
				t.Fatalf("restored Syntax = %q, want %q", got.Syntax, SyntaxAerospike)
			}
		})
	}
}

func TestSyntaxStateFromEnv(t *testing.T) {
	cases := []struct {
		name       string
		env        string
		wantSyntax string
	}{
		{
			name:       "env syntax",
			env:        SyntaxNeptun,
			wantSyntax: SyntaxNeptun,
		},
		{
			name:       "empty env falls back",
			wantSyntax: SyntaxAerospike,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("GREMLIN_SYNTAX", test.env)

			got := SyntaxStateFromEnv()
			if got.Syntax != test.wantSyntax {
				t.Fatalf("Syntax = %q, want %q", got.Syntax, test.wantSyntax)
			}
		})
	}
}
