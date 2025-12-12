package model

import (
	"sync"
	"time"
)

type Balance struct {
	Balance   JSONDecimal  `json:"current" db:"balance"`
	Withdrawn JSONDecimal  `json:"withdrawn" db:"withdrawn"`
	User      string       `db:"user_login" json:"-"`
	mx        sync.RWMutex `json:"-" db:"-"`
}

type Withdraw struct {
	OrderNumber string      `json:"order" db:"number"`
	Sum         JSONDecimal `json:"sum" db:"sum"`
	ProcessedAt time.Time   `json:"processed_at,omitempty" db:"processed_at"`
}

func (o *Balance) Lock() {
	o.mx.Lock()
}

func (o *Balance) Unlock() {
	o.mx.Unlock()
}

func (o *Balance) RLock() {
	o.mx.RLock()
}

func (o *Balance) RUnlock() {
	o.mx.RUnlock()
}
