package controller

import (
	"jobprocessor/api/constant"
	"jobprocessor/api/constant/codes"
	"jobprocessor/internal/errors"
	"net/http"
)

//HTTPErrorController ...
type HTTPErrorController struct {
	BaseController
}

//ResourceNotFound ...
func (c *HTTPErrorController) ResourceNotFound(w http.ResponseWriter, r *http.Request) {
	err := errors.New(codes.NotFound, constant.ResourceNotFound)
	c.WriteError(r, w, err)
	return
}
