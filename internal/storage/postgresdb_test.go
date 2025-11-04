package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/AleGaliev/gofermart/internal/config/db"
	model "github.com/AleGaliev/gofermart/internal/model"
	"github.com/AleGaliev/gofermart/internal/service/passhash"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresDBStorage_Register(t *testing.T) {
	dbMock, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbMock.Close()

	storage := &PostgresDBStorage{
		dbConfig: db.PostgresDB{
			DB:             dbMock,
			DefaultTimeout: 5,
		},
	}

	tests := []struct {
		name        string
		user        model.User
		mock        func()
		expectedErr error
	}{
		{
			name: "successful registration",
			user: model.User{
				Login:    "testuser",
				Password: "password123",
			},
			mock: func() {
				// Mock checkUser - user doesn't exist
				sqlMock.ExpectQuery("SELECT EXISTS").
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				// Mock createUser
				sqlMock.ExpectExec("INSERT INTO users").
					WithArgs("testuser", sqlmock.AnyArg(), sqlmock.AnyArg(), 0, 0).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedErr: nil,
		},
		{
			name: "user already exists",
			user: model.User{
				Login:    "existinguser",
				Password: "password123",
			},
			mock: func() {
				sqlMock.ExpectQuery("SELECT EXISTS").
					WithArgs("existinguser").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			expectedErr: fmt.Errorf("user already exists"),
		},
		{
			name: "check user error",
			user: model.User{
				Login:    "testuser",
				Password: "password123",
			},
			mock: func() {
				sqlMock.ExpectQuery("SELECT EXISTS").
					WithArgs("testuser").
					WillReturnError(errors.New("db error"))
			},
			expectedErr: fmt.Errorf("failed to check user existence: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := storage.Register(tt.user)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestPostgresDBStorage_Login(t *testing.T) {
	dbMock, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbMock.Close()

	storage := &PostgresDBStorage{
		dbConfig: db.PostgresDB{
			DB:             dbMock,
			DefaultTimeout: 5,
		},
	}

	tests := []struct {
		name        string
		user        model.User
		mock        func()
		expectedErr error
	}{
		{
			name: "successful login",
			user: model.User{
				Login:    "testuser",
				Password: "password123",
			},
			mock: func() {
				hashedPassword, _ := passhash.HashPassword("password123")
				sqlMock.ExpectQuery("SELECT password_hash FROM users").
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).AddRow(hashedPassword))
			},
			expectedErr: nil,
		},
		{
			name: "user not found",
			user: model.User{
				Login:    "nonexistent",
				Password: "password123",
			},
			mock: func() {
				sqlMock.ExpectQuery("SELECT password_hash FROM users").
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr: fmt.Errorf("failed to check user existence"),
		},
		{
			name: "wrong password",
			user: model.User{
				Login:    "testuser",
				Password: "wrongpassword",
			},
			mock: func() {
				hashedPassword, _ := passhash.HashPassword("password123")
				sqlMock.ExpectQuery("SELECT password_hash FROM users").
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).AddRow(hashedPassword))
			},
			expectedErr: fmt.Errorf("invalid login/password"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := storage.Login(tt.user)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestPostgresDBStorage_UploadOrder(t *testing.T) {
	dbMock, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbMock.Close()

	storage := &PostgresDBStorage{
		dbConfig: db.PostgresDB{
			DB:             dbMock,
			DefaultTimeout: 5,
		},
	}

	tests := []struct {
		name        string
		orderUser   string
		orderNumber string
		mock        func()
		expectedErr error
	}{
		{
			name:        "successful upload",
			orderUser:   "user1",
			orderNumber: "2377225624",
			mock: func() {
				sqlMock.ExpectExec("INSERT INTO orders").
					WithArgs("2377225624", "user1", sqlmock.AnyArg(), 0, "NEW").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedErr: nil,
		},
		{
			name:        "upload error",
			orderUser:   "user1",
			orderNumber: "2377225624",
			mock: func() {
				sqlMock.ExpectExec("INSERT INTO orders").
					WithArgs("2377225624", "user1", sqlmock.AnyArg(), 0, "NEW").
					WillReturnError(errors.New("db error"))
			},
			expectedErr: fmt.Errorf("failed to insert order: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := storage.UploadOrder(tt.orderUser, tt.orderNumber)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestPostgresDBStorage_GetOrder(t *testing.T) {
	dbMock, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbMock.Close()

	storage := &PostgresDBStorage{
		dbConfig: db.PostgresDB{
			DB:             dbMock,
			DefaultTimeout: 5,
		},
	}

	tests := []struct {
		name          string
		orderNumber   string
		mock          func()
		expectedUser  string
		expectedError error
	}{
		{
			name:        "order found",
			orderNumber: "2377225624",
			mock: func() {
				sqlMock.ExpectQuery("SELECT order_user FROM orders").
					WithArgs("2377225624").
					WillReturnRows(sqlmock.NewRows([]string{"order_user"}).AddRow("user1"))
			},
			expectedUser:  "user1",
			expectedError: nil,
		},
		{
			name:        "order not found",
			orderNumber: "2377225624",
			mock: func() {
				sqlMock.ExpectQuery("SELECT order_user FROM orders").
					WithArgs("2377225624").
					WillReturnError(sql.ErrNoRows)
			},
			expectedUser:  "",
			expectedError: nil,
		},
		{
			name:        "query error",
			orderNumber: "2377225624",
			mock: func() {
				sqlMock.ExpectQuery("SELECT order_user FROM orders").
					WithArgs("2377225624").
					WillReturnError(errors.New("db error"))
			},
			expectedUser:  "",
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			user, err := storage.GetOrder(tt.orderNumber)

			assert.Equal(t, tt.expectedUser, user)
			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestPostgresDBStorage_GetOrdersUser(t *testing.T) {
	dbMock, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbMock.Close()

	storage := &PostgresDBStorage{
		dbConfig: db.PostgresDB{
			DB:             dbMock,
			DefaultTimeout: 5,
		},
	}

	now := time.Now()

	tests := []struct {
		name           string
		login          string
		mock           func()
		expectedOrders []model.OrderOut
		expectedError  bool
	}{
		{
			name:  "successful get orders",
			login: "user1",
			mock: func() {
				rows := sqlmock.NewRows([]string{"order_number", "accrual_service", "status", "uploaded_at"}).
					AddRow("2377225624", 100.5, "PROCESSED", now).
					AddRow("12345678903", 50.0, "NEW", now.Add(-time.Hour))
				sqlMock.ExpectQuery("SELECT order_number, accrual_service, status, uploaded_at FROM orders").
					WithArgs("user1").
					WillReturnRows(rows)
			},
			expectedOrders: []model.OrderOut{
				{OrderNumber: "2377225624", Accrual: 100.5, Status: "PROCESSED", UploadedAt: now},
				{OrderNumber: "12345678903", Accrual: 50.0, Status: "NEW", UploadedAt: now.Add(-time.Hour)},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			orders, err := storage.GetOrdersUser(tt.login)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOrders, orders)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestPostgresDBStorage_GetUserBalance(t *testing.T) {
	dbMock, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbMock.Close()

	storage := &PostgresDBStorage{
		dbConfig: db.PostgresDB{
			DB:             dbMock,
			DefaultTimeout: 5,
		},
	}

	tests := []struct {
		name            string
		user            string
		mock            func()
		expectedBalance model.UserBalanceOut
		expectedError   bool
	}{
		{
			name: "successful get balance",
			user: "user1",
			mock: func() {
				sqlMock.ExpectQuery("SELECT balance, withdrawn FROM users").
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"balance", "withdrawn"}).AddRow(100.5, 50.0))
			},
			expectedBalance: model.UserBalanceOut{Current: 100.5, Withdrawn: 50.0},
			expectedError:   false,
		},
		{
			name: "user not found",
			user: "nonexistent",
			mock: func() {
				sqlMock.ExpectQuery("SELECT balance, withdrawn FROM users").
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			expectedBalance: model.UserBalanceOut{},
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			balance, err := storage.GetUserBalance(tt.user)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBalance, balance)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestPostgresDBStorage_UploadOrderWithdraw(t *testing.T) {
	dbMock, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbMock.Close()

	storage := &PostgresDBStorage{
		dbConfig: db.PostgresDB{
			DB:             dbMock,
			DefaultTimeout: 5,
		},
	}

	withdraw := model.Withdraw{
		OrderNumber: "12345678903",
		Sum:         decimal.NewFromFloat32(50.0),
	}

	tests := []struct {
		name        string
		user        string
		withdraw    model.Withdraw
		mock        func()
		expectedErr error
	}{
		{
			name:     "successful withdraw",
			user:     "user1",
			withdraw: withdraw,
			mock: func() {
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec("INSERT INTO orders").
					WithArgs("12345678903", "user1", sqlmock.AnyArg(), 0, "NEW").
					WillReturnResult(sqlmock.NewResult(1, 1))
				sqlMock.ExpectExec("UPDATE users").
					WithArgs(decimal.NewFromFloat32(50.0), "user1", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectExec("INSERT INTO withdrawals").
					WithArgs("12345678903", "user1", decimal.NewFromFloat32(50.0), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				sqlMock.ExpectCommit()
			},
			expectedErr: nil,
		},
		{
			name:     "transaction rollback on error",
			user:     "user1",
			withdraw: withdraw,
			mock: func() {
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec("INSERT INTO orders").
					WithArgs("12345678903", "user1", sqlmock.AnyArg(), 0, "NEW").
					WillReturnError(errors.New("db error"))
				sqlMock.ExpectRollback()
			},
			expectedErr: fmt.Errorf("failed to create order: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := storage.UploadOrderWithdraw(tt.user, tt.withdraw)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestPostgresDBStorage_GetUserWithdrawals(t *testing.T) {
	dbMock, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbMock.Close()

	storage := &PostgresDBStorage{
		dbConfig: db.PostgresDB{
			DB:             dbMock,
			DefaultTimeout: 5,
		},
	}

	now := time.Now()

	tests := []struct {
		name              string
		user              string
		mock              func()
		expectedWithdraws []model.WithdrawOut
		expectedError     bool
	}{
		{
			name: "successful get withdrawals",
			user: "user1",
			mock: func() {
				rows := sqlmock.NewRows([]string{"order_number", "sum_withdrawals", "processed_at"}).
					AddRow("12345678903", 50.0, now).
					AddRow("346436439", 25.5, now.Add(-time.Hour))
				sqlMock.ExpectQuery("SELECT order_number, sum_withdrawals, processed_at FROM withdrawals").
					WithArgs("user1").
					WillReturnRows(rows)
			},
			expectedWithdraws: []model.WithdrawOut{
				{OrderNumber: "12345678903", Sum: 50.0, ProcessedAt: now},
				{OrderNumber: "346436439", Sum: 25.5, ProcessedAt: now.Add(-time.Hour)},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			withdrawals, err := storage.GetUserWithdrawals(tt.user)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedWithdraws, withdrawals)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestPostgresDBStorage_GetOrderNotFinish(t *testing.T) {
	dbMock, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbMock.Close()

	storage := &PostgresDBStorage{
		dbConfig: db.PostgresDB{
			DB:             dbMock,
			DefaultTimeout: 5,
		},
	}

	tests := []struct {
		name           string
		mock           func()
		expectedOrders []model.Order
		expectedError  bool
	}{
		{
			name: "successful get unfinished orders",
			mock: func() {
				rows := sqlmock.NewRows([]string{"order_user", "order_number", "status"}).
					AddRow("user1", "9278923470", "NEW").
					AddRow("user2", "346436439", "PROCESSING")
				sqlMock.ExpectQuery("SELECT order_user, order_number, status FROM orders").
					WillReturnRows(rows)
			},
			expectedOrders: []model.Order{
				{OrderUser: "user1", OrderNumber: "9278923470", Status: "NEW"},
				{OrderUser: "user2", OrderNumber: "346436439", Status: "PROCESSING"},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			orders, err := storage.GetOrderNotFinish()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOrders, orders)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestPostgresDBStorage_UpdateOrderAndBalance(t *testing.T) {
	dbMock, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbMock.Close()

	storage := &PostgresDBStorage{
		dbConfig: db.PostgresDB{
			DB:             dbMock,
			DefaultTimeout: 5,
		},
	}

	order := model.Order{
		OrderNumber: "9278923470",
		OrderUser:   "user1",
		Status:      "PROCESSED",
		Accrual:     decimal.NewFromFloat32(100.5),
	}

	tests := []struct {
		name        string
		order       model.Order
		mock        func()
		expectedErr error
	}{
		{
			name:  "successful update",
			order: order,
			mock: func() {
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec("UPDATE orders").
					WithArgs("PROCESSED", decimal.NewFromFloat32(100.5), "9278923470").
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectExec("UPDATE users").
					WithArgs(decimal.NewFromFloat32(100.5), "user1").
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()
			},
			expectedErr: nil,
		},
		{
			name:  "transaction rollback on error",
			order: order,
			mock: func() {
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec("UPDATE orders").
					WithArgs("PROCESSED", decimal.NewFromFloat32(100.5), "9278923470").
					WillReturnError(errors.New("db error"))
				sqlMock.ExpectRollback()
			},
			expectedErr: fmt.Errorf("failed to update order: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := storage.UpdateOrderAndBalance(tt.order)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}
