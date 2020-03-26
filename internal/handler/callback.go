package handler

import (
	"bytes"
	"fmt"
	"net/http"
)

//ErrorHandler ...
type ErrorHandler struct {
	error  string
	status string
	Callback
}

//Callback ...
type Callback struct {
	url     string
	message string
}

//NewErrorHandler ...
func NewErrorHandler(url string) *ErrorHandler {
	return &ErrorHandler{
		status: "Failed",
		Callback: Callback{
			url: url,
		},
	}
}

// NewCallback ...
func NewCallback(url string) *Callback {
	return &Callback{
		url: url,
	}
}

//SendErrorMessage ...
func (e *ErrorHandler) SendErrorMessage(msg string) {
	payload := fmt.Sprintf(`{"message" : %v, "status": %v}`, msg, e.status)
	jsonStr := []byte(payload)

	req, err := http.NewRequest("POST", e.url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

}

//SendSucessMessage ...
func (b *Callback) SendSucessMessage(msg string) {

	payload := fmt.Sprintf("message : %v", msg)
	jsonStr := []byte(payload)

	req, err := http.NewRequest("POST", b.url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
}
