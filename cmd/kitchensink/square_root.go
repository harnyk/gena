package main

import (
	"math"

	"github.com/harnyk/gena"
)

type SquareRootHandlerParams struct {
	X float64 `mapstructure:"x"`
}

type SquareRootHandler struct {
}

func NewSquareRootHandler() *SquareRootHandler {
	return &SquareRootHandler{}
}

func (h *SquareRootHandler) Execute(params gena.H) (any, error) {
	return gena.ExecuteTyped(h.execute, params)
}

func (h *SquareRootHandler) execute(params SquareRootHandlerParams) (float64, error) {
	return math.Sqrt(params.X), nil
}
