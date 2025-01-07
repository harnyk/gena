package gena

import (
	"github.com/mitchellh/mapstructure"
)

type TypedExecutor[P any, R any] func(P) (R, error)

// type UntypedExecutor func(H) (any, error)

type ToolHandler interface {
	Execute(params H) (any, error)
}

func ExecuteTyped[P any, R any](handler TypedExecutor[P, R], params H) (any, error) {
	var paramsTyped P
	if err := mapstructure.Decode(params, &paramsTyped); err != nil {
		return nil, err
	}

	return handler(paramsTyped)
}
