package service

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/jackc/pgconn"
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

	var status, errorMessage string

	// Вызываем функцию
	err := as.database.QueryRowx(
		`SELECT status, error_message FROM accum_system.create_order($1, $2)`,
		orderNumber,
		userLogin,
	).Scan(&status, &errorMessage)

	if err != nil {
		return fmt.Errorf("ошибка добавления в DB: %w", err)
	}

	// Обрабатываем статус
	switch status {
	case "success":
		return nil

	case "already_exists":
		return model.ErrorDoubleOperation

	case "conflict":
		return model.ErrorConflict

	default:
		if errorMessage != "" {
			return fmt.Errorf("ошибка добавления в DB: %s", errorMessage)
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
		// Обработка ошибки: нет данных, ошибка подключения и т.д.
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения списока заказов из db: %w", err)
	}

	return orders, nil
}

func (as *AccumulationSystem) getBalanceDB(userLogin string) (*model.Balance, error) {
	var balance *model.Balance

	query := `
        SELECT balance, withdrawn
        FROM accum_system.users_balance
        WHERE user_login = $1
    `

	err := as.database.Select(&balance, query, userLogin)
	if err != nil {
		// Обработка ошибки: нет данных, ошибка подключения и т.д.
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения сумму накоплений и списаний из db: %w", err)
	}

	return balance, nil
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
		// Обработка ошибки: нет данных, ошибка подключения и т.д.
		if err == sql.ErrNoRows {
			return nil, model.ErrorNotContent
		}
		return nil, fmt.Errorf("ошибка получения сумму накоплений и списаний из db: %w", err)
	}

	return withdrawals, nil
}
