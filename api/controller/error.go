package controller

import (
	"net/http"

	"github.com/YOVO-LABS/workflow/api/constant"
	"github.com/YOVO-LABS/workflow/api/constant/codes"
	"github.com/YOVO-LABS/workflow/internal/errors"
)

//HTTPErrorController ...
type HTTPErrorController struct {
	BaseController
}

//ResourceNotFound ...
func (c *HTTPErrorController) ResourceNotFound(w http.ResponseWriter, r *http.Request) {
	err := errors.New(codes.NotFound, constant.ResourceNotFound)
	c.RespondWithJSON(w, http.StatusBadRequest, err)
}
