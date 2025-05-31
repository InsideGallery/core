package middlewares

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	coreJWT "github.com/InsideGallery/core/server/jwt"
	jwtModel "github.com/InsideGallery/core/server/jwt/model"
	"github.com/InsideGallery/core/server/webserver"
	"github.com/gofiber/fiber/v2"
)

type ScopeMiddleware struct {
	ctx context.Context
}

func NewScopeMiddleware(ctx context.Context) *ScopeMiddleware {
	return &ScopeMiddleware{
		ctx: ctx,
	}
}

func (m *ScopeMiddleware) CheckScope(c *fiber.Ctx) error {
	claims, err := coreJWT.DecodeClaims(c)
	if err != nil {
		slog.Default().Error("Decode JWT claims", "err", err)
		return c.Status(http.StatusBadRequest).JSON(webserver.GetResponseWithError(err, 0))
	}

	scope := parseScope(c)

	slog.Default().Debug("Checking scope", "scope", scope, "claims", claims)

	checker, err := jwtModel.NewScopeChecker(m.ctx, scope)
	if err != nil {
		slog.Default().Error("Checking scope", "scope", scope, "method", c.Method(), "path", c.OriginalURL())
		return c.Status(http.StatusForbidden).JSON(webserver.GetResponseWithError(err, 0))
	}

	if !checker.IsAllowed(claims.Role, claims.Scopes, claims.ChangePassword) {
		slog.Default().Debug("Action not allowed", "user_id", claims.UserID, "scope", scope)

		return c.Status(http.StatusForbidden).JSON(webserver.GetResponseWithError(errors.New("action not allowed"), 0))
	}

	return c.Next()
}

func parseScope(c *fiber.Ctx) string {
	s := jwtModel.Scope{
		AccessType: jwtModel.MethodMap[c.Method()],
	}

	items := strings.Split(strings.TrimPrefix(c.Path(), "/"), "/")

	if len(items[0]) == 2 && items[0][:1] == "v" {
		items = items[1:]
	}

	s.Service = items[0]
	s.Action = strings.Join(items[1:], "/")

	return s.String()
}
