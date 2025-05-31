package model

const (
	UserRoleSuperRoot UserRole = "super_root"
	UserRoleRoot      UserRole = "root"
	UserRoleManager   UserRole = "manager"
	UserRoleEmpty     UserRole = ""
)

type UserRole string

func (r UserRole) String() string {
	return string(r)
}

func (r UserRole) IsRoot() bool {
	return r == UserRoleRoot
}

func (r UserRole) IsSuperRoot() bool {
	return r == UserRoleSuperRoot
}

func (r UserRole) IsAnyRoot() bool {
	return r == UserRoleRoot || r == UserRoleSuperRoot
}

func (r UserRole) IsManager() bool {
	return r == UserRoleManager
}

func (r UserRole) IsEmptyRole() bool {
	return r == UserRoleEmpty
}
