package middleware

import (
	"crypto/rand"
	"fmt"
)

func generateRandomRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
