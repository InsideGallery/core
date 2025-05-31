//go:build unit
// +build unit

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Role(t *testing.T) {
	tests := []struct {
		name        string
		role        UserRole
		isRoot      bool
		isAFTRoot   bool
		isAnyRoot   bool
		isManager   bool
		isEmptyRole bool
	}{
		{"root", UserRoleRoot, true, false, true, false, false},
		{"Super root", UserRoleSuperRoot, false, true, true, false, false},
		{"manager", UserRoleManager, false, false, false, true, false},
		{"empty", UserRoleEmpty, false, false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := tt.role

			assert.Equal(t, tt.isRoot, role.IsRoot())
			assert.Equal(t, tt.isAFTRoot, role.IsSuperRoot())
			assert.Equal(t, tt.isAnyRoot, role.IsAnyRoot())
			assert.Equal(t, tt.isManager, role.IsManager())
			assert.Equal(t, tt.isEmptyRole, role.IsEmptyRole())
		})
	}
}
