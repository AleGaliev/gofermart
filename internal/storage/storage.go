package storage

import models "github.com/AleGaliev/gofermart/internal/model"

type Storage interface {
	Register(user models.User) error
	Login(user models.User) error
	UploadOrder(orderUser, orderNumber string) error
	GetOrder(OrderNumber string) (string, error)
	GetOrdersUser(user string) ([]models.Order, error)
	GetUserBalance(user string) (models.UserBalance, error)
	PointsDebiting(user string, sum float32) error
	WriteWithdrawals(user string, withdraw models.Withdraw) error
	GetUserWithdrawals(user string) ([]models.Withdraw, error)
	UpdateOrderAndBalance(order models.Order) error
	GetOrderNotFinish() ([]models.Order, error)
}
