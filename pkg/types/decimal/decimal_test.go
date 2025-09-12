package decimal

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDecimalHandlingForPayments(t *testing.T) {
	t.Run("Creating small decimal values", func(t *testing.T) {
		// Create a decimal representing 0.001
		smallAmount := decimal.NewFromFloat(0.001)
		assert.Equal(t, "0.001", smallAmount.String())

		// Alternative way to create precise values
		smallAmountFromString := decimal.RequireFromString("0.001")
		assert.Equal(t, smallAmount, smallAmountFromString)
	})

	t.Run("Adding small decimal values", func(t *testing.T) {
		// 0.001 + 0.002 = 0.003
		a := decimal.NewFromFloat(0.001)
		b := decimal.NewFromFloat(0.002)
		sum := a.Add(b)

		assert.Equal(t, "0.003", sum.String())
	})

	t.Run("Multiplying small values", func(t *testing.T) {
		// 0.001 * 3 = 0.003
		price := decimal.NewFromFloat(0.001)
		quantity := decimal.NewFromInt(3)
		total := price.Mul(quantity)

		assert.Equal(t, "0.003", total.String())
	})

	t.Run("Percentage calculations", func(t *testing.T) {
		// Calculate 5% of 0.001
		amount := decimal.NewFromFloat(0.001)
		fivePercent := decimal.NewFromFloat(0.05)
		result := amount.Mul(fivePercent)

		assert.Equal(t, "0.00005", result.String())
	})

	t.Run("Rounding small values", func(t *testing.T) {
		// Round 0.0015 to 2 decimal places (should be 0.00)
		smallValue := decimal.NewFromFloat(0.0015)
		rounded := smallValue.Round(2)

		assert.Equal(t, "0", rounded.String())

		// Round 0.0015 to 3 decimal places (should be 0.002)
		roundedToThree := smallValue.Round(3)
		assert.Equal(t, "0.002", roundedToThree.String())
	})

	t.Run("Format with fixed decimal places", func(t *testing.T) {
		// Format 0.001 with exactly 4 decimal places
		amount := decimal.NewFromFloat(0.001)
		formatted := amount.StringFixed(4)

		assert.Equal(t, "0.0010", formatted)
	})

	t.Run("Comparing small values", func(t *testing.T) {
		a := decimal.NewFromFloat(0.001)
		b := decimal.NewFromFloat(0.0010)
		c := decimal.NewFromFloat(0.00101)

		assert.True(t, a.Equal(b))
		assert.True(t, a.LessThan(c))
		assert.True(t, c.GreaterThan(a))
	})

	t.Run("Currency conversion with small rates", func(t *testing.T) {
		// Convert 10 USD to a currency with a 0.001 exchange rate
		amountUSD := decimal.NewFromFloat(10.00)
		exchangeRate := decimal.NewFromFloat(0.001)
		convertedAmount := amountUSD.Mul(exchangeRate)

		assert.Equal(t, "0.01", convertedAmount.String())
	})
}
