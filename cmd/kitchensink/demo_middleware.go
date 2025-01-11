package main

import (
	"fmt"

	"github.com/harnyk/gena"
)

type DemoMiddleware struct {
}

var _ gena.ToolMiddleware = (*DemoMiddleware)(nil)

func NewDemoMiddleware() *DemoMiddleware {
	return &DemoMiddleware{}
}

func (m *DemoMiddleware) Execute(params gena.H, tool *gena.Tool) (gena.ToolMiddlewareResult, error) {
	fmt.Printf("🎉 Hello from middleware! Tool: %s; Params: %v\n", tool.Name, params)

	return gena.ToolMiddlewareResult{
		Params: params,
		Stop:   false,
		Result: nil,
	}, nil
}
