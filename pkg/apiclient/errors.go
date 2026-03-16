package apiclient

import (
	"encoding/json"
	"errors"
	"fmt"
)

type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"error"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("fly api error (status %d): %s", e.StatusCode, e.Message)
}

func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}

func IsConflict(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 409
	}
	return false
}

func parseAPIError(statusCode int, body []byte) error {
	apiErr := &APIError{StatusCode: statusCode}
	if len(body) > 0 {
		_ = json.Unmarshal(body, apiErr)
	}
	if apiErr.Message == "" {
		apiErr.Message = fmt.Sprintf("unexpected status code %d", statusCode)
	}
	return apiErr
}
