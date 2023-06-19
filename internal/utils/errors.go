package utils

import "errors"

var (
	ErrNonExist = errors.New("object is not exists")
	ErrConflict = errors.New("object is already exists")
)
