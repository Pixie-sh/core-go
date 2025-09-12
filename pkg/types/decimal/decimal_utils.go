package decimal

// CalculateFee calculates a 0.1% transaction fee
func CalculateFee(amount Decimal, fee Decimal) Decimal {
	return amount.Mul(fee)
}

// WithCurrency formats a decimal for display with currency symbol
func WithCurrency(amount Decimal, currency string, floatingPoint int32) string {
	formattedAmount := amount.StringFixed(floatingPoint)

	switch currency {
	case "USD":
		return "$" + formattedAmount
	case "EUR":
		return "€" + formattedAmount
	default:
		return formattedAmount + " " + currency
	}
}

// Split divides a payment among multiple recipients
func Split(totalAmount Decimal, shares int64) []Decimal {
	if shares <= 0 {
		return []Decimal{}
	}

	shareCount := NewFromInt(shares)
	shareAmount := totalAmount.Div(shareCount)

	// Handle remaining fraction (due to division)
	remainder := totalAmount.Sub(shareAmount.Mul(shareCount))

	result := make([]Decimal, shares)
	for i := int64(0); i < shares; i++ {
		result[i] = shareAmount
	}

	// Add remainder to first share to ensure total sum is exactly the original amount
	if !remainder.IsZero() {
		result[0] = result[0].Add(remainder)
	}

	return result
}
