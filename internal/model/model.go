package models

import "time"

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Hash     string `json:"hash,omitempty"`
}

type UserBalance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type Order struct {
	OrderUser   string    `json:"orderUser,omitempty"`
	OrderNumber string    `json:"number,omitempty"`
	Accrual     float32   `json:"accrual,omitempty"`
	Status      string    `json:"status,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at,omitempty"`
}

type Withdraw struct {
	OrderUser   string    `json:"orderUser,omitempty"`
	OrderNumber string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
}
