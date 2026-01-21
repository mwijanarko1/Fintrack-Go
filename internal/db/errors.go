package db

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrCategoryNotFound  = errors.New("category not found")
	ErrDuplicateCategory = errors.New("category name already exists for this user")
	ErrTransactionNotFound = errors.New("transaction not found")
)
