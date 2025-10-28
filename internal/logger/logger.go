package logger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Logger struct {
	logger zap.SugaredLogger
}

func CreateLogger() (Logger, error) {
	// создаём предустановленный регистратор zap
	log, err := zap.NewDevelopment()
	if err != nil {
		return Logger{}, fmt.Errorf("error creating logger: %v", err)
	}
	defer log.Sync()

	// делаем регистратор SugaredLogger
	sugar := *log.Sugar()

	return Logger{sugar}, nil
}

func (l Logger) StartServerLog(addr string) {
	l.logger.Infow(
		"Starting server",
		"addr", addr,
	)
}

func (l Logger) CreateRequestLog(url, method string, timestamp time.Time) {
	duration := time.Since(timestamp)
	l.logger.Infow(
		"message",
		"url", url,
		"method", method,
		"timestamp", duration,
	)

}

func (l Logger) CreateResponseLog(statusCode int, large int64) {
	l.logger.Infow(
		"message",
		"statusCode", statusCode,
		"large", large,
	)
}

func (l Logger) CreateErrorLog(service, message string) {
	l.logger.Errorw(
		"message",
		"service", service,
		"error", message,
	)
}
