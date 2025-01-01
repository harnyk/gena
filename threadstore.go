package gena

import "github.com/sashabaranov/go-openai"

type ThreadStore interface {
	GetSnapshot() []openai.ChatCompletionMessage
	AddMessage(message openai.ChatCompletionMessage)
}
