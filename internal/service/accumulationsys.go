package service

import (
	"context"
	"sync"
	"time"

	"github.com/annakonkova23/gophermart/internal/client"
	"github.com/annakonkova23/gophermart/internal/config"
	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/annakonkova23/gophermart/internal/repository"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type AccumulationSystem struct {
	users        sync.Map
	database     *repository.DBStore
	client       *client.AccrualClient
	newOrders    chan string
	currentOrder sync.Map
	balance      sync.Map
}

func NewAccumulationSystem(ctx context.Context, database *sqlx.DB, cfg *config.Config) (*AccumulationSystem, error) {
	as := &AccumulationSystem{
		users:     sync.Map{},
		database:  repository.NewDBStore(database, cfg.Timeout),
		client:    client.NewAccrualClient(ctx, cfg.AccrualAddress, cfg.CountProcess, cfg.Timeout, cfg.BufferSize),
		newOrders: make(chan string, cfg.BufferSize),
	}
	go as.monitorNewOrder(ctx)

	err := as.InitBalance()
	if err != nil {
		return nil, err
	}
	err = as.InitOrders()
	if err != nil {
		return nil, err
	}
	return as, nil
}

func (as *AccumulationSystem) GetInputChan() <-chan *model.StatusOrder {
	return as.client.Results()
}

func (as *AccumulationSystem) monitorNewOrder(ctx context.Context) {
	logrus.Infoln("Старт мониторинга новых заказов")
	for {
		select {
		case <-ctx.Done():
			logrus.Infoln("Прерывание мониторинга новых заказов")
		case order := <-as.newOrders:
			as.client.AddOrder(ctx, order)
		}
	}
}

func (as *AccumulationSystem) CompareAndSaveStatusOrder(ctx context.Context, orderStatus *model.StatusOrder) error {
	logrus.Infoln("Обновление заказа:", orderStatus.Number)
	orderCurrent, ok := as.GetCurrentOrder(orderStatus.Number)
	if !ok {
		logrus.Errorf("Не найден заказ %s в системе", orderStatus.Number)
	}
	orderCurrent.Lock()
	defer orderCurrent.Unlock()
	var balanceUser *model.Balance
	if model.StatusesXmapCalc[orderStatus.Status] != orderCurrent.Status {
		orderCurrent.Status = model.StatusesXmapCalc[orderStatus.Status]
		orderCurrent.Accrual = &orderStatus.Accrual
		if model.StatusesCalcFinish[orderStatus.Status] {
			if orderStatus.Accrual.GreaterThan(model.Zero()) {
				logrus.Infof("Получен баланс %s для заказа %s:", orderStatus.Accrual.String(), orderStatus.Number)
				balanceUser = as.getBalance(orderCurrent.User)
				balanceUser.Lock()
				defer balanceUser.Unlock()
				balanceUser.Balance = balanceUser.Balance.Add(orderStatus.Accrual)
				as.UpdateBalance(orderCurrent.User, balanceUser)
			}
			as.database.SaveEndStatusOrderDB(ctx, orderCurrent.User, orderCurrent, balanceUser)
			as.DeleteProcessedOrder(orderCurrent.Number)
		} else {
			err := as.database.SaveStatusOrderDB(ctx, orderCurrent.User, orderCurrent, balanceUser)
			if err != nil {
				logrus.Errorf("Ошибка сохранения статуса заказа %s в базу:%s", orderCurrent.Number, err.Error())
				return err
			}
			as.UpdateCurrentOrder(orderCurrent.Number, orderCurrent)
		}

	} else {
		logrus.Infof("Статус заказа %s не обновился", orderStatus.Number)
	}
	return nil
}

func (as *AccumulationSystem) withdrawBalance(ctx context.Context, user string, req *model.Withdraw) error {
	balanceUser := as.getBalance(user)

	balanceUser.Lock()
	defer balanceUser.Unlock()

	if balanceUser.Balance.GreaterThanOrEqual(req.Sum) {
		balanceUser.Balance = balanceUser.Balance.Sub(req.Sum)
		balanceUser.Withdrawn = balanceUser.Withdrawn.Add(req.Sum)
		logrus.Infof("У пользователя %s списываем баллы %s", user, req.Sum.String())
		req.ProcessedAt = time.Now()
		err := as.database.SaveBalanceAndWithdrawDB(ctx, user, balanceUser, req)
		if err != nil {
			return err
		}
		as.UpdateBalance(user, balanceUser)
	} else {
		logrus.Errorf("У пользователя %s недостаточно средств", user)
		return model.ErrorInsufficientFunds
	}
	return nil
}
