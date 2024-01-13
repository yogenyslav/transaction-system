package model

import (
	"time"
)

const TransactionsTable = "transactions"

type Transaction struct {
	Id        uint
	AccountId uint
	Amount    float64
	Currency  string
	Operation Operation
	Status    Status
	CreatedAt time.Time
}

type TransactionRequest struct {
	AccountId uint    `json:"accountId"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}
