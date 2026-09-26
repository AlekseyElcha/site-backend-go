package middleware

import (
	"context"
	"net/http"
	"time"
)

type contextKey string

const (
	requestIDKey contextKey = "requestID"
	headerName   string     = "X-Request-ID"
)

func GetRequestID(ctx context.Context) string {
	id, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return ""
	}
	return id
}

type responseWriteWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriteWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (mm *MiddlewareManager) LoggingAndRequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		reqID := r.Header.Get(headerName)
		if reqID == "" {
			reqID = generateRandomRequestID()
		}

		w.Header().Set(headerName, reqID)

		requestLogger := mm.logger.With("request_id", reqID)

		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		r = r.WithContext(ctx)

		requestLogger.Info(
			"incoming request",
			"method", r.Method,
			"path", r.URL.String(),
			"remote_addr", r.RemoteAddr,
		)

		wrappedWriter := &responseWriteWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrappedWriter, r)

		requestLogger.Info(
			"request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"status_code", wrappedWriter.statusCode,
			"duration", time.Since(start),
		)
	})
}
