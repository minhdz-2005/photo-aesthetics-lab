package genreclassifier

import (
	"errors"
	"fmt"
)

// InputError: dữ liệu phía ta sai (ảnh, taxonomy, cờ, token thiếu). Exit 1.
type InputError struct{ msg string }

func (e *InputError) Error() string { return e.msg }

// ProviderError: thế giới bên ngoài trả về thứ không dùng được. Exit 2.
type ProviderError struct{ msg string }

func (e *ProviderError) Error() string { return e.msg }

// InputErrorf and ProviderErrorf take %v, never %w: the concrete error is
// rendered into the message, and the wrapper carries its exit-code class.
func InputErrorf(format string, a ...any) error { return &InputError{fmt.Sprintf(format, a...)} }

func ProviderErrorf(format string, a ...any) error { return &ProviderError{fmt.Sprintf(format, a...)} }

func isInput(err error) bool {
	var target *InputError
	return errors.As(err, &target)
}

func isProvider(err error) bool {
	var target *ProviderError
	return errors.As(err, &target)
}
