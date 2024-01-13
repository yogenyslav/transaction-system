package model

import (
	"time"
)

const AccountsTable = "accounts"

type Account struct {
	// валюта по умолчанию - рубль RUB
	Id        uint      `json:"id"`
	Balance   float64   `json:"balance"`
	Frozen    float64   `json:"frozen"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
