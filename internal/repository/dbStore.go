package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type DBStore struct {
	database        *sqlx.DB
	timeoutInterval time.Duration
}

func NewDBStore(db *sqlx.DB, timeout time.Duration) *DBStore {
	return &DBStore{
		database:        db,
		timeoutInterval: timeout,
	}
}
