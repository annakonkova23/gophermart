package service

import (
	"context"
	"sort"
	"time"

	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/sirupsen/logrus"
)

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
	as.AddCurrentOrder(&model.Order{Number: number, Status: model.StatusNew, UploadedAt: time.Now(), User: user})
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
	logrus.Infof("Запрос на списание пользователь %s, заказ %s, сумма %s", user, req.OrderNumber, req.Sum.String())
	_, err := as.getUserByLoginDB(user)
	if err != nil {
		return model.ErrorNotAuthorization
	}

	if !as.LuhnValid(req.OrderNumber) {
		return model.ErrorNotValidNumber
	}
	err = as.withdrawBalance(context.Background(), user, req)
	if err != nil {
		return err
	}
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
