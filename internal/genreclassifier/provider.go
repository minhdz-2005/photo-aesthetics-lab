package genreclassifier

import (
	"context"
	"errors"
	"fmt"
)

// InputError marks a problem with our own inputs: the photo, the taxonomy, a
// flag, a missing token. Exit 1.
type InputError struct{ msg string }

func (e *InputError) Error() string { return e.msg }

// ProviderError marks the outside world returning something unusable. Exit 2.
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

// Provider scores every candidate in one call and keys the result by candidate
// id. It never sets thresholds and never decides other or needs_review: that
// is the caller's job (ai/model-provider README).
type Provider interface {
	Score(ctx context.Context, photo Photo, cands []Candidate) (map[string]float64, error)
}
