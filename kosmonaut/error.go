package kosmonaut

import (
	"fmt"
	"strconv"
)

// Error represents a WebRocket protocol error.
type Error struct {
	Status string
	Code   int
}

// Error returns stringified status code with message.
func (err *Error) Error() string {
	return fmt.Sprintf("%d %s", err.Code, err.Status)
}

// Status codes.
const (
	EBadRequest         = 400
	EUnauthorized       = 402
	EForbidden          = 403
	EInvalidChannelName = 451
	EChannelNotFound    = 454
	EInternalError      = 597
	EOF                 = 598
	EUnknown            = 0
)

// Possible status messages.
var statusMessages = map[int]string{
	EBadRequest:         "Bad request",
	EUnauthorized:       "Unauthorized",
	EForbidden:          "Forbidden",
	EInvalidChannelName: "Invalid channel name",
	EChannelNotFound:    "Channel not found",
	EInternalError:      "Internal error",
	EOF:                 "End of file",
}

// parseError takes received frames and extracts error information form it.
//
// frames - The data to be parsed.
//
// Returns an error instance.
func parseError(frames []string) *Error {
	if len(frames) == 1 {
		code, _ := strconv.Atoi(frames[0])
		if status, ok := statusMessages[code]; ok {
			return &Error{status, code}
		}
	}
	return &Error{"Unknown server error", EUnknown}
}
