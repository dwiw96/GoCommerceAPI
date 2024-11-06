package responses

import "errors"

var (
	CodeSuccess       = 200 // 200
	CodeSuccessCreate = 201 // 201
	CodeSuccessUpdate = 201 // 201

	CodeFailedServer       = 500 // 500
	CodeFailedUser         = 400 // 400
	CodeFailedValidation   = 422 // 422
	CodeFailedUnauthorized = 401 // 401
	CodeFailedDuplicated   = 409 // 409
)

var (
	ErrNoData    = errors.New("no data found")
	ErrDuplicate = errors.New("duplicate data")
	ErrViolation = errors.New("input is no valid")
)
