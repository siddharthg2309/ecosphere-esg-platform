package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type Principal struct {
	UserID       id.ID
	Role         domain.Role
	DepartmentID *id.ID
}
type principalKey struct{}

func PrincipalFrom(ctx context.Context) (Principal, bool) {
	value, ok := ctx.Value(principalKey{}).(Principal)
	return value, ok
}

func Authenticate(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				httpserver.Error(w, errs.Unauthorized("authentication_required", "Authentication is required"))
				return
			}
			claims, err := VerifyAccess(strings.TrimPrefix(header, "Bearer "), secret)
			if err != nil {
				httpserver.Error(w, errs.Unauthorized("invalid_access_token", "Access token is invalid or expired"))
				return
			}
			userID, err := id.Parse(claims.Subject)
			if err != nil {
				httpserver.Error(w, errs.Unauthorized("invalid_access_token", "Access token is invalid"))
				return
			}
			principal := Principal{UserID: userID, Role: domain.Role(claims.Role)}
			if claims.DepartmentID != nil {
				departmentID, parseErr := id.Parse(*claims.DepartmentID)
				if parseErr == nil {
					principal.DepartmentID = &departmentID
				}
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalKey{}, principal)))
		})
	}
}

func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	allowed := map[domain.Role]bool{}
	for _, role := range roles {
		allowed[role] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := PrincipalFrom(r.Context())
			if !ok {
				httpserver.Error(w, errs.Unauthorized("authentication_required", "Authentication is required"))
				return
			}
			if !allowed[principal.Role] {
				httpserver.Error(w, errs.Forbidden("role_forbidden", "You do not have permission to perform this action"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
