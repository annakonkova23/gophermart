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

	query := `
        SELECT balance, withdrawn
        FROM accum_system.users_balance
        WHERE user_login = $1
    `

	err := ds.database.Get(&balance, query, userLogin)
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

	query := `
        SELECT  number,
    			sum,
    			processed_at
        FROM accum_system.orders_withdrawals
        WHERE user_login = $1
    `

	err := ds.database.Select(&withdrawals, query, userLogin)
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

	_, err = tx.ExecContext(ctx, `
        INSERT INTO accum_system.orders_withdrawals (user_login, number, sum, processed_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (number) DO NOTHING
    `, user, withdraw.OrderNumber, withdraw.Sum, withdraw.ProcessedAt)
	if err != nil {
		return fmt.Errorf("ошибка вставки в accum_system.orders_withdrawals: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
        INSERT INTO accum_system.users_balance (user_login, balance, withdrawn)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_login) DO UPDATE
        SET 
            balance = EXCLUDED.balance,
            withdrawn = EXCLUDED.withdrawn
    `, user, balance.Balance, balance.Withdrawn)
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

	query := `
        SELECT balance, withdrawn, user_login 
        FROM accum_system.users_balance
    `

	err := ds.database.Select(&balances, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения балансов пользователей db: %w", err)
	}

	return balances, nil
}
