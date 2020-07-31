package encoder

import "errors"

var (
	ErrBounds       = errors.New("index out of bounds")
	ErrLength       = errors.New("code length does not match encoder length")
	ErrTargetLength = errors.New("target data is not same length as categorical data")
)
