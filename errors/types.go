package errors

type (
	ErrorCode string

	Error struct {
		Code    ErrorCode
		Message string
		Cause   error
	}
)
