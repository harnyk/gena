package gena

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/sashabaranov/go-openai"
)

type Agent struct {
	OpenAIKey    string
	OpenAIModel  string
	SystemPrompt string
	MaxTokens    int
	Temperature  float32
	Tools        []Tool
	client       *openai.Client
	ChatHistory  []openai.ChatCompletionMessage
	log          *slog.Logger
}

func NewAgent() *Agent {
	return &Agent{
		ChatHistory: []openai.ChatCompletionMessage{},
	}
}

func (a *Agent) WithLogger(logger *slog.Logger) *Agent {
	a.log = logger
	return a
}

func (a *Agent) Build() *Agent {
	a.client = openai.NewClient(a.OpenAIKey)
	if a.log == nil {
		a.log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}
	return a
}

func (a *Agent) WithOpenAIKey(key string) *Agent {
	a.OpenAIKey = key
	return a
}

func (a *Agent) WithOpenAIModel(model string) *Agent {
	a.OpenAIModel = model
	return a
}

func (a *Agent) WithSystemPrompt(prompt string) *Agent {
	a.SystemPrompt = prompt
	return a
}

func (a *Agent) WithMaxTokens(tokens int) *Agent {
	a.MaxTokens = tokens
	return a
}

func (a *Agent) WithTemperature(temperature float32) *Agent {
	a.Temperature = temperature
	return a
}

func (a *Agent) WithTool(tool *Tool) *Agent {
	a.Tools = append(a.Tools, *tool)
	return a
}

func (a *Agent) Ask(ctx context.Context, question string) (string, error) {
	a.ChatHistory = append(a.ChatHistory, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	})

	for i := 0; i < 64; i++ {
		messages := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleSystem,
			Content: a.SystemPrompt,
		}}
		messages = append(messages, a.ChatHistory...)

		resp, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       a.OpenAIModel,
			MaxTokens:   a.MaxTokens,
			Temperature: a.Temperature,
			Messages:    messages,
			Functions:   a.getOpenAITools(),
		})
		if err != nil {
			return "", err
		}
		if len(resp.Choices) == 0 {
			return "", errors.New("no choices returned from OpenAI")
		}

		choice := resp.Choices[0]
		finishReason := choice.FinishReason

		a.ChatHistory = append(a.ChatHistory, openai.ChatCompletionMessage{
			Role:         openai.ChatMessageRoleAssistant,
			Content:      choice.Message.Content,
			FunctionCall: choice.Message.FunctionCall,
		})

		switch finishReason {
		case "function_call":
			a.log.Debug("tool call",
				"tool", choice.Message.FunctionCall.Name,
				"args", choice.Message.FunctionCall.Arguments,
			)
			callResult, err := a.handleFunctionCall(choice.Message)
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
				a.ChatHistory = append(a.ChatHistory, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: text,
				})
				return text, nil
			} else {
				return "", nil
			}
		}
	}

	return "", errors.New("too many iterations without a final answer")
}

func (a *Agent) getOpenAITools() []openai.FunctionDefinition {
	var functions []openai.FunctionDefinition
	for _, tool := range a.Tools {
		functions = append(functions, openai.FunctionDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Schema,
		})
	}
	return functions
}

func (a *Agent) handleFunctionCall(message openai.ChatCompletionMessage) (string, error) {
	if message.FunctionCall == nil {
		return "", errors.New("no function call in message")
	}

	for _, tool := range a.Tools {
		if tool.Name == message.FunctionCall.Name {

			argsJSON := message.FunctionCall.Arguments

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

			a.ChatHistory = append(a.ChatHistory, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: response,
			})

			return response, nil
		}
	}

	return "", fmt.Errorf("tool not found: %s", message.FunctionCall.Name)
}
