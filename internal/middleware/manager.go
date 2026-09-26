package middleware

import (
	"crypto/rsa"
	"log/slog"
)

type MiddlewareManager struct {
	logger    *slog.Logger
	verifyKey *rsa.PublicKey
}

func NewMiddlewareManager(logger *slog.Logger, verifyKey *rsa.PublicKey) *MiddlewareManager {
	return &MiddlewareManager{
		logger:    logger,
		verifyKey: verifyKey,
	}
}
