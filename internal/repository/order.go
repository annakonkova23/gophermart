package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/sirupsen/logrus"
)

func (ds *DBStore) CreateOrderDB(userLogin, orderNumber string) error {

	var status string
	var errorMessage sql.NullString

	err := ds.database.QueryRowx(
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

func (ds *DBStore) GetOrderDB(userLogin string) ([]*model.Order, error) {
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

	err := ds.database.Select(&orders, query, userLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения списока заказов из db: %w", err)
	}

	return orders, nil
}

func (ds *DBStore) SaveStatusOrderDB(ctx context.Context, user string, order *model.Order, balance *model.Balance) error {

	tx, err := ds.database.BeginTxx(ctx, nil)
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

func (ds *DBStore) SaveEndStatusOrderDB(ctx context.Context, user string, order *model.Order, balance *model.Balance) {

	for {
		select {
		case <-ctx.Done():
			logrus.Errorf("Отмена контекста для сохранения конечного статуса")
		default:
			err := ds.SaveStatusOrderDB(ctx, user, order, balance)
			if err != nil {
				logrus.Errorf("Ошибка сохранения заказа %s в базу:%s", order.Number, err.Error())
				time.Sleep(ds.timeoutInterval)
			}
			return
		}
	}
}

func (ds *DBStore) LoadOrderNotProcessedDB() ([]*model.Order, error) {
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

	err := ds.database.Select(&orders, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения данных о необработанных заказах db: %w", err)
	}

	return orders, nil
}
