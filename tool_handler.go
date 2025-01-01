package gena

import (
	"github.com/mitchellh/mapstructure"
)

type TypelessHandler func(H) (any, error)

type TypedHandler[T any, R any] func(T) (R, error)

func NewTypedHandler[T any, R any](handler TypedHandler[T, R]) TypedHandler[T, R] {
	return handler
}

func (t TypedHandler[T, R]) AcceptingMapOfAny() TypelessHandler {
	return func(params H) (any, error) {
		var typedParams T
		if err := mapstructure.Decode(params, &typedParams); err != nil {
			return nil, err
		}
		return t(typedParams)
	}
}
