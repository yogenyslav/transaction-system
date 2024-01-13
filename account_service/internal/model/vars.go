package model

type Operation int8

const (
	_ Operation = iota
	Invoice
	Withdraw
)

type Status int8

const (
	_ Status = iota
	Success
	Error
	Created
)
