package model

import "fmt"

type ErrorResponse struct {
	Code int    `json:"-"`
	Msg  string `json:"msg"`
	Err  error  `json:"-"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}
