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
	StatusProcessing:     StatusProcessing,
}

var StatusesCalcFinish = map[string]bool{
	"REGISTERED": false,
	"PROCESSED":  true,
	"INVALID":    true,
	"PROCESSING": false,
}

type Order struct {
	Number     string          `json:"number"`
	Status     string          `json:"status"`
	Accrual    decimal.Decimal `json:"accrual,omitempty"`
	UploadedAt time.Time       `json:"uploaded_at"`
	User       string          `json:"-"`
	Mx         sync.RWMutex    `json:"-"`
}

type Balance struct {
	Balance   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
	Mx        sync.RWMutex    `json:"-"`
}

type Withdraw struct {
	OrderNumber string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at,omitempty"`
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
