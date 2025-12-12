package service

import (
	"github.com/annakonkova23/gophermart/internal/model"
)

func (as *AccumulationSystem) getBalance(user string) *model.Balance {
	balanceUser := &model.Balance{Balance: model.Zero()}
	balance, ok := as.balance.LoadOrStore(user, balanceUser)
	if ok {
		balanceUser = balance.(*model.Balance)
	}
	return balanceUser
}

func (as *AccumulationSystem) InitBalance() error {
	balances, err := as.database.GetAllBalanceDB()
	if err != nil {
		return err
	}
	for _, b := range balances {
		as.balance.Store(b.User, b)
	}
	return nil
}

func (as *AccumulationSystem) UpdateBalance(user string, balance *model.Balance) {
	as.balance.Store(user, balance)
}
