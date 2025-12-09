package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/jackc/pgconn"
	"github.com/sirupsen/logrus"
)

func (as *AccumulationSystem) сreateUserDB(user *model.User) error {
	query := `
		INSERT INTO accum_system.users (login, password)
		VALUES ($1, $2);
	`

	_, err := as.database.Exec(query, user.Login, user.Password)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return model.ErrorConflict
			}
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (as *AccumulationSystem) getUserByLoginDB(login string) (*model.User, error) {
	var loginDB, passwordDB string
	query := `SELECT login, password FROM accum_system.users WHERE login = $1`

	err := as.database.QueryRow(query, login).Scan(&loginDB, &passwordDB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with login '%s' not found", login)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &model.User{
		Login:    loginDB,
		Password: passwordDB,
	}, nil
}

func (as *AccumulationSystem) createOrderDB(userLogin, orderNumber string) error {

	var status string
	var errorMessage sql.NullString

	err := as.database.QueryRowx(
		`SELECT status, error_message FROM accum_system.create_order($1, $2)`,
		orderNumber,
		userLogin,
	).Scan(&status, &errorMessage)

	if err != nil {
		return fmt.Errorf("ошибка добавления в DB: %w", err)
	}

	switch status {
	case "success":
		return nil

	case "already_exists":
		return model.ErrorDoubleOperation

	case "conflict":
		return model.ErrorConflict

	default:
		if errorMessage.Valid {
			return fmt.Errorf("ошибка добавления в DB: %s", errorMessage.String)
		}
		return fmt.Errorf("ошибка добавления в DB: %s", status)
	}
}

func (as *AccumulationSystem) getOrderDB(userLogin string) ([]*model.Order, error) {
	var orders []*model.Order

	query := `
        SELECT 
            o.number AS number,
            os.status,
            os.accrual,
            os.uploaded_at
        FROM accum_system.orders o
        JOIN accum_system.orders_status os ON o.number = os.number
        WHERE o.user_login = $1
        ORDER BY os.uploaded_at DESC
    `

	err := as.database.Select(&orders, query, userLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения списока заказов из db: %w", err)
	}

	return orders, nil
}

func (as *AccumulationSystem) getBalanceDB(userLogin string) (*model.Balance, error) {
	var balance model.Balance

	query := `
        SELECT balance, withdrawn
        FROM accum_system.users_balance
        WHERE user_login = $1
    `

	err := as.database.Get(&balance, query, userLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения суммы накоплений и списаний из db: %w", err)
	}

	return &balance, nil
}

func (as *AccumulationSystem) getWithdrawalsDB(userLogin string) ([]*model.Withdraw, error) {
	var withdrawals []*model.Withdraw

	query := `
        SELECT  number,
    			sum,
    			processed_at
        FROM accum_system.orders_withdrawals
        WHERE user_login = $1
    `

	err := as.database.Select(&withdrawals, query, userLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения сумму накоплений и списаний из db: %w", err)
	}

	return withdrawals, nil
}

func (as *AccumulationSystem) saveStatusOrderDB(ctx context.Context, user string, order *model.Order, balance *model.Balance) error {

	tx, err := as.database.BeginTxx(ctx, nil)
	if err != nil {
		msg := fmt.Sprintf("Ошибка создания транзакции: %s", err.Error())
		return errors.New(msg)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	query := `
        INSERT INTO accum_system.orders_status (number, status, accrual, uploaded_at)
        VALUES ($1, $2, $3, NOW())
        ON CONFLICT (number) DO UPDATE
        SET 
            status = EXCLUDED.status,
            accrual = EXCLUDED.accrual,
            uploaded_at = NOW()
    `

	_, err = tx.ExecContext(ctx, query,
		order.Number,
		order.Status,
		order.Accrual,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса: %w", err)
	}
	if balance != nil {
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
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("ошибка коммита транзакции: %w", err)
	}

	return nil

}

func (as *AccumulationSystem) saveEndStatusOrderDB(ctx context.Context, user string, order *model.Order, balance *model.Balance) {

	for {
		select {
		case <-ctx.Done():
			logrus.Errorf("Отмена контекста для сохранения конечного статуса")
		default:
			err := as.saveStatusOrderDB(ctx, user, order, balance)
			if err != nil {
				logrus.Errorf("Ошибка сохранения заказа %s в базу:%s", order.Number, err.Error())
				time.Sleep(as.timeoutInterval)
			}
			return
		}
	}
}

func (as *AccumulationSystem) saveBalanceAndWithdrawDB(ctx context.Context, user string, balance *model.Balance, withdraw *model.Withdraw) error {
	tx, err := as.database.BeginTxx(ctx, nil)
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

func (as *AccumulationSystem) loadOrderNotProcessedDB() ([]*model.Order, error) {
	var orders []*model.Order

	query := `
        SELECT 
            o.number AS number,
            os.status,
            os.accrual,
            os.uploaded_at,
			o.user_login
        FROM accum_system.orders o
        JOIN accum_system.orders_status os ON o.number = os.number
        WHERE status not in ('PROCESSED', 'INVALID')
    `

	err := as.database.Select(&orders, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения данных о необработанных заказах db: %w", err)
	}

	return orders, nil
}

func (as *AccumulationSystem) getAllBalanceDB() ([]*model.Balance, error) {
	var balances []*model.Balance

	query := `
        SELECT balance, withdrawn, user_login 
        FROM accum_system.users_balance
    `

	err := as.database.Select(&balances, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения балансов пользователей db: %w", err)
	}

	return balances, nil
}
