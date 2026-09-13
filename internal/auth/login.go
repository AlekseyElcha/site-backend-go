package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
)

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

func init() {
	privateKeyBytes, err := os.ReadFile("certs/private.pem")
	if err != nil {
		log.Fatalf("public keyq	 error: %v", err)
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		log.Fatalf("private key error: %v", err)
	}

	publicKeyBytes, err := os.ReadFile("certs/public.pem")
	if err != nil {
		log.Fatalf("private key error %v", err)
	}
	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		log.Fatalf("private key error: %v", err)
	}
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	Iat    int64     `json:"iat"`
	Exp    int64     `json:"exp"`
	jwt.RegisteredClaims
}

func generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse("472c5567-3ec1-4e70-afb3-8301bf3c6069")
	userRole := "user"
	expirationTimeAccess := time.Now().Add(5 * time.Hour)
	expirationTimeRefresh := time.Now().Add(7 * 24 * time.Hour)

	claims := &Claims{
		UserID:    userID,
		Role:      userRole,
		Iat:       time.Now().Unix(),
		Exp:       expirationTimeAccess.Unix(),
		ExpiresAt: jwt.NewNumericDate(expirationTimeAccess),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	accessTokenString, err := token.SignedString(signKey)
	refreshTokenString := generateRandomToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    accessTokenString,
		Expires:  expirationTimeAccess,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshTokenString,
		Expires:  expirationTimeRefresh,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/auth/refresh",
	})
	fmt.Printf("%s\n, %s\n", accessTokenString, refreshTokenString)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

func AuthCheckMiddleware(reqRole string, next http.Handler) http.Handler {
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
			return verifyKey, nil
		})

		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("token invalid"))
			return
		}

		if claims.Role != reqRole {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("forbidden"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
