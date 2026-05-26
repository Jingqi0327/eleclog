package api

import (
	"github.com/Jingqi0327/eleclog/util"
	"github.com/go-playground/validator/v10"
)

var validRole validator.Func = func(fl validator.FieldLevel) bool {
	if role, ok := fl.Field().Interface().(string); ok {
		return util.IsSupportedRole(role)
	}
	return false
}
