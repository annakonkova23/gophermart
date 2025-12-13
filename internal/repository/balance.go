package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/annakonkova23/gophermart/internal/model"
)

func (ds *DBStore) GetBalanceDB(userLogin string) (*model.Balance, error) {
	var balance model.Balance

	err := ds.database.Get(&balance, selectBalanceUser, userLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения суммы накоплений и списаний из db: %w", err)
	}

	return &balance, nil
}

func (ds *DBStore) GetWithdrawalsDB(userLogin string) ([]*model.Withdraw, error) {
	var withdrawals []*model.Withdraw

	err := ds.database.Select(&withdrawals, selectWithdrawals, userLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения сумму накоплений и списаний из db: %w", err)
	}

	return withdrawals, nil
}

func (ds *DBStore) SaveBalanceAndWithdrawDB(ctx context.Context, user string, balance *model.Balance, withdraw *model.Withdraw) error {
	tx, err := ds.database.BeginTxx(ctx, nil)
	if err != nil {
		msg := fmt.Sprintf("ошибка создания транзакции: %s", err.Error())
		return errors.New(msg)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.ExecContext(ctx, insertwithdrawals, user, withdraw.OrderNumber, withdraw.Sum, withdraw.ProcessedAt)
	if err != nil {
		return fmt.Errorf("ошибка вставки в accum_system.orders_withdrawals: %w", err)
	}
	_, err = tx.ExecContext(ctx, updateUserBalance, user, balance.Balance, balance.Withdrawn)
	if err != nil {
		return fmt.Errorf(" ошибка вставки в accum_system.users_balance: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("ошибка коммита транзакции: %w", err)
	}

	return nil
}

func (ds *DBStore) GetAllBalanceDB() ([]*model.Balance, error) {
	var balances []*model.Balance

	err := ds.database.Select(&balances, selectAllBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения балансов пользователей db: %w", err)
	}

	return balances, nil
}
