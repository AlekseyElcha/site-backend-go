package middleware

import (
	"fmt"
	"net/http"
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

func (mm *MiddlewareManager) AuthCheckMiddleware(reqRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
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

			if claims.Role != reqRole {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("forbidden"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
