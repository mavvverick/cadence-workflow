package model

import (
	"context"
	"jobprocessor/api/constant"
	"jobprocessor/api/constant/codes"
	"jobprocessor/internal/errors"

	"gopkg.in/validator.v2"
)

// RequestValidator ...
type RequestValidator interface {
	Validate(ctx context.Context) error
}

// ValidateFields checks if the required fields in a model is filled.
func ValidateFields(model interface{}) error {
	err := validator.Validate(model)
	if err != nil {
		errs, ok := err.(validator.ErrorMap)
		if ok {
			for f := range errs {
				return errors.New(codes.ValidateField, constant.ValidateFieldErr+"-"+f)
			}
		} else {
			return errors.New(codes.ValidationUnknown, constant.ValidationUnknownErr)
		}
	}

	return nil
}
