package auth

import (
	"math/rand"
	"strconv"
)

func GenerateAuthCode() string {
	code := rand.Intn(900_000) + 100_000
	return strconv.Itoa(code)
}
