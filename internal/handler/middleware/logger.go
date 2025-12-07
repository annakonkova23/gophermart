package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type responseLogger struct {
	http.ResponseWriter
	status int
	size   int
}

func (l *responseLogger) WriteHeader(code int) {
	if l.status != 0 {
		return
	}
	l.status = code
	l.ResponseWriter.WriteHeader(code)
}

func (l *responseLogger) Write(b []byte) (int, error) {
	if l.status == 0 {
		l.status = http.StatusOK
	}
	size, err := l.ResponseWriter.Write(b)
	l.size += size
	return size, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.URL.RequestURI()
		method := r.Method

		logger := &responseLogger{ResponseWriter: w}

		next.ServeHTTP(logger, r)

		logrus.WithFields(logrus.Fields{
			"method":      method,
			"uri":         uri,
			"status":      logger.status,
			"duration":    time.Since(start).String(),
			"content_len": logger.size,
		}).Info("Запрос выполнен")
	})
}
