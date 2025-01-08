package gena

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
)

type Agent struct {
	openAIKey               string
	openAIModel             string
	apiURL                  string
	systemPrompt            string
	maxTokens               int
	temperature             float32
	tools                   []Tool
	client                  *openai.Client
	log                     *slog.Logger
	threadStore             ThreadStore
	maxAutonomousIterations int
}

func NewAgent() *Agent {
	return &Agent{}
}

// WithThreadStore
//
// The thread store is used to store the conversation history.
// Customize it to your needs, for example, a file store or a database.
// If not provided, a default in-memory store will be used.
func (a *Agent) WithThreadStore(store ThreadStore) *Agent {
	a.threadStore = store
	return a
}

func (a *Agent) WithMaxAutonomousIterations(iterations int) *Agent {
	a.maxAutonomousIterations = iterations
	return a
}

func (a *Agent) WithLogger(logger *slog.Logger) *Agent {
	a.log = logger
	return a
}

func (a *Agent) WithAPIURL(url string) *Agent {
	a.apiURL = url
	return a
}

func (a *Agent) Build() *Agent {
	config := openai.DefaultConfig(a.openAIKey)
	if a.apiURL != "" {
		config.BaseURL = a.apiURL
	}
	a.client = openai.NewClientWithConfig(config)

	if a.log == nil {
		a.log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	if a.threadStore == nil {
		a.threadStore = NewThreadStoreMemory()
	}

	if a.maxAutonomousIterations == 0 {
		a.maxAutonomousIterations = 64
	}

	return a
}

func (a *Agent) WithOpenAIKey(key string) *Agent {
	a.openAIKey = key
	return a
}

func (a *Agent) WithOpenAIModel(model string) *Agent {
	a.openAIModel = model
	return a
}

func (a *Agent) WithSystemPrompt(prompt string) *Agent {
	a.systemPrompt = prompt
	return a
}

func (a *Agent) WithMaxTokens(tokens int) *Agent {
	a.maxTokens = tokens
	return a
}

func (a *Agent) WithTemperature(temperature float32) *Agent {
	a.temperature = temperature
	return a
}

func (a *Agent) WithTool(tool *Tool) *Agent {
	a.tools = append(a.tools, *tool)
	return a
}

func (a *Agent) createChatCompletionWithRetryOnRateLimitError(ctx context.Context, request openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	var resp openai.ChatCompletionResponse
	var err error
	for i := 0; i < 5; i++ {
		resp, err = a.client.CreateChatCompletion(ctx, request)
		if err == nil {
			return &resp, nil
		}
		if a.isRateLimitError(err) {
			a.log.Warn("rate limit error, retrying after 5 seconds")
			time.Sleep(5 * time.Second)
			continue
		}
		return nil, err
	}
	return nil, err
}

func (a *Agent) isRateLimitError(err error) bool {
	openaiErr, ok := err.(*openai.RequestError)
	if !ok {
		return false
	}
	return openaiErr.HTTPStatusCode == http.StatusTooManyRequests
}

func (a *Agent) Ask(ctx context.Context, question string) (string, error) {
	if err := a.threadStore.AddMessage(openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	}); err != nil {
		return "", err
	}

	for i := 0; i < a.maxAutonomousIterations; i++ {
		threadWithSystemPrompt := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleSystem,
			Content: a.systemPrompt,
		}}

		thread, err := a.threadStore.GetSnapshot()
		if err != nil {
			return "", err
		}

		threadWithSystemPrompt = append(threadWithSystemPrompt, thread...)

		resp, err := a.createChatCompletionWithRetryOnRateLimitError(ctx, openai.ChatCompletionRequest{
			Model:       a.openAIModel,
			MaxTokens:   a.maxTokens,
			Temperature: a.temperature,
			Messages:    threadWithSystemPrompt,
			// Functions:   a.getOpenAITools(),
			Tools: a.getOpenAITools(),
		})
		if err != nil {
			return "", err
		}
		if len(resp.Choices) == 0 {
			return "", errors.New("no choices returned from OpenAI")
		}

		choice := resp.Choices[0]
		finishReason := choice.FinishReason

		if err := a.threadStore.AddMessage(openai.ChatCompletionMessage{
			Role:         openai.ChatMessageRoleAssistant,
			Content:      choice.Message.Content,
			FunctionCall: choice.Message.FunctionCall,
			ToolCalls:    choice.Message.ToolCalls,
		}); err != nil {
			return "", err
		}

		switch finishReason {
		case "tool_calls":
			toolCalls := choice.Message.ToolCalls
			a.log.Debug("tool calls",
				"count", len(toolCalls),
			)
			err := a.handleToolCalls(toolCalls)
			if err != nil {
				return "", err
			}
			continue
		case "function_call":
			a.log.Debug("function call",
				"tool", choice.Message.FunctionCall.Name,
				"args", choice.Message.FunctionCall.Arguments,
			)
			callResult, err := a.handleFunctionCall(choice.Message.FunctionCall)
			if err != nil {
				return "", err
			}
			a.log.Debug("tool call result",
				"tool", choice.Message.FunctionCall.Name,
				"result", callResult,
			)
			continue

		case "stop":
			return choice.Message.Content, nil

		default:
			text := choice.Message.Content
			if text != "" {
				if err := a.threadStore.AddMessage(openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: text,
				}); err != nil {
					return "", err
				}
				return text, nil
			} else {
				return "", nil
			}
		}
	}

	return "", errors.New("too many iterations without a final answer")
}

