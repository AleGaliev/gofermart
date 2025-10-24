package checkerror

import (
	"errors"
	"net"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func IsRetriableError(err error) bool {
	var conectionError net.Error
	if errors.As(err, &conectionError) {
		return conectionError.Timeout()
	}

	var postgresErr *pgconn.PgError
	if errors.As(err, &postgresErr) {
		return pgerrcode.IsConnectionException(postgresErr.Code)
	}
	problemsList := []string{
		"connection refused",
		"connection reset",
		"no route to host",
		"timeout",
		"too many connections",
		"connection limit exceeded",
		"error sending reques",
	}
	for _, problem := range problemsList {
		if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(problem)) {
			return true
		}
	}
	return false
}
