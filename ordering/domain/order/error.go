package order

import "errors"

var (
	ErrOrderIsSubmitted       = errors.New("order is submitted")
	ErrItemAmountLessThanZero = errors.New("item amount is less than zero")
	ErrItemNotFound           = errors.New("item not found")
)