func (a *Agent) getOpenAITools() []openai.Tool {
	var functions []openai.Tool
	for _, tool := range a.tools {
		functions = append(functions, openai.Tool{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Schema,
			},
		})
	}
	return functions
}

func (a *Agent) toolByName(name string) *Tool {
	for _, tool := range a.tools {
		if tool.Name == name {
			return &tool
		}
	}
	return nil
}

func (a *Agent) handleToolCalls(toolCalls []openai.ToolCall) error {
	for _, toolCall := range toolCalls {
		a.log.Debug("tool call",
			"tool", toolCall.Function.Name,
			"args", toolCall.Function.Arguments,
		)

		tool := a.toolByName(toolCall.Function.Name)
		if tool == nil {
			return errors.New("tool not found")
		}

		argumentsStr := toolCall.Function.Arguments

		var argumentsMap H
		if err := json.Unmarshal([]byte(argumentsStr), &argumentsMap); err != nil {
			return fmt.Errorf("failed to unmarshal tool call arguments: %w", err)
		}

		result, err := tool.Run(argumentsMap)
		if err != nil {
			result = fmt.Sprintf("error: %s", err.Error())
		}

		marshalledResult, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal tool call result: %w", err)
		}

		a.log.Debug("tool call result",
			"tool", tool.Name,
			"result", string(marshalledResult),
		)

		toolReplyMessage := openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Name:       tool.Name,
			ToolCallID: toolCall.ID,
			Content:    string(marshalledResult),
		}

		if err := a.threadStore.AddMessage(toolReplyMessage); err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) handleFunctionCall(functionCall *openai.FunctionCall) (string, error) {
	if functionCall == nil {
		return "", errors.New("no function call in message")
	}

	for _, tool := range a.tools {
		if tool.Name == functionCall.Name {

			argsJSON := functionCall.Arguments

			var argsMap map[string]interface{}
			if err := json.Unmarshal([]byte(argsJSON), &argsMap); err != nil {
				return "", fmt.Errorf("failed to unmarshal function call arguments: %w", err)
			}

			result, err := tool.Run(argsMap)
			if err != nil {
				result = fmt.Sprintf("error: %s", err.Error())
			}

			marshalledResult, err := json.Marshal(result)
			if err != nil {
				return "", fmt.Errorf("failed to marshal function call result: %w", err)
			}
			response := fmt.Sprintf("Function '%s' result: %s", tool.Name, marshalledResult)

			if err := a.threadStore.AddMessage(openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: response,
			}); err != nil {
				return "", err
			}

			return response, nil
		}
	}

	return "", fmt.Errorf("tool not found: %s", functionCall.Name)
}
