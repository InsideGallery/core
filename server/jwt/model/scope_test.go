//go:build unit
// +build unit

package model

import (
	"context"
	"reflect"
	"testing"

	guuid "github.com/google/uuid"

	"github.com/InsideGallery/core/testutils"
)

func TestScopeChecker_IsAllowed(t *testing.T) {
	ctx := context.TODO()

	tests := []struct {
		name           string
		scope          string
		userRole       UserRole
		userScopes     []string
		changePassword bool // todo: use
		want           bool
	}{
		// exact
		{
			name:  "success (write by write - exact)",
			scope: "write:service:action",
			userScopes: []string{
				"write:service1:action1",
				"write:service:action",
				"write:service3:action2",
			},
			want: true,
		},
		{
			name:  "success (read by read - exact)",
			scope: "read:service:action",
			userScopes: []string{
				"read:service1:action1",
				"read:service:action",
				"read:service3:action2",
			},
			want: true,
		},
		{
			name:  "success (read by read - exact)",
			scope: "read:service:action",
			userScopes: []string{
				"write:service1:action1",
				"read:service:action",
				"read:service3:action2",
			},
			want: true,
		},
		{
			name:  "success (read by all - exact)",
			scope: "read:service:action",
			userScopes: []string{
				"read:service1:action1",
				"all:service:action",
				"all:service3:action2",
			},
			want: true,
		},
		{
			name:  "success (write by all - exact)",
			scope: "write:service:action",
			userScopes: []string{
				"all:service1:action1",
				"all:service:action",
				"all:service3:action2",
			},
			want: true,
		},
		// match
		{
			name:  "success (write by write - match)",
			scope: "write:service:action",
			userScopes: []string{
				"write:service1:action1",
				"write:service:action*",
				"write:service3:action2",
			},
			want: true,
		},
		{
			name:  "success (write by write - match)",
			scope: "write:service:action",
			userScopes: []string{
				"write:service:act*",
			},
			want: true,
		},
		{
			name:  "success (read by read - match)",
			scope: "read:service:something/1/2/3",
			userScopes: []string{
				"read:service:something/1/*",
			},
			want: true,
		},
		{
			name:  "success (write by write - match)",
			scope: "write:service:action2",
			userScopes: []string{
				"write:service1:action1",
				"write:service:action*",
				"write:service3:action2",
			},
			want: true,
		},
		{
			name:  "success (read by read - match)",
			scope: "read:service:action",
			userScopes: []string{
				"read:1service:action1",
				"read:service:*",
				"read:3service:action2",
			},
			want: true,
		},
		{
			name:  "success (read by read - match)",
			scope: "read:service:action",
			userScopes: []string{
				"read:1service:action1",
				"read:service*",
				"read:3service:action2",
			},
			want: true,
		},
		{
			name:  "success (read by read - match)",
			scope: "read:service:action",
			userScopes: []string{
				"read:1service:action1",
				"read:service:*",
				"read:3service:action2",
			},
			want: true,
		},
		{
			name:  "success (read by all - match)",
			scope: "read:service:action",
			userScopes: []string{
				"all:service1:action1",
				"all:service:*",
				"all:service3:action2",
			},
			want: true,
		},
		{
			name:  "success (write by all - match)",
			scope: "write:service:action",
			userScopes: []string{
				"all:service1:action1",
				"all:service:action*",
				"all:service3:action2",
			},
			want: true,
		},
		{
			name:  "success (root user scopes)",
			scope: "read:" + guuid.NewString()[:4] + ":" + guuid.NewString()[:4], // any scope
			userScopes: []string{
				"*",
			},
			want: true,
		},
		{
			name:  "success (root user scopes)",
			scope: "write:" + guuid.NewString()[:4] + ":" + guuid.NewString()[:4], // any scope
			userScopes: []string{
				"*",
			},
			want: true,
		},
		{
			name:  "success (root user scopes)",
			scope: "write:service:" + guuid.NewString(), // any scope
			userScopes: []string{
				"all:*",
			},
			want: true,
		},
		{
			name:       "success (root user role)",
			scope:      "write:service:" + guuid.NewString(), // any scope
			userRole:   UserRoleRoot,
			userScopes: []string{},
			want:       true,
		},
		// failed
		{
			name:  "failed (write by read)",
			scope: "write:service:action",
			userScopes: []string{
				"read:1service:action1",
				"read:service:action",
				"read:3service:action2",
			},
			want: false,
		},
		{
			name:  "failed (read by write)",
			scope: "read:service:action",
			userScopes: []string{
				"write:1service:action1",
				"write:service:action",
				"write:3service:action2",
			},
			want: false,
		},
		{
			name:  "failed (write by read - match)",
			scope: "write:service:action",
			userScopes: []string{
				"read:service1:action1",
				"read:*",
				"read:service3:action2",
			},
			want: false,
		},
		{
			name:       "empty user scopes",
			scope:      "write:service:action",
			userScopes: []string{},
			want:       false,
		},
		{
			name:  "failed (read by read - match)",
			scope: "read:service:action",
			userScopes: []string{
				"read:serv:*",
			},
			want: false,
		},
		// request a password change
		{
			name:           "request a password change - success",
			scope:          "write:access:password",
			userRole:       UserRoleManager,
			userScopes:     nil,
			changePassword: true,
			want:           true,
		},
		{
			name:           "request a password change - failed",
			scope:          "write:access:user",
			userRole:       UserRoleRoot,
			userScopes:     nil,
			changePassword: true,
			want:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewScopeChecker(ctx, tt.scope)
			testutils.Equal(t, err, nil)

			u := &User{
				Scopes:         toScopes(tt.userScopes),
				Role:           tt.userRole,
				ChangePassword: tt.changePassword,
			}
			if got := s.IsAllowed(u.Role, u.Scopes, u.ChangePassword); got != tt.want {
				t.Errorf("IsAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScopes_Normalize(t *testing.T) {
	ctx := context.TODO()

	tests := []struct {
		name       string
		userScopes []string
		wantScopes []string
	}{
		{
			name: "filtered - exact",
			userScopes: []string{
				"write:service:action",
				"all:service:action",
			},
			wantScopes: []string{
				"all:service:action",
			},
		},
		{
			name: "filtered - exact",
			userScopes: []string{
				"read:service:action",
				"all:service:action",
			},
			wantScopes: []string{
				"all:service:action",
			},
		},
		{
			name: "filtered - write match",
			userScopes: []string{
				"write:*",
				"write:service:*",
				"write:service:action*",
			},
			wantScopes: []string{
				"write:*",
			},
		},
		{
			name: "filtered - write and all to all",
			userScopes: []string{
				"write:service:*",
				"all:service:*",
			},
			wantScopes: []string{
				"all:service:*",
			},
		},
		{
			name: "filtered - write and all to all",
			userScopes: []string{
				"write:service:*",
				"all:*",
			},
			wantScopes: []string{
				"all:*",
			},
		},
		{
			name: "filtered - read and write to all",
			userScopes: []string{
				"read:service:*",
				"write:service:*",
			},
			wantScopes: []string{
				"all:service:*",
			},
		},
		{
			name: "filtered - all to all",
			userScopes: []string{
				"all:service:action",
				"all:service:*",
			},
			wantScopes: []string{
				"all:service:*",
			},
		},
		{
			name: "not filtered - exact action",
			userScopes: []string{
				"write:service1:action",
				"read:service2:action",
			},
			wantScopes: []string{
				"read:service2:action",
				"write:service1:action",
			},
		},
		{
			name: "not filtered - exact service",
			userScopes: []string{
				"write:service:action1",
				"read:service:action2",
			},
			wantScopes: []string{
				"read:service:action2",
				"write:service:action1",
			},
		},
		{
			name:       "nil scopes",
			userScopes: nil,
			wantScopes: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scopes := toScopes(tt.userScopes)
			actual, err := NormalizeScopes(ctx, scopes)
			testutils.Equal(t, err, nil)

			if !reflect.DeepEqual(actual, toScopes(tt.wantScopes)) {
				t.Errorf("NormalizeScopes() = %v\n\t\twant %v", actual, toScopes(tt.wantScopes))
			}
		})
	}
}

func toScopes(scopes []string) (list Scopes) {
	list = Scopes{}
	for _, s := range scopes {
		from, _ := ScopeFrom(s)

		list = append(list, from)
	}

	return
}

func TestScopeFrom(t *testing.T) {
	tests := []struct {
		name    string
		scope   string
		want    Scope
		wantErr error
	}{
		{
			name:  "normal",
			scope: "read:service:action",
			want:  Scope{AccessType: AccessTypeRead, Service: "service", Action: "action"},
		},
		{
			name:  "normal",
			scope: "read:service*",
			want:  Scope{AccessType: AccessTypeRead, Service: "service*", Action: ""},
		},
		{
			name:  "normal",
			scope: "read:service:*",
			want:  Scope{AccessType: AccessTypeRead, Service: "service", Action: "*"},
		},
		{
			name:  "normal",
			scope: "read:*",
			want:  Scope{AccessType: AccessTypeRead, Service: "*", Action: ""},
		},
		{
			name:  "normal",
			scope: "*",
			want:  Scope{AccessType: "*", Service: "", Action: ""},
		},
		{
			name:  "normal",
			scope: "read:service:",
			want:  Scope{AccessType: "read", Service: "service", Action: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ScopeFrom(tt.scope)
			testutils.Equal(t, err, tt.wantErr)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScopeFrom() got = %v, want %v", got, tt.want)
			}
		})
	}
}
