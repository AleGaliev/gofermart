package consumer

import (
	"context"
	"fmt"
	"log"
	"time"

	models "github.com/AleGaliev/gofermart/internal/model"
)

const (
	checkInterval int = 1
)

type Rep interface {
	GetInfoOrders(orders string) (models.Order, error)
}

type Storage interface {
	UpdateOrderAndBalance(order models.Order) error
	GetOrderNotFinish() ([]models.Order, error)
}

type OrderConsumer struct {
	storage       Storage
	workers       int
	rep           Rep
	checkInterval int
}

func NewOrderConsumer(workers int, rep Rep, storage Storage) *OrderConsumer {
	return &OrderConsumer{
		storage:       storage,
		workers:       workers,
		rep:           rep,
		checkInterval: checkInterval,
	}
}

func (oc *OrderConsumer) ConsumerRun(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(oc.checkInterval) * time.Second)
	defer ticker.Stop()
	errCh := make(chan error, 1)
	jobs := make(chan models.Order, 100)
	for w := 1; w <= oc.workers; w++ {
		go func() {
			if err := oc.ConsumerWorker(jobs); err != nil {
				errCh <- fmt.Errorf("consumer worker failed: %w", err)
				log.Printf("Consumer worker failed: %v", err)
			}
		}()
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				orders, _ := oc.storage.GetOrderNotFinish()
				for _, m := range orders {
					jobs <- m
				}
			}
		}
	}()
	<-ctx.Done()
	return nil

}

func (oc *OrderConsumer) ConsumerWorker(statusOrders <-chan models.Order) error {
	for orders := range statusOrders {

		ordersAccrual, err := oc.rep.GetInfoOrders(orders.OrderNumber)
		if err != nil {
			fmt.Println(err)
		}

		if ordersAccrual.Status == orders.Status {
			continue
		}
		orders.Status = ordersAccrual.Status
		orders.Accrual = ordersAccrual.Accrual

		if err := oc.storage.UpdateOrderAndBalance(orders); err != nil {
			fmt.Println(err)
		}

	}
	return nil
}
