package server

import (
	"flag"
	"os"
)

const (
	consumerCountWorkers = 3
)

type ServerConfig struct {
	AdrHost              string
	AccrualSystemAddress string
	DatabaseDSN          string
	HashKey              string
	ConsumerCountWorkers int
}

func NewServerConfig() ServerConfig {
	adrHost := flag.String("a", "localhost:8080", "Endpoint http server")
	databaseDSN := flag.String("d", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", "database DSN")
	accrualSystemAddress := flag.String("r", "localhost:8081", "Accrual system address")
	hashKey := flag.String("k", "", "key jwtManager")
	flag.Parse()

	varAdrHost, ok := os.LookupEnv("RUN_ADDRESS")
	if ok {
		adrHost = &varAdrHost
	}

	varDatabaseDSN, ok := os.LookupEnv("DATABASE_URI")
	if ok {
		databaseDSN = &varDatabaseDSN
	}

	varAccrualSystemAddress, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if ok {
		accrualSystemAddress = &varAccrualSystemAddress
	}

	varHashKey, ok := os.LookupEnv("KEY")
	if ok {
		hashKey = &varHashKey
	}

	return ServerConfig{
		AdrHost:              *adrHost,
		AccrualSystemAddress: *accrualSystemAddress,
		DatabaseDSN:          *databaseDSN,
		ConsumerCountWorkers: consumerCountWorkers,
		HashKey:              *hashKey,
	}
}
