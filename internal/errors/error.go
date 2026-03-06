package errors

import "errors"

// Details is a typed error wrapper with compile-time safe metadata.
// T may be any struct, including nested structures.
type Details[T any] struct {
	Meta  T
	Cause error
}

func (d *Details[T]) Error() string {
	if d == nil || d.Cause == nil {
		return "<nil>"
	}

	return d.Cause.Error()
}

func (d *Details[T]) Unwrap() error {
	if d == nil {
		return nil
	}

	return d.Cause
}

// WithMeta replaces metadata in-place and returns the same pointer for chaining.
func (d *Details[T]) WithMeta(meta T) *Details[T] {
	if d == nil {
		return nil
	}

	d.Meta = meta

	return d
}

// EditMeta mutates metadata in-place via typed callback without any type assertions.
func (d *Details[T]) EditMeta(edit func(meta *T)) *Details[T] {
	if d == nil || edit == nil {
		return d
	}

	edit(&d.Meta)

	return d
}

// Wrap attaches typed metadata to err and preserves errors.Is/errors.As via Unwrap.
func Wrap[T any](err error, meta T) *Details[T] {
	if err == nil {
		return nil
	}

	return &Details[T]{Meta: meta, Cause: err}
}

// WrapZero wraps err using zero-value metadata of T.
func WrapZero[T any](err error) *Details[T] {
	if err == nil {
		return nil
	}

	var zero T
	return &Details[T]{Meta: zero, Cause: err}
}

func Meta[T any](err error) (*T, bool) {
	var details *Details[T]
	if !errors.As(err, &details) || details == nil {
		return nil, false
	}
	return &details.Meta, true
}
