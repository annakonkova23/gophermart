package service

import (
	"context"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/annakonkova23/gophermart/internal/config"
	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/annakonkova23/gophermart/internal/service/accrual"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

var loginRegexp = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,32}$`)

type AccumulationSystem struct {
	users     sync.Map
	database  *sqlx.DB
	client    *accrual.AccrualClient
	newOrders chan string
}

func NewAccumulationSystem(ctx context.Context, database *sqlx.DB, cfg config.Config) *AccumulationSystem {
	as := &AccumulationSystem{
		users:     sync.Map{},
		database:  database,
		client:    accrual.NewAccrualClient(ctx, cfg.AccrualAddress, 10, 10*time.Second, 1000),
		newOrders: make(chan string, 1000),
	}

	go as.monitorNewOrder(ctx)
	go as.monitorResults(ctx)
}

func (as *AccumulationSystem) NewUser(user *model.User) error {
	if user.Empty() {
		return model.ErrorIncorrectRequest
	}

	err := as.LoginValid(user.Login)
	if err != nil {
		return model.ErrorIncorrectRequest
	}

	err = user.HashPassword()
	if err != nil {
		return err
	}

	err = as.сreateUserDB(user)
	if err != nil {
		return err
	}
	return nil
}

func (as *AccumulationSystem) AuthUser(user *model.User) error {
	if user.Empty() {
		return model.ErrorIncorrectRequest
	}

	userCurrent, err := as.getUserByLoginDB(user.Login)
	if err != nil {
		return err
	}

	if userCurrent.Equals(user) {
		return nil
	} else {
		return model.ErrorNotEqual
	}

}

func (as *AccumulationSystem) NewOrder(user, number string) error {

	_, err := as.getUserByLoginDB(user)
	if err != nil {
		return model.ErrorNotAuthorization
	}

	if !as.LuhnValid(number) {
		return model.ErrorNotValidNumber
	}
	err = as.createOrderDB(user, number)
	if err != nil {
		return err
	}
	return nil
}

func (as *AccumulationSystem) GetOrders(user string) ([]*model.Order, error) {

	_, err := as.getUserByLoginDB(user)
	if err != nil {
		return nil, model.ErrorNotAuthorization
	}

	orders, err := as.getOrderDB(user)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (as *AccumulationSystem) GetBalance(user string) (*model.Balance, error) {

	_, err := as.getUserByLoginDB(user)
	if err != nil {
		return nil, model.ErrorNotAuthorization
	}

	balance, err := as.getBalanceDB(user)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (as *AccumulationSystem) WithdrawBonus(user string, req *model.Withdraw) error {

	_, err := as.getUserByLoginDB(user)
	if err != nil {
		return model.ErrorNotAuthorization
	}

	if !as.LuhnValid(req.OrderNumber) {
		return model.ErrorNotValidNumber
	}

	/*balance, err := as.getBalanceDB(user)
	if err != nil {
		return nil, err
	}*/
	/*Написать списание средств*/
	return nil
}

func (as *AccumulationSystem) GetWithdrawals(user string) ([]*model.Withdraw, error) {

	_, err := as.getUserByLoginDB(user)
	if err != nil {
		return nil, model.ErrorNotAuthorization
	}

	withdrawals, err := as.getWithdrawalsDB(user)
	if err != nil {
		return nil, err
	}
	sort.Slice(withdrawals, func(i, j int) bool {
		return withdrawals[i].ProcessedAt.After(withdrawals[j].ProcessedAt)
	})

	return withdrawals, nil
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

func (as *AccumulationSystem) monitorResults(ctx context.Context) {
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
