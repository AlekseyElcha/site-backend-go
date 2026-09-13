package auth

import (
	"net/http"
	"uuid"
)

func GetUserIDFromCookies(r *http.Request) (uuid.UUID, error) {
	//cookie, err := r.Cookie("access_token")
	//if err != nil {
	//	return uuid.Nil(), err
	//}
	//
	//tokenString := cookie.Value
	//claims := &Claims{}
	//
	//token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
	//	return verifyKey, nil
	//})
	//
	//if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
	//	return uuid.Nil(), fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	//}
	//
	//if err != nil || !token.Valid {
	//	return uuid.Nil(), errors.New("invalid token")
	//}
	//
	//userID, err := uuid.Parse(claims.UserID.String())
	//if err != nil {
	//	return uuid.Nil(), errors.New("no info")
	//}
	//
	//return userID, nil
	return uuid.Parse("57b2c40b-c756-46a8-aef2-1b3710c1cc15")
}
