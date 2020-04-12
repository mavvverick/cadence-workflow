package model

//SuccessResponse ...
type SuccessResponse struct {
	Message   string `json:"message,omitempty"`
	StatusURL string `json:"status_url,omitempty"`
}

//ErrorResponse ...
type ErrorResponse struct {
	Success bool      `json:"success" example:"false"`
	Error   HTTPError `json:"error,omitempty"`
}

//HTTPError ...
type HTTPError struct {
	Error   uint32 `json:"error,omitempty"`
	Message string `json:"message" example:"status bad request"`
}
