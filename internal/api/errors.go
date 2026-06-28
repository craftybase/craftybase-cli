package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/craftybase/stocksmith-cli/internal/brand"
)

type APIError struct {
	StatusCode int
	Message    string
	RequestID  string
}

func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("%s (request ID: %s)", e.Message, e.RequestID)
	}
	return e.Message
}

var (
	ErrUnauthorized = &APIError{StatusCode: 401, Message: "Not authenticated — run '" + brand.BinaryName + " auth login'."}
	ErrForbidden    = &APIError{StatusCode: 403, Message: "API access is not enabled for this account or plan."}
	ErrNotFound     = &APIError{StatusCode: 404, Message: "Resource not found."}
	ErrRateLimited  = &APIError{StatusCode: 429, Message: "Rate limit exceeded. Try again later."}
	ErrServerError  = &APIError{StatusCode: 500, Message: brand.ProductName + " API server error."}
)

func ExitCode(err error) int {
	if apiErr, ok := err.(*APIError); ok {
		switch apiErr.StatusCode {
		case 401:
			return 3
		case 403:
			return 1
		case 404:
			return 4
		case 429:
			return 1
		default:
			if apiErr.StatusCode >= 500 {
				return 1
			}
		}
	}
	return 1
}

func errorBody(body []byte) string {
	var envelope struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error != "" {
		return envelope.Error
	}
	return ""
}

func MapHTTPError(resp *http.Response) error {
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	requestID := resp.Header.Get("X-Request-Id")

	switch resp.StatusCode {
	case 401:
		return &APIError{
			StatusCode: 401,
			Message:    "Not authenticated — run '" + brand.BinaryName + " auth login'.",
			RequestID:  requestID,
		}
	case 403:
		msg := errorBody(body)
		if msg == "" {
			msg = "API access is not enabled for this account or plan."
		}
		return &APIError{StatusCode: 403, Message: msg, RequestID: requestID}
	case 404:
		msg := errorBody(body)
		if msg == "" {
			msg = "Resource not found."
		}
		return &APIError{StatusCode: 404, Message: msg, RequestID: requestID}
	case 429:
		return &APIError{StatusCode: 429, Message: "Rate limit exceeded.", RequestID: requestID}
	default:
		if resp.StatusCode >= 500 {
			msg := errorBody(body)
			if msg == "" {
				msg = fmt.Sprintf("%s API error (%d).", brand.ProductName, resp.StatusCode)
			}
			if requestID != "" {
				msg = fmt.Sprintf("%s Request ID: %s", msg, requestID)
			}
			return &APIError{StatusCode: resp.StatusCode, Message: msg, RequestID: requestID}
		}
	}
	return nil
}
