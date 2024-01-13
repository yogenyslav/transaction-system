package service

import (
	"accountservice/internal/errs"
	"strings"
)

var fakeCurrencyRates map[string]float64

func init() {
	fakeCurrencyRates = map[string]float64{
		"rub": 1.0,
		"usd": 88.38,
		"eur": 96.81,
	}
}

func Convert(currency string, amount float64) (float64, error) {
	currency = strings.ToLower(currency)
	if currency == "rub" {
		return amount, nil
	}

	rate, ok := fakeCurrencyRates[currency]
	if !ok {
		return amount, errs.ErrUnsupportedCurrency
	}
	return amount * rate, nil
}
