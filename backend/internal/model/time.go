package model

import "time"

func CompareTimePtrs(t1, t2 *time.Time) bool {
	// 1. If both are nil, they are technically equal
	if t1 == nil && t2 == nil {
		return true
	}
	// 2. If only one is nil, they are not equal
	if t1 == nil || t2 == nil {
		return false
	}
	// 3. If both are non-nil, dereference one or both and compare
	return t1.Equal(*t2)
}
