package exceptions

import "errors"

var (
	ErrDBError       = errors.New("database error")
	ErrRedisError    = errors.New("redis error")
	ErrNotFound      = errors.New("not found")
	ErrTokenNotFound = errors.New("token not found")
)
