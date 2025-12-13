package model

import (
	"sync"
	"time"
)

type Order struct {
	Number     string       `db:"number" json:"number"`
	Status     string       `db:"status" json:"status"`
	Accrual    *JSONDecimal `db:"accrual" json:"accrual,omitempty"`
	UploadedAt time.Time    `db:"uploaded_at" json:"uploaded_at"`
	User       string       `db:"user_login" json:"-"`
	mx         sync.RWMutex `json:"-"`
}

func (o *Order) Lock() {
	o.mx.Lock()
}

func (o *Order) Unlock() {
	o.mx.Unlock()
}

func (o *Order) RLock() {
	o.mx.RLock()
}

func (o *Order) RUnlock() {
	o.mx.RUnlock()
}

type StatusOrder struct {
	Number  string      `json:"order"`
	Status  string      `json:"status"`
	Accrual JSONDecimal `json:"accrual,omitempty"`
}

type UserOrder struct {
	User  string
	Order string
}
