package middleware

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	Iat    int64     `json:"iat"`
	Exp    int64     `json:"exp"`
	jwt.RegisteredClaims
}

type ContextKey string

const (
	UserIDKey   ContextKey = "userID"
	UserRoleKey ContextKey = "userRole"
)

func (mm *MiddlewareManager) AuthCheckMiddleware(reqRoles []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("unauthorized"))
				return
			}

			tokenString := cookie.Value
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return mm.verifyKey, nil
			})

			if err != nil || !token.Valid {
				mm.logger.Warn("invalid token attempt", "err", err)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("token invalid"))
				return
			}

			userID := claims.UserID
			userRole := claims.Role

			if !slices.Contains(reqRoles, claims.Role) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("forbidden"))
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserRoleKey, userRole)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
