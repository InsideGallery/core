# server/jwt/model

Import path: `github.com/InsideGallery/core/server/jwt/model`

`model` contains JWT authorization models: user roles, scopes, scope checking,
scope normalization, and the user record shape used by the JWT package.

## Main APIs

- `UserRole` constants: `UserRoleSuperRoot`, `UserRoleRoot`,
  `UserRoleManager`, and `UserRoleEmpty`.
- `Scope`, `AccessType`, and `Scopes`: scope values such as
  `read:gallery:view`, `write:gallery:*`, `all:*`, or `*`.
- `MethodMap`: maps HTTP methods to read or write access types.
- `ScopeFrom(scopeName)`: parses scope strings.
- `NewScopeChecker(ctx, scope)`: creates a checker for one target scope.
- `ScopeChecker.IsAllowed(role, scopes, requestPasswordChange)`: evaluates
  root roles, wildcard scopes, exact matches, and read/write/all permissions.
- `NormalizeScopes(ctx, scopes)`: sorts and collapses redundant scopes.
- `User`, `NewUser`, `SetPassword`, and `IsPasswordValid`: user model helpers.

## Usage

```go
scope, err := model.ScopeFrom("read:gallery:view")
if err != nil {
	return err
}

checker, err := model.NewScopeChecker(ctx, scope.String())
if err != nil {
	return err
}

allowed := checker.IsAllowed(model.UserRoleManager, model.Scopes{scope}, false)
```

## Operational Notes

`User` uses MongoDB `primitive.ObjectID` values and stores bcrypt password
hashes. Scope checks allow root and super-root roles automatically. When
`requestPasswordChange` is true, only `write:access:password` is allowed.
