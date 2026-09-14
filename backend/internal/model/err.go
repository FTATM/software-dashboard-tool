package model

import "errors"

// Custome Error
var ErrDuplicate = errors.New("t_dup")
var ErrSecurityViolation = errors.New("Security Violation")
var ErrInvalidBody = errors.New("t_invalid_body")
var ErrNotActive = errors.New("t_not_active")
var ErrInUsed = errors.New("t_delete_in_used")
