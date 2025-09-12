package decimal

import "github.com/shopspring/decimal"

type Decimal = decimal.Decimal

func NewFromFloat(f float64) Decimal {
	return decimal.NewFromFloat(f)
}

func NewFromInt(i int64) Decimal {
	return decimal.NewFromInt(i)
}
