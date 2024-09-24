package api

import (
	"github.com/go-playground/validator/v10"
	"learn.bleckshiba/banking/enum"
)

var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		return enum.IsSupportedCurrency(currency)
	}

	return false
}
