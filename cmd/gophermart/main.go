package main

import (
	"context"
	"net/http"

	"github.com/AleGaliev/gofermart/internal/config/db"
	"github.com/AleGaliev/gofermart/internal/config/server"
	"github.com/AleGaliev/gofermart/internal/consumer"
	"github.com/AleGaliev/gofermart/internal/handler"
	"github.com/AleGaliev/gofermart/internal/logger"
	"github.com/AleGaliev/gofermart/internal/repository/accrual"
	"github.com/AleGaliev/gofermart/internal/storage"
)

func main() {
	serverConfig := server.NewServerConfig()
	logServer, err := logger.CreateLogger()
	if err != nil {
		panic(err)
	}
	dbConfig, err := db.NewPostgresDB(serverConfig.DatabaseDSN)
	if err != nil {
		panic(err)
	}
	if err := dbConfig.CreateMigration(); err != nil {
		panic(err)
	}

	dbStorage := storage.NewPostgresDBStorage(dbConfig)
	accrualConfig := accrual.NewAccrualConfig(logServer, serverConfig.AccrualSystemAddress)
	consumerConfig := consumer.NewOrderConsumer(serverConfig.ConsumerCountWorkers, accrualConfig, dbStorage)

	go consumerConfig.ConsumerRun(context.Background())

	r := handler.CreateMyHandler(dbStorage, logServer, "")
	logServer.StartServerLog(serverConfig.AdrHost)
	err = http.ListenAndServe(serverConfig.AdrHost, r)

	defer dbConfig.Close()
}
