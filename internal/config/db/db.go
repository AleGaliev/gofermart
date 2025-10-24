package db

import (
	"context"
	"database/sql"
	"time"
)

const (
	connMaxLifetime time.Duration = 30 * time.Minute
	connMaxIdleTime time.Duration = 5 * time.Minute
	maxIdleConns    int           = 5
	maxOpenConns    int           = 25
	connTimeout     time.Duration = 5 * time.Second
	defaultTimeout  time.Duration = 5 * time.Second
	queryMigration                = `
			CREATE TABLE IF NOT EXISTS users (
				username VARCHAR(50) NOT NULL,
				password_hash VARCHAR(200) NOT NULL,
			    balance float,
				withdrawn float,
				created_at TIMESTAMP WITH TIME ZONE,
				uploaded_at TIMESTAMP WITH TIME ZONE
			);
			CREATE TABLE IF NOT EXISTS orders (
				order_user VARCHAR(50) NOT NULL,
				order_number VARCHAR(50) UNIQUE,
				uploaded_at TIMESTAMP WITH TIME ZONE,
				accrual_service float,
				status text
			);
			CREATE TABLE IF NOT EXISTS withdrawals (
				order_number VARCHAR(50),
				order_user VARCHAR(50),
				sum_withdrawals float,
				processed_at TIMESTAMP WITH TIME ZONE
			)
	`
)

type PostgresDB struct {
	DB             *sql.DB
	DefaultTimeout time.Duration
}

func NewPostgresDB(PostgresURL string) (PostgresDB, error) {
	db, err := sql.Open("pgx", PostgresURL)
	if err != nil {
		return PostgresDB{}, err
	}
	configureConnectionPool(db)

	return PostgresDB{
		DB:             db,
		DefaultTimeout: defaultTimeout,
	}, nil
}

func configureConnectionPool(db *sql.DB) {
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxIdleTime(connMaxIdleTime)
}

func (p *PostgresDB) CreateMigration() error {
	ctx, cancel := context.WithTimeout(context.Background(), p.DefaultTimeout)
	defer cancel()

	_, err := p.DB.ExecContext(ctx, queryMigration)
	return err
}

func (p *PostgresDB) Close() error {
	return p.DB.Close()
}
