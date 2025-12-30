package errors

// Standard API error codes (match docs/API.md).
const (
	CodeBadRequest      = 40001
	CodeAPIKeyInvalid   = 40101
	CodeSignatureFailed = 40102
	CodeForbidden       = 40301
	CodeNotFound        = 40401
	CodeRateLimited     = 42901
	CodeInternal        = 50001
)
