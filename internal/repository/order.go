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

	err := ds.database.QueryRowx(execCreateOrder,
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

	err := ds.database.Select(&orders, selectOrderStatus, userLogin)
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

	_, err = tx.ExecContext(ctx, updateOrderStatus,
		order.Number,
		order.Status,
		order.Accrual,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса: %w", err)
	}
	if balance != nil {
		_, err = tx.ExecContext(ctx, updateUserBalance, user, balance.Balance, balance.Withdrawn)
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

	err := ds.database.Select(&orders, selectNewOrderStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения данных о необработанных заказах db: %w", err)
	}

	return orders, nil
}
