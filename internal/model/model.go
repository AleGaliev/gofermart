package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Hash     string `json:"hash,omitempty"`
}

type UserBalance struct {
	Current   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

type UserBalanceOut struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type Order struct {
	OrderUser   string          `json:"orderUser,omitempty"`
	OrderNumber string          `json:"number,omitempty"`
	Accrual     decimal.Decimal `json:"accrual,omitempty"`
	Status      string          `json:"status,omitempty"`
	UploadedAt  time.Time       `json:"uploaded_at,omitempty"`
}

type OrderOut struct {
	OrderUser   string    `json:"orderUser,omitempty"`
	OrderNumber string    `json:"number,omitempty"`
	Accrual     float32   `json:"accrual,omitempty"`
	Status      string    `json:"status,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at,omitempty"`
}

type WithdrawOut struct {
	OrderUser   string    `json:"orderUser,omitempty"`
	OrderNumber string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
}
type Withdraw struct {
	OrderUser   string          `json:"orderUser,omitempty"`
	OrderNumber string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at,omitempty"`
}
