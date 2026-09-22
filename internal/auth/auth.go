package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"site-backend-go/internal/dtos"
	"site-backend-go/internal/redis_client"
	"site-backend-go/internal/service"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
)

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

type AuthHandler struct {
	userService *service.UserService
	rdb         *redis_client.RedisClient
}

func NewAuthHandler(rdb *redis_client.RedisClient, userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		rdb:         rdb,
	}
}

func init() {
	privateKeyBytes, err := os.ReadFile("certs/private.pem")
	if err != nil {
		log.Fatalf("public keys error: %v", err)
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

func (s *AuthHandler) GetSelfInfoFromCookies(w http.ResponseWriter, r *http.Request) {
	cookieToken, err := r.Cookie("auth_token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(cookieToken.Value, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return verifyKey, nil
	})
	if err != nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	resp := dtos.SelfInfoResponse{
		UserID: claims.UserID.String(),
		Role:   claims.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *AuthHandler) RequestAuthCodeHandler(w http.ResponseWriter, r *http.Request) {
	var m dtos.AuthCodeRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Ошибка декодирования данных."}`))
		return
	}

	authCodeStr := GenerateAuthCode()
	err = s.rdb.AddAuthCodeToRedis(r.Context(), m.Email, authCodeStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Код отправлен на почту."}`))
}

func (s *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var lr dtos.LoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&lr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	userInfo, err := s.userService.GetUserByEmail(ctx, lr.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
			return
		}
	}

	var userID uuid.UUID
	var userRole string
	if userInfo == nil {
		newUser := dtos.UserCreateRequest{
			Email: lr.Email,
		}
		userID, err = s.userService.Create(ctx, newUser)
		userRole = "user" // для создания пользователя
	} else {
		userID = userInfo.ID
		userRole = userInfo.Role
	}

	//userID, _ = uuid.Parse("77cd1069-5ed9-4fce-8867-4b6829298936") // заглушка
	//userID, err = GetUserIDFromCookies(r)
	//if err != nil {
	//
	//}

	expirationTimeAccess := time.Now().Add(5 * time.Hour)
	expirationTimeRefresh := time.Now().Add(7 * 24 * time.Hour)

	isCodeValid, err := s.rdb.IsAuthCodeFromUserValid(r.Context(), lr.Email, lr.AuthCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`)) // переписать
		return
	}

	if !isCodeValid {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"valid": false}`)) // переисать
		return
	}

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
	// fmt.Printf("%s\n, %s\n", accessTokenString, refreshTokenString)

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

func (s *AuthHandler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Refresh token missing"}`))
		return
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return signKey, nil
	})

	if err != nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Invalid or expired refresh token"}`))
		return
	}

	expirationTimeAccess := time.Now().Add(5 * time.Hour)
	expirationTimeRefresh := time.Now().Add(7 * 24 * time.Hour)

	accessClaims := &Claims{
		UserID:    claims.UserID,
		Role:      claims.Role,
		Iat:       time.Now().Unix(),
		Exp:       expirationTimeAccess.Unix(),
		ExpiresAt: jwt.NewNumericDate(expirationTimeAccess),
	}
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessTokenString, err := accessTokenObj.SignedString(signKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	refreshClaims := &Claims{
		UserID:    claims.UserID,
		Role:      claims.Role,
		Iat:       time.Now().Unix(),
		Exp:       expirationTimeRefresh.Unix(),
		ExpiresAt: jwt.NewNumericDate(expirationTimeRefresh),
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshTokenString, err := refreshTokenObj.SignedString(signKey)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

func (s *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}
