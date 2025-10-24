package logger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

type logger struct {
	logger zap.SugaredLogger
}

func CreateLogger() (logger, error) {
	// создаём предустановленный регистратор zap
	log, err := zap.NewDevelopment()
	if err != nil {
		return logger{}, fmt.Errorf("error creating logger: %v", err)
	}
	defer log.Sync()

	// делаем регистратор SugaredLogger
	sugar := *log.Sugar()

	return logger{sugar}, nil
}

func (l logger) StartServerLog(addr string) {
	l.logger.Infow(
		"Starting server",
		"addr", addr,
	)
}

func (l logger) CreateRequestLog(url, method string, timestamp time.Time) {
	duration := time.Since(timestamp)
	l.logger.Infow(
		"message",
		"url", url,
		"method", method,
		"timestamp", duration,
	)

}

func (l logger) CreateResponseLog(statusCode int, large int64) {
	l.logger.Infow(
		"message",
		"statusCode", statusCode,
		"large", large,
	)
}
