package client

import (
	"fmt"
	"strings"
	"time"
)

type ResponseError struct {
	Errors     []Error       `json:"errors"`
	StatusCode int           `json:"-"`
	RetryAfter time.Duration `json:"-"`
}

func (r ResponseError) Error() string {
	var builder strings.Builder
	for _, e := range r.Errors {
		const prefix = "\n\t- "
		builder.WriteString(prefix)
		builder.WriteString(fmt.Sprintf("%v: %v", e.Code, e.Message))
	}
	return builder.String()
}

type Error struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}
