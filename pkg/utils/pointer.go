package utils

func Ptr[T any](value T) *T {
	return &value
}

func PtrClone[T any](value *T) *T {
	if value == nil {
		return nil
	}

	copyValue := *value

	return &copyValue
}

func PtrEqual[T comparable](left, right *T) bool {
	if left == nil && right == nil {
		return true
	}

	if left == nil || right == nil {
		return false
	}

	return *left == *right
}
