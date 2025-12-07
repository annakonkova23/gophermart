package service

import "github.com/annakonkova23/gophermart/internal/model"

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
