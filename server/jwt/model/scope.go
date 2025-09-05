package model

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
)

const matchAllSymbol = "*"
const (
	AccessTypeRead  AccessType = "read"
	AccessTypeWrite AccessType = "write"
	AccessTypeAll   AccessType = "all"
	AccessTypeAny   AccessType = matchAllSymbol
)

const tokensTotalCount = 3

var MethodMap = map[string]AccessType{
	http.MethodGet:     AccessTypeRead,
	http.MethodHead:    AccessTypeRead,
	http.MethodPost:    AccessTypeWrite,
	http.MethodPut:     AccessTypeWrite,
	http.MethodPatch:   AccessTypeWrite,
	http.MethodDelete:  AccessTypeWrite,
	http.MethodConnect: AccessTypeRead,
	http.MethodOptions: AccessTypeRead,
	http.MethodTrace:   AccessTypeRead,
}

type (
	Scope struct {
		AccessType AccessType `json:"access_type"`
		Service    string     `json:"service"`
		Action     string     `json:"action"`
	}
	AccessType   string
	Scopes       []Scope
	ScopeChecker struct {
		ctx   context.Context
		scope Scope
	}
)

func (s Scope) String() string {
	if s.Service == "" && s.Action == "" {
		return string(s.AccessType)
	}

	return strings.Join([]string{string(s.AccessType), s.Service, s.Action}, ":")
}

func ScopeFrom(scopeName string) (Scope, error) {
	s, err := parseScope(scopeName)
	if err != nil {
		return Scope{}, err
	}

	/*
		    // TODO We have an error when aft_root can not execute search request
			validAccessType := s.AccessType != ""
			validService := s.Service != "" || strings.Contains(string(s.AccessType), matchAllSymbol) && s.Service == ""

			if !validAccessType {
				return Scope{}, errors.New("wrong access type")
			}
			if !validService {
				return Scope{}, errors.New("wrong service")
			}
	*/

	return s, nil
}

func parseScope(name string) (Scope, error) {
	tokens := strings.Split(name, ":")
	if len(tokens) == 0 {
		return Scope{}, fmt.Errorf("empty tokens")
	}

	var s Scope

	switch len(tokens) {
	case tokensTotalCount:
		s = Scope{AccessType(tokens[0]), tokens[1], tokens[2]}
	case tokensTotalCount - 1:
		s = Scope{AccessType(tokens[0]), tokens[1], ""}
	case 1:
		s = Scope{AccessType(tokens[0]), "", ""}
	}

	return s, nil
}

func NewScopeChecker(ctx context.Context, s string) (*ScopeChecker, error) {
	scope, err := ScopeFrom(s)
	if err != nil {
		return nil, err
	}

	return &ScopeChecker{ctx: ctx, scope: scope}, nil
}

func (s *ScopeChecker) IsAllowed(role UserRole, scopes Scopes, requestPasswordChange bool) bool {
	if requestPasswordChange {
		return s.scope.String() == "write:access:password"
	}

	if role.IsAnyRoot() {
		return true
	}

	slog.Default().Debug("Check scope allowed", "target_scope", s.scope.String(), "role", string(role), "scopes", scopes)

	for _, userScope := range scopes {
		if userScope.AccessType == AccessTypeAny || s.scope.String() == userScope.String() {
			return true
		}

		if userScope.AccessType != s.scope.AccessType && userScope.AccessType != AccessTypeAll {
			continue
		}

		if s.scope.Service == userScope.Service && s.scope.Action == userScope.Action {
			return true
		}

		serviceAsteriskPos := strings.Index(userScope.Service, matchAllSymbol)
		if hasAsterisk := serviceAsteriskPos >= 0; hasAsterisk {
			if s.scope.Service[:serviceAsteriskPos] == userScope.Service[:serviceAsteriskPos] {
				return true
			}
		} else if s.scope.Service != userScope.Service {
			continue
		}

		actionAsteriskPos := strings.Index(userScope.Action, matchAllSymbol)
		if hasAsterisk := actionAsteriskPos >= 0; hasAsterisk {
			if s.scope.Action[:actionAsteriskPos] == userScope.Action[:actionAsteriskPos] {
				return true
			}
		}
	}

	return false
}

func NormalizeScopes(ctx context.Context, scopes Scopes) (filtered Scopes, err error) {
	filtered = Scopes{}

	if len(scopes) == 0 {
		return
	}

	sort.Slice(scopes, func(i, j int) bool {
		return scopes[i].String() < scopes[j].String()
	})

	hasScope := func(objects Scopes, item Scope) bool {
		for _, obj := range objects {
			if obj.String() == item.String() {
				return true
			}
		}

		return false
	}
	removeScope := func(list Scopes, index int) Scopes {
		return append(list[:index], list[index+1:]...)
	}

	for i, scope := range scopes {
		// check opposite permission
		var foundOpposite bool

		switch scope.AccessType {
		case AccessTypeRead:
			opposite := scope
			opposite.AccessType = AccessTypeWrite

			if hasScope(scopes, opposite) {
				opposite.AccessType = AccessTypeAll

				if !hasScope(filtered, opposite) {
					filtered = append(filtered, opposite)
				}

				foundOpposite = true
			}
		case AccessTypeWrite:
			opposite := scope
			opposite.AccessType = AccessTypeRead

			if hasScope(scopes, opposite) {
				opposite.AccessType = AccessTypeAll
				if !hasScope(filtered, opposite) {
					filtered = append(filtered, opposite)
				}

				foundOpposite = true
			}
		}

		if foundOpposite {
			continue
		}

		scopeChecker, err := NewScopeChecker(ctx, scope.String())
		if err != nil {
			return nil, err
		}

		// check current in others
		scopesCopy := make(Scopes, len(scopes))
		copy(scopesCopy, scopes)

		exceptCurrent := removeScope(scopesCopy, i)

		allowed := scopeChecker.IsAllowed(UserRoleEmpty, exceptCurrent, false)
		if !allowed || len(filtered) == 0 {
			filtered = append(filtered, scope)
		}
	}

	return
}
