package service

import (
	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/sirupsen/logrus"
)

func (as *AccumulationSystem) AddCurrentOrder(order *model.Order) {
	as.currentOrder.Store(order.Number, order)
	as.newOrders <- order.Number
}

func (as *AccumulationSystem) GetCurrentOrder(number string) (*model.Order, bool) {
	order, ok := as.currentOrder.Load(number)
	if !ok {
		return nil, ok
	}
	return order.(*model.Order), ok
}

func (as *AccumulationSystem) DeleteProcessedOrder(number string) {
	as.currentOrder.Delete(number)
}

func (as *AccumulationSystem) UpdateCurrentOrder(number string, order *model.Order) {
	as.currentOrder.Store(order.Number, order)
}

func (as *AccumulationSystem) InitOrders() error {
	logrus.Infoln("Инициализация необработанных заказов")
	orders, err := as.database.LoadOrderNotProcessedDB()
	if err != nil {
		return err
	}
	for _, o := range orders {
		as.AddCurrentOrder(o)
	}
	return nil
}
