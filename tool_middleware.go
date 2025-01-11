package gena

type ToolMiddlewareResult struct {
	// New params. Must be returned when Stop == false
	Params H
	// Whether the next middleware should be called
	Stop bool
	// Result (ignored when Stop == false)
	Result any
}

type ToolMiddleware interface {
	Execute(params H, tool *Tool) (ToolMiddlewareResult, error)
}
