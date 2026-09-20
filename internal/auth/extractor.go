package auth

import (
	"errors"
	"fmt"
	"net/http"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
)

func GetUserIDFromCookies(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return uuid.Nil(), err
	}

	tokenString := cookie.Value
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return uuid.Nil(), fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	if err != nil || !token.Valid {
		return uuid.Nil(), errors.New("invalid token")
	}

	userID, err := uuid.Parse(claims.UserID.String())
	if err != nil {
		return uuid.Nil(), errors.New("no info")
	}

	fmt.Println(userID)
	return userID, nil
	//return uuid.Parse("77cd1069-5ed9-4fce-8867-4b6829298936")
}
