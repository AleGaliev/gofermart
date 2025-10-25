package main

import (
	"context"
	"net/http"

	"github.com/AleGaliev/gofermart/internal/config/server"
)

func main() {
	serverConfig, err := server.NewServerParams().AppConfig()
	if err != nil {
		panic(err)
	}

	go serverConfig.OrderConsumer.ConsumerRun(context.Background())

	serverConfig.Logger.StartServerLog(serverConfig.AdrHost)
	err = http.ListenAndServe(serverConfig.AdrHost, serverConfig.HandlerApp)
	if err != nil {
		panic(err)
	}
	defer serverConfig.DBConfig.Close()
}
