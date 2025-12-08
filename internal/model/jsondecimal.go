package model

import (
	"strings"

	"github.com/shopspring/decimal"
)

type JSONDecimal struct {
	decimal.Decimal
}

func (d JSONDecimal) MarshalJSON() ([]byte, error) {
	// пишем число как токен JSON без кавычек
	return []byte(d.Decimal.String()), nil
}

func (d *JSONDecimal) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	dec, err := decimal.NewFromString(s)
	if err != nil {
		return err
	}
	d.Decimal = dec
	return nil
}

func (d JSONDecimal) Add(other JSONDecimal) JSONDecimal {
	return JSONDecimal{Decimal: d.Decimal.Add(other.Decimal)}
}

func (d JSONDecimal) Sub(other JSONDecimal) JSONDecimal {
	return JSONDecimal{Decimal: d.Decimal.Sub(other.Decimal)}
}

func (d JSONDecimal) GreaterThan(other JSONDecimal) bool {
	return d.Decimal.GreaterThan(other.Decimal)
}

func (d JSONDecimal) GreaterThanOrEqual(other JSONDecimal) bool {
	return d.Decimal.GreaterThanOrEqual(other.Decimal)
}

func Zero() JSONDecimal {
	return JSONDecimal{Decimal: decimal.NewFromInt(0)}
}
