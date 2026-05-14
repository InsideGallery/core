package model

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUser(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "new user sets defaults",
			run: func(t *testing.T) {
				t.Helper()

				orgID := primitive.NewObjectID()
				user := NewUser(orgID, "root", true)

				if user.ID.IsZero() {
					t.Fatal("user id is zero")
				}

				if user.OrgID != orgID {
					t.Fatalf("org id = %s, want %s", user.OrgID.Hex(), orgID.Hex())
				}

				if user.Role != UserRoleRoot || user.Name != "root" || !user.ChangePassword {
					t.Fatalf("user defaults = %#v", user)
				}
			},
		},
		{
			name: "password hash validates original password only",
			run: func(t *testing.T) {
				t.Helper()

				user := &User{}
				if err := user.SetPassword("secret"); err != nil {
					t.Fatalf("set password: %v", err)
				}

				if len(user.PasswordHash) == 0 {
					t.Fatal("password hash is empty")
				}

				if !user.IsPasswordValid("secret") {
					t.Fatal("password should be valid")
				}

				if user.IsPasswordValid("other") {
					t.Fatal("different password should be invalid")
				}
			},
		},
		{
			name: "role strings returns raw value",
			run: func(t *testing.T) {
				t.Helper()

				if got := UserRoleManager.String(); got != "manager" {
					t.Fatalf("role strings = %q, want manager", got)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
