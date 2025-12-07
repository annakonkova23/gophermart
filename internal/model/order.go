package model

import (
	"sync"
	"time"

	"github.com/shopspring/decimal"
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
	Number     string           `db:"number" json:"number"`
	Status     string           `db:"status" json:"status"`
	Accrual    *decimal.Decimal `db:"accrual" json:"accrual,omitempty"`
	UploadedAt time.Time        `db:"uploaded_at" json:"uploaded_at"`
	User       string           `json:"-"`
	Mx         sync.RWMutex     `json:"-"`
}

type Balance struct {
	Balance   decimal.Decimal `json:"current" db:"balance"`
	Withdrawn decimal.Decimal `json:"withdrawn" db:"withdrawn"`
	Mx        sync.RWMutex    `json:"-" db:"-"`
}

type Withdraw struct {
	OrderNumber string          `json:"order" db:"number"`
	Sum         decimal.Decimal `json:"sum" db:"sum"`
	ProcessedAt time.Time       `json:"processed_at,omitempty" db:"processed_at"`
}

type StatusOrder struct {
	Number  string          `json:"order"`
	Status  string          `json:"status"`
	Accrual decimal.Decimal `json:"accrual,omitempty"`
}

type UserOrder struct {
	User  string
	Order string
}
