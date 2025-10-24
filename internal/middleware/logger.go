package middleware

import (
	"net/http"
	"time"
)

type Logger interface {
	CreateRequestLog(url, method string, timestamp time.Time)
}

func MiddlewareHandlerLogger(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(res, req)
		logger.CreateRequestLog(req.RequestURI, req.Method, start)
	})
}
