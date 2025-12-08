package model

import (
	"sync"
	"time"
)

const (
	StatusNew        = "NEW"
	StatusProcessed  = "PROCESSED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
)

const (
	StatusCalcRegistered = "REGISTERED"
	StatusCalcProcessed  = "PROCESSED"
	StatusCalcInvalid    = "INVALID"
	StatusCalcProcessing = "PROCESSING"
)

var StatusesXmapCalc = map[string]string{
	StatusCalcRegistered: StatusProcessing,
	StatusCalcProcessed:  StatusProcessed,
	StatusCalcInvalid:    StatusInvalid,
	StatusCalcProcessing: StatusProcessing,
}

var StatusesCalcFinish = map[string]bool{
	"REGISTERED": false,
	"PROCESSED":  true,
	"INVALID":    true,
	"PROCESSING": false,
}

type Order struct {
	Number     string       `db:"number" json:"number"`
	Status     string       `db:"status" json:"status"`
	Accrual    *JSONDecimal `db:"accrual" json:"accrual,omitempty"`
	UploadedAt time.Time    `db:"uploaded_at" json:"uploaded_at"`
	User       string       `db:"user_login" json:"-"`
	Mx         sync.RWMutex `json:"-"`
}

type Balance struct {
	Balance   JSONDecimal  `json:"current" db:"balance"`
	Withdrawn JSONDecimal  `json:"withdrawn" db:"withdrawn"`
	User      string       `db:"user_login" json:"-"`
	Mx        sync.RWMutex `json:"-" db:"-"`
}

type Withdraw struct {
	OrderNumber string      `json:"order" db:"number"`
	Sum         JSONDecimal `json:"sum" db:"sum"`
	ProcessedAt time.Time   `json:"processed_at,omitempty" db:"processed_at"`
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
