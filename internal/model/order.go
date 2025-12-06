package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	StatusNew        = "NEW"
	StatusProcessed  = "PROCESSED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
)

var StatusesCalc = map[string]bool{
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
}

type Balance struct {
	Balance   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
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
