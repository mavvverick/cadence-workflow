package model

import (
	"context"

	"github.com/YOVO-LABS/workflow/api/constant"
	"github.com/YOVO-LABS/workflow/api/constant/codes"
	"github.com/YOVO-LABS/workflow/internal/errors"

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
