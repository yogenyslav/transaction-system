package service

import (
	"accountservice/internal/errs"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
)

type currencyRate struct {
	Valute map[string]struct {
		CharCode string  `json:"CharCode"`
		Nominal  int     `json:"Nominal"`
		Value    float64 `json:"Value"`
	} `json:"Valute"`
}

func Convert(currency string, amount float64) (float64, error) {
	if currency == "RUB" {
		return amount, nil
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get("https://www.cbr-xml-daily.ru/daily_json.js")
	if err != nil {
		return amount, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return amount, err
	}

	rates := currencyRate{}
	if err := json.Unmarshal(raw, &rates); err != nil {
		return amount, err
	}

	for _, rate := range rates.Valute {
		if rate.CharCode == currency {
			return amount * (rate.Value / float64(rate.Nominal)), nil
		}
	}

	return amount, errs.ErrUnsupportedCurrency
}
