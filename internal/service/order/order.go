package order

import (
	"fmt"
	"strconv"

	models "github.com/AleGaliev/gofermart/internal/model"
	"github.com/shopspring/decimal"
)

type Storage interface {
	UploadOrder(OrderUser, OrderNumber string) error
	GetOrder(OrderNumber string) (string, error)
	GetUserBalance(user string) (models.UserBalanceOut, error)
	UploadOrderWithdraw(user string, withdraw models.Withdraw) error
}

func UploadOrder(user, orderNumber string, storage Storage) error {
	orderNumberInt, err := strconv.Atoi(orderNumber)
	if err != nil {
		return fmt.Errorf("not valid order number")
	}
	isValidOrder := orderLuhnCheck(orderNumberInt)
	if !isValidOrder {
		return fmt.Errorf("not valid order number")
	}

	orderUser, err := storage.GetOrder(orderNumber)
	if err != nil {
		return err
	}

	if orderUser == "" {
		if err := storage.UploadOrder(user, orderNumber); err != nil {
			return err
		}
		return nil
	}

	if orderUser != user {
		return fmt.Errorf("invalid order user")
	}

	return fmt.Errorf("the order number was already taken by this user")
}

func UploadOrderWithdraw(user string, withdraw models.Withdraw, storage Storage) error {
	orderNumberInt, err := strconv.Atoi(withdraw.OrderNumber)
	if err != nil {
		return fmt.Errorf("not valid order number")
	}

	isValidOrder := orderLuhnCheck(orderNumberInt)
	if !isValidOrder {
		return fmt.Errorf("not valid order number")
	}

	balance, err := storage.GetUserBalance(user)
	if err != nil {
		return err
	}

	balanceCurrentDecimal := decimal.NewFromFloat32(balance.Current)

	if balanceCurrentDecimal.LessThan(withdraw.Sum) {
		return fmt.Errorf("insufficient funds")
	}

	orderUser, err := storage.GetOrder(withdraw.OrderNumber)
	if err != nil {
		return err
	}

	if orderUser == "" {
		if err := storage.UploadOrderWithdraw(user, withdraw); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("the order number exists")

}

func orderLuhnCheck(number int) bool {
	if number < 10 {
		return false
	}

	sum := 0
	double := false

	for number > 0 {
		digit := number % 10
		number /= 10

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return sum%10 == 0
}
