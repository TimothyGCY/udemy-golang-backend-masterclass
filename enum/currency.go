package enum

const (
	USD = "USD"
	MYR = "MYR"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, MYR:
		return true
	}
	return false
}

func SupportedCurrencies() []string {
	return []string{USD, MYR}
}
