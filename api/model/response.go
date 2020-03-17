package model

//SuccessResponse ...
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data,omitempty" `
}

//ErrorResponse ...
type ErrorResponse struct {
	Success bool      `json:"success" example:"false"`
	Error   HTTPError `json:"error,omitempty"`
}

//HTTPError ...
type HTTPError struct {
	Code    uint32 `json:"code" example:"40001"`
	Message string `json:"message" example:"status bad request"`
}
