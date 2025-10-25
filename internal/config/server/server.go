package server

import (
	"flag"
	"net/http"
	"os"

	"github.com/AleGaliev/gofermart/internal/config/db"
	"github.com/AleGaliev/gofermart/internal/consumer"
	"github.com/AleGaliev/gofermart/internal/handler"
	"github.com/AleGaliev/gofermart/internal/logger"
	"github.com/AleGaliev/gofermart/internal/repository/accrual"
	"github.com/AleGaliev/gofermart/internal/storage"
)

const (
	consumerCountWorkers = 3
)

type ServerParams struct {
	AdrHost              string
	AccrualSystemAddress string
	DatabaseDSN          string
	HashKey              string
	ConsumerCountWorkers int
}

type AppConfig struct {
	AdrHost       string
	HandlerApp    http.Handler
	OrderConsumer *consumer.OrderConsumer
	Logger        logger.Logger
	DBConfig      db.PostgresDB
}

func NewServerParams() ServerParams {
	adrHost := flag.String("a", "localhost:8080", "Endpoint http server")
	databaseDSN := flag.String("d", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", "database DSN")
	accrualSystemAddress := flag.String("r", "http://localhost:8081", "Accrual system address")
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

	return ServerParams{
		AdrHost:              *adrHost,
		AccrualSystemAddress: *accrualSystemAddress,
		DatabaseDSN:          *databaseDSN,
		ConsumerCountWorkers: consumerCountWorkers,
		HashKey:              *hashKey,
	}
}

func (p ServerParams) AppConfig() (AppConfig, error) {
	logServer, err := logger.CreateLogger()
	if err != nil {
		return AppConfig{}, err
	}
	dbConfig, err := db.NewPostgresDB(p.DatabaseDSN)
	if err != nil {
		return AppConfig{}, err
	}

	dbStorage := storage.NewPostgresDBStorage(dbConfig)
	accrualConfig := accrual.NewAccrualConfig(logServer, p.AccrualSystemAddress)
	consumerConfig := consumer.NewOrderConsumer(p.ConsumerCountWorkers, accrualConfig, dbStorage)

	r := handler.CreateMyHandler(dbStorage, logServer, p.HashKey)
	return AppConfig{
		AdrHost:       p.AdrHost,
		DBConfig:      dbConfig,
		HandlerApp:    r,
		OrderConsumer: consumerConfig,
		Logger:        logServer,
	}, err
}
