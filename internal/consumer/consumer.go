package consumer

import (
	"context"
	"sync"
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

type logger interface {
	CreateErrorLog(service, message string)
}

type OrderConsumer struct {
	logger        logger
	storage       Storage
	workers       int
	rep           Rep
	checkInterval int
}

func NewOrderConsumer(workers int, rep Rep, storage Storage, logger logger) *OrderConsumer {
	return &OrderConsumer{
		storage:       storage,
		workers:       workers,
		rep:           rep,
		checkInterval: checkInterval,
		logger:        logger,
	}
}

func (oc *OrderConsumer) ConsumerRun(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(oc.checkInterval) * time.Second)
	defer ticker.Stop()
	jobs := make(chan models.Order, 100)
	wg := &sync.WaitGroup{}
	for w := 1; w <= oc.workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			oc.ConsumerWorker(ctx, jobs)
		}()
	}
	go func() {
		defer close(jobs)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				orders, err := oc.storage.GetOrderNotFinish()
				if err != nil {
					oc.logger.CreateErrorLog("OrderConsumer", err.Error())
					continue
				}
				for _, m := range orders {
					jobs <- m
				}
			}
		}
	}()
	<-ctx.Done()

	wg.Wait()
}

func (oc *OrderConsumer) ConsumerWorker(ctx context.Context, statusOrders <-chan models.Order) {
	for {
		select {
		case <-ctx.Done():
			return
		case orders, ok := <-statusOrders:
			if !ok {
				return
			}
			ordersAccrual, err := oc.rep.GetInfoOrders(orders.OrderNumber)
			if err != nil {
				oc.logger.CreateErrorLog("OrderConsumer", err.Error())
			}
			if ordersAccrual.Status == orders.Status {
				continue
			}
			orders.Status = ordersAccrual.Status
			orders.Accrual = ordersAccrual.Accrual

			if err := oc.storage.UpdateOrderAndBalance(orders); err != nil {
				oc.logger.CreateErrorLog("OrderConsumer", err.Error())
			}
		}
	}
}
