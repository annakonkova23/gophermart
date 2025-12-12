package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/annakonkova23/gophermart/internal/model"
	"github.com/jackc/pgconn"
)

func (ds *DBStore) CreateUserDB(user *model.User) error {
	query := `
		INSERT INTO accum_system.users (login, password)
		VALUES ($1, $2);
	`

	_, err := ds.database.Exec(query, user.Login, user.Password)
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

func (ds *DBStore) GetUserByLoginDB(login string) (*model.User, error) {
	var loginDB, passwordDB string
	query := `SELECT login, password FROM accum_system.users WHERE login = $1`

	err := ds.database.QueryRow(query, login).Scan(&loginDB, &passwordDB)
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
