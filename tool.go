package gena

import "errors"

type Tool struct {
	Name        string
	Description string
	Schema      H
	Handler     ToolHandler
	Middlewares []ToolMiddleware
}

func NewTool() *Tool {
	return &Tool{}
}

func (t *Tool) WithName(name string) *Tool {
	t.Name = name
	return t
}

func (t *Tool) WithDescription(description string) *Tool {
	t.Description = description
	return t
}

func (t *Tool) WithSchema(schema H) *Tool {
	t.Schema = schema
	return t
}

func (t *Tool) WithHandler(handler ToolHandler) *Tool {
	t.Handler = handler
	return t
}

func (t *Tool) WithMiddleware(middleware ToolMiddleware) *Tool {
	t.Middlewares = append(t.Middlewares, middleware)
	return t
}

func (t *Tool) Run(params H) (any, error) {
	if t.Handler == nil {
		return nil, errors.New("no handler defined for tool")
	}

	currentParams := params

	for _, middleware := range t.Middlewares {
		result, err := middleware.Execute(currentParams, t)
		if err != nil {
			return nil, err
		}
		if result.Stop {
			return result.Result, nil
		}
		currentParams = result.Params
	}

	return t.Handler.Execute(params)
}

func (t *Tool) String() string {
	return t.Name
}
