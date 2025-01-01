package gena

import "github.com/sashabaranov/go-openai"

type ThreadStore interface {
	GetSnapshot() ([]openai.ChatCompletionMessage, error)
	AddMessage(message openai.ChatCompletionMessage) error
}
