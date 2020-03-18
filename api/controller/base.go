package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/YOVO-LABS/workflow/api/model"
	"github.com/YOVO-LABS/workflow/config"
	"github.com/YOVO-LABS/workflow/internal/errors"

	"github.com/YOVO-LABS/workflow/api/constant/codes"

	"github.com/YOVO-LABS/workflow/api/constant"
)

//BaseController ...
type BaseController struct {
	// Logging logging.Logger
	Config config.AppConfig
}

// WriteWithStatus ...
func (c *BaseController) WriteWithStatus(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}

// WriteJSON ...
func (c *BaseController) WriteJSON(r *http.Request, w http.ResponseWriter, statusCode int, v interface{}) error {
	w.WriteHeader(statusCode)
	resp := model.SuccessResponse{
		Success: true,
		Data:    v,
	}
	return c.writeJSON(w, resp)
}

// WriteError ...
func (c *BaseController) WriteError(r *http.Request, w http.ResponseWriter, err error) error {
	code := getErrorCode(err)
	statusCode := getHTTPStatus(code)
	w.WriteHeader(statusCode)
	return c.writeJSON(w, &model.ErrorResponse{
		Success: false,
		Error: model.HTTPError{
			Code:    code,
			Message: err.Error(),
		},
	})
}

//WriteErrorWithMessage ...
func (c *BaseController) WriteErrorWithMessage(r *http.Request, w http.ResponseWriter, err error, message string) error {
	code := getErrorCode(err)
	statusCode := getHTTPStatus(code)

	w.WriteHeader(statusCode)
	return c.writeJSON(w, &model.ErrorResponse{
		Success: false,
		Error: model.HTTPError{
			Code:    code,
			Message: message,
		},
	})
}

func (c *BaseController) writeJSON(w http.ResponseWriter, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func getErrorCode(err error) uint32 {
	clError, ok := err.(errors.CustomError)
	if !ok {
		return codes.InternalServerError
	}
	return clError.GetCode()
}

func getHTTPStatus(code uint32) (status int) {
	firstThreeDigits := code / 100
	switch firstThreeDigits {
	case 400:
		status = http.StatusBadRequest
	case 401:
		status = http.StatusUnauthorized
	case 403:
		status = http.StatusForbidden
	case 404:
		status = http.StatusNotFound
	case 405:
		status = http.StatusMethodNotAllowed
	case 406:
		status = http.StatusNotAcceptable
	case 408:
		status = http.StatusRequestTimeout
	default:
		status = http.StatusInternalServerError
	}
	return
}

func (c *BaseController) decodeAndValidate(r *http.Request, v model.RequestValidator) error {
	err := c.decodeRequestBody(r, v)
	if err != nil {
		return err
	}
	return v.Validate(r.Context())
}

func (c *BaseController) decodeRequestBody(r *http.Request, v interface{}) (err error) {

	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		err := errors.New(codes.InternalServerError, constant.ServerIssue)
		return err
	}

	payloadBytes, err := ioutil.ReadAll(r.Body)
	fmt.Println(payloadBytes)
	defer r.Body.Close()

	json.Unmarshal(payloadBytes, &v)
	if err != nil {
		err = errors.Wrap(err, codes.FailedToDecodeRequestBody, constant.DecodeRequestBodyErr)
		return err
	}
	fmt.Println(&v)

	return nil
}
