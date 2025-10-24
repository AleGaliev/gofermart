package server

import (
	"errors"
	"time"

	"github.com/AleGaliev/runtimemetrics/internal/config/db"
	"github.com/AleGaliev/runtimemetrics/internal/config/server"
	"github.com/AleGaliev/runtimemetrics/internal/filestore"
	"github.com/AleGaliev/runtimemetrics/internal/handler"
	"github.com/AleGaliev/runtimemetrics/internal/service/retry"
	"github.com/AleGaliev/runtimemetrics/internal/storage"
)

type ServerMemStorage struct {
	MemStorage handler.Storage
	DBConfig   db.PostgresDB
}

func NewServerMemStorage(serverConf server.ServerConfig) (ServerMemStorage, error) {
	if serverConf.DatabaseDSN == "" {
		fileStore := filestore.NewFileStore(serverConf.FileStoragePath)

		memStorage, err := storage.CreateStorage(fileStore, serverConf.StoreInterval, serverConf.Restore)
		if err != nil {
			return ServerMemStorage{}, err
		}

		if serverConf.StoreInterval > 0 && serverConf.DatabaseDSN == "" {
			go func() {
				for {
					time.Sleep(time.Duration(serverConf.StoreInterval) * time.Second)
					if err := memStorage.SaveMetricToFile(); err != nil {
						panic(errors.Unwrap(err))
					}
				}
			}()
		}
		defer memStorage.SaveMetricToFile()
		return ServerMemStorage{
			MemStorage: memStorage,
			DBConfig:   db.PostgresDB{},
		}, nil
	}
	dbConfig, err := db.NewPostgresDB(serverConf.DatabaseDSN)

	if err != nil {
		return ServerMemStorage{}, err
	}

	if err = dbConfig.CreateMigration(); err != nil {
		return ServerMemStorage{}, err
	}

	dbMemStorage := storage.NewPostgresDBStorage(dbConfig, retry.CreateRetry())

	return ServerMemStorage{
		MemStorage: dbMemStorage,
		DBConfig:   dbConfig,
	}, nil
}
