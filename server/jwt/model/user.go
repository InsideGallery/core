package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

const DefaultUser = "root"

type User struct {
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	Role           UserRole           `bson:"role" json:"role"`
	Name           string             `bson:"name" json:"name"`
	Scopes         Scopes             `bson:"scopes" json:"scopes"`
	PasswordHash   []byte             `bson:"password_hash" json:"password_hash"`
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	OrgID          primitive.ObjectID `bson:"org_id" json:"org_id"`
	ChangePassword bool               `bson:"change_password" json:"change_password"`
}

func NewUser(orgID primitive.ObjectID, name string, needChangePassword bool) *User {
	return &User{
		ID:             primitive.NewObjectID(),
		OrgID:          orgID,
		Role:           UserRoleRoot,
		Name:           name,
		ChangePassword: needChangePassword,
	}
}

func (u *User) SetPassword(s string) (err error) {
	u.PasswordHash, err = u.hashPassword([]byte(s))

	return
}

func (u *User) hashPassword(data []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(data, bcrypt.DefaultCost)
}

func (u *User) IsPasswordValid(s string) bool {
	return bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(s)) == nil
}
