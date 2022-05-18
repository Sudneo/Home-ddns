package api

import (
	"fmt"
)

type ErrAPIFailed struct {
	Message string
	Code    string
}

func (e *ErrAPIFailed) Error() string {
	return fmt.Sprintf("API call failed with code %s: %s", e.Code, e.Message)
}
