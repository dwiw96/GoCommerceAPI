package responses

import "errors"

var (
	CodeSuccess       = 200 // 200, OK
	CodeSuccessCreate = 201 // 201, Created
	CodeSuccessUpdate = 201 // 201

	CodeFailedServer       = 500 // 500, Internal Server Error
	CodeFailedUser         = 400 // 400, Bad Request
	CodeFailedValidation   = 422 // 422, Unprocessably Entity
	CodeFailedUnauthorized = 401 // 401, Unauthorized
	CodeFailedDuplicated   = 409 // 409, Conflict
)

var (
	ErrNoData               = errors.New("no data found")              // no data found
	ErrDuplicate            = errors.New("duplicate data")             // duplicate data
	ErrViolation            = errors.New("invalid input")              // invalid input
	ErrCheckConstraint      = errors.New("invalid input")              // invalid input
	ErrNotNull              = errors.New("input is empty")             // input is empty
	ErrInvalidInput         = errors.New("invalid input")              // invalid input
	ErrServer               = errors.New("server error")               // server error
	ErrLessOrEqualToZero    = errors.New("amount must be more than 0") // amount must be more than 0
	ErrGreaterOrEqualToZero = errors.New("amount must be less than 0") // amount must be less than 0
)
