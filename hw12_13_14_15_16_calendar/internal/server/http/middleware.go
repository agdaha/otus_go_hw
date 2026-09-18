package internalhttp

import (
	"fmt"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.code = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, code: http.StatusOK}
			next.ServeHTTP(rw, r)
			logger.Info(fmt.Sprintf("%s [%s] %s %s %s %d %dms %q",
				r.RemoteAddr,
				start.Format("02/Jan/2006:15:04:05 -0700"),
				r.Method,
				r.RequestURI,
				r.Proto,
				rw.code,
				time.Since(start).Milliseconds(),
				r.UserAgent(),
			))
		})
	}
}
