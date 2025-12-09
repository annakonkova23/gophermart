package service

import (
	"context"
	"sync"
	"time"

	"github.com/annakonkova23/gophermart/internal/config"
	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/annakonkova23/gophermart/internal/service/accrual"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type AccumulationSystem struct {
	users           sync.Map
	database        *sqlx.DB
	client          *accrual.AccrualClient
	newOrders       chan string
	currentOrder    sync.Map
	balance         sync.Map
	timeoutInterval time.Duration
}

func NewAccumulationSystem(ctx context.Context, database *sqlx.DB, cfg *config.Config) (*AccumulationSystem, error) {
	as := &AccumulationSystem{
		users:           sync.Map{},
		database:        database,
		client:          accrual.NewAccrualClient(ctx, cfg.AccrualAddress, cfg.CountProcess, cfg.Timeout, cfg.BufferSize),
		newOrders:       make(chan string, cfg.BufferSize),
		timeoutInterval: cfg.Timeout,
	}
	go as.monitorNewOrder(ctx)

	go as.startResultWorkers(ctx, as.client.Results(), 10)
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

func (as *AccumulationSystem) startResultWorkers(ctx context.Context, in <-chan *model.StatusOrder, workers int) {
	for i := 0; i < workers; i++ {
		go as.resultWorker(ctx, i+1, in)
	}
}

func (as *AccumulationSystem) resultWorker(ctx context.Context, id int, in <-chan *model.StatusOrder) {
	logrus.Infof("worker обработки полученного результата статуса заказа %d запущен", id)
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("worker обработки полученного результата статуса заказа %d остановлен (ctx done)", id)
			return

		case res, ok := <-in:
			if !ok {
				logrus.Infof("worker обработки полученного результата статуса заказа %d: канал закрыт", id)
				return
			}

			logrus.Infof("Получен результат для сохранения статуса, заказ %s, статус: %s", res.Number, res.Status)
			err := as.compareAndSaveStatusOrder(ctx, res)
			if err != nil {
				logrus.Errorln("Ошибка сохранения статуса заказа", err)
			}
		}
	}
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

func (as *AccumulationSystem) compareAndSaveStatusOrder(ctx context.Context, orderStatus *model.StatusOrder) error {
	logrus.Infoln("Обновление заказа:", orderStatus.Number)
	orderCurrent, ok := as.GetCurrentOrder(orderStatus.Number)
	if !ok {
		logrus.Errorf("Не найден заказ %s в системе", orderStatus.Number)
	}
	orderCurrent.Mx.Lock()
	defer orderCurrent.Mx.Unlock()
	var balanceUser *model.Balance
	if model.StatusesXmapCalc[orderStatus.Status] != orderCurrent.Status {
		orderCurrent.Status = model.StatusesXmapCalc[orderStatus.Status]
		orderCurrent.Accrual = &orderStatus.Accrual
		if model.StatusesCalcFinish[orderStatus.Status] {
			if orderStatus.Accrual.GreaterThan(model.Zero()) {
				logrus.Infof("Получен баланс %s для заказа %s:", orderStatus.Accrual.String(), orderStatus.Number)
				balanceUser = as.getBalance(orderCurrent.User)
				balanceUser.Mx.Lock()
				defer balanceUser.Mx.Unlock()
				balanceUser.Balance = balanceUser.Balance.Add(orderStatus.Accrual)
				as.UpdateBalance(orderCurrent.User, balanceUser)
			}
			as.saveEndStatusOrderDB(ctx, orderCurrent.User, orderCurrent, balanceUser)
			as.DeleteProcessedOrder(orderCurrent.Number)
		} else {
			err := as.saveStatusOrderDB(ctx, orderCurrent.User, orderCurrent, balanceUser)
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
	balance, ok := as.balance.Load(user)
	if !ok {
		logrus.Errorf("У пользователя %s нет баланса", user)
		return model.ErrorInsufficientFunds
	}

	balanceUser := balance.(*model.Balance)

	balanceUser.Mx.Lock()
	defer balanceUser.Mx.Unlock()

	if balanceUser.Balance.GreaterThanOrEqual(req.Sum) {
		balanceUser.Balance = balanceUser.Balance.Sub(req.Sum)
		balanceUser.Withdrawn = balanceUser.Withdrawn.Add(req.Sum)
		logrus.Infof("У пользователя %s списываем баллы %s", user, req.Sum.String())
		req.ProcessedAt = time.Now()
		err := as.saveBalanceAndWithdrawDB(ctx, user, balanceUser, req)
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
