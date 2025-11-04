package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AleGaliev/gofermart/internal/config/server"
)

func main() {
	fmt.Println("Starting server...")
	serverConfig, err := server.NewServerParams().AppConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err = serverConfig.DBConfig.CreateMigration(); err != nil {
		log.Fatal(err)
	}
	defer serverConfig.DBConfig.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go serverConfig.OrderConsumer.ConsumerRun(ctx)

	go func() {
		serverConfig.Logger.StartServerLog(serverConfig.Server.Addr)
		if err := serverConfig.Server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	fmt.Println("Shutting down server...")
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()
	cancel()

	if err := serverConfig.Server.Shutdown(stopCtx); err != nil {
		log.Fatal(err)
	}

}
