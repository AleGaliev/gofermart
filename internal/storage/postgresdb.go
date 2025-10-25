package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AleGaliev/gofermart/internal/config/db"
	models "github.com/AleGaliev/gofermart/internal/model"
	"github.com/AleGaliev/gofermart/internal/service/passhash"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresDBStorage struct {
	dbConfig db.PostgresDB
}

func NewPostgresDBStorage(db db.PostgresDB) *PostgresDBStorage {
	return &PostgresDBStorage{
		dbConfig: db,
	}
}

func (p *PostgresDBStorage) Register(user models.User) error {
	exists, err := p.checkUser(user.Login)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("user already exists")
	}

	if err = p.createUser(user.Login, user.Password); err != nil {
		return err
	}
	return nil
}

func (p *PostgresDBStorage) checkUser(username string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.dbConfig.DefaultTimeout*time.Second)
	defer cancel()

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	err := p.dbConfig.DB.QueryRowContext(ctx, query, username).Scan(&exists)

	if err != nil {
		return exists, fmt.Errorf("failed to check user existence: %v", err)
	}

	if exists {
		return exists, nil
	}

	return exists, nil
}

func (p *PostgresDBStorage) createUser(username, password string) error {
	// Хешируем пароль
	passwordHash, err := passhash.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	createdAt := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), p.dbConfig.DefaultTimeout*time.Second)
	defer cancel()
	query := `INSERT INTO users (username, password_hash, created_at, balance, withdrawn ) VALUES ($1, $2, $3, $4, $5)`
	_, err = p.dbConfig.DB.ExecContext(ctx, query, username, passwordHash, createdAt, 0, 0)

	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

func (p *PostgresDBStorage) Login(user models.User) error {
	passwordHash, err := p.getUserPass(user.Login)
	if err != nil {
		return err
	}

	if !passhash.CheckPasswordHash(user.Password, passwordHash) {
		return fmt.Errorf("invalid login/password")
	}

	return nil
}

func (p *PostgresDBStorage) getUserPass(username string) (string, error) {
	var password_hash string
	ctx, cancel := context.WithTimeout(context.Background(), p.dbConfig.DefaultTimeout*time.Second)
	defer cancel()

	query := `SELECT password_hash FROM users WHERE username = $1`
	if err := p.dbConfig.DB.QueryRowContext(ctx, query, username).Scan(&password_hash); err != nil {
		return password_hash, fmt.Errorf("failed to check user existence: %v", err)
	}

	return password_hash, nil
}

func (p *PostgresDBStorage) UploadOrder(orderUser, orderNumber string) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.dbConfig.DefaultTimeout*time.Second)
	defer cancel()

	query := `INSERT INTO orders (order_number, order_user, uploaded_at, accrual_service, status) VALUES ($1, $2, $3, $4, $5)`
	_, err := p.dbConfig.DB.ExecContext(
		ctx, query, orderNumber, orderUser, time.Now(), 0, "NEW")
	if err != nil {
		return fmt.Errorf("failed to insert order: %v", err)
	}

	return nil
}

func (p *PostgresDBStorage) GetOrder(OrderNumber string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.dbConfig.DefaultTimeout*time.Second)
	defer cancel()
	var orderUser string
	query := `SELECT order_user FROM orders WHERE order_number = $1`
	if err := p.dbConfig.DB.QueryRowContext(ctx, query, OrderNumber).Scan(&orderUser); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return orderUser, nil
}

func (p *PostgresDBStorage) GetOrdersUser(login string) ([]models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.dbConfig.DefaultTimeout*time.Second)
	defer cancel()
	var orders []models.Order

	query := `
        SELECT order_number, accrual_service, status, uploaded_at 
        FROM orders 
        WHERE order_user = $1 
        ORDER BY uploaded_at DESC
    `

	rows, err := p.dbConfig.DB.QueryContext(ctx, query, login)
	if err != nil {
		return nil, fmt.Errorf("failed to query user orders: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.OrderNumber, &order.Accrual, &order.Status, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order row: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order rows: %w", err)
	}

	return orders, nil
}

func (p *PostgresDBStorage) GetUserBalance(user string) (models.UserBalance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.dbConfig.DefaultTimeout*time.Second)
	defer cancel()

	var userBalance models.UserBalance
	query := `SELECT balance, withdrawn FROM users WHERE username=$1`
	row := p.dbConfig.DB.QueryRowContext(ctx, query, user)

	if err := row.Scan(&userBalance.Current, &userBalance.Withdrawn); err != nil {
		return models.UserBalance{}, err
	}
	return userBalance, nil
}

func (p *PostgresDBStorage) UploadOrderWithdraw(user string, withdraw models.Withdraw) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.dbConfig.DefaultTimeout*time.Second)
	defer cancel()

	tx, err := p.dbConfig.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	queryCreateOrder := `
		INSERT INTO orders (order_number, order_user, uploaded_at, accrual_service, status) VALUES ($1, $2, $3, $4, $5)
    `

	queryUpdateBalance := `
		UPDATE users
		SET
		balance = balance - $1,
			withdrawn = withdrawn + $1,
			uploaded_at = $3
		WHERE username = $2
	`
	queryCreateWithdrawn := `
		INSERT INTO withdrawals (order_number, order_user, sum_withdrawals, processed_at) VALUES ($1, $2, $3, $4)
    `

	_, err = tx.ExecContext(ctx, queryCreateOrder, withdraw.OrderNumber, user, time.Now(), 0, "NEW")
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	_, err = tx.ExecContext(ctx, queryUpdateBalance, withdraw.Sum, user, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	_, err = tx.ExecContext(ctx, queryCreateWithdrawn, withdraw.OrderNumber, user, withdraw.Sum, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create withdrawn: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *PostgresDBStorage) GetUserWithdrawals(user string) ([]models.Withdraw, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.dbConfig.DefaultTimeout*time.Second)
	defer cancel()

	query := `SELECT order_number, sum_withdrawals, processed_at 
              FROM withdrawals 
              WHERE order_user = $1 
              ORDER BY processed_at DESC`

	rows, err := p.dbConfig.DB.QueryContext(ctx, query, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []models.Withdraw
	for rows.Next() {
		var withdraw models.Withdraw

		err := rows.Scan(&withdraw.OrderNumber, &withdraw.Sum, &withdraw.ProcessedAt)
		if err != nil {
			return nil, err
		}

		withdrawals = append(withdrawals, withdraw)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (p *PostgresDBStorage) GetOrderNotFinish() ([]models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.dbConfig.DefaultTimeout*time.Second)
	defer cancel()
	query := `
        SELECT order_user, order_number, status 
        FROM orders 
        WHERE status NOT IN ('PROCESSED', 'INVALID')
    `

	rows, err := p.dbConfig.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.OrderUser,
			&order.OrderNumber,
			&order.Status,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		return nil, nil
	}

	return orders, nil
}

func (p *PostgresDBStorage) UpdateOrderAndBalance(order models.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.dbConfig.DefaultTimeout*time.Second)
	defer cancel()

	tx, err := p.dbConfig.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	queryUpdateOrder := `
        UPDATE orders 
        SET status = $1, accrual_service = accrual_service + $2
        WHERE order_number = $3
    `
	queryUpdateBalance := `
        UPDATE users 
        SET balance = balance + $1
        WHERE username = $2
    `

	_, err = tx.ExecContext(ctx, queryUpdateOrder, order.Status, order.Accrual, order.OrderNumber)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	_, err = tx.ExecContext(ctx, queryUpdateBalance, order.Accrual, order.OrderUser)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
