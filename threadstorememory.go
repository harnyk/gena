package gena

import "github.com/sashabaranov/go-openai"

type ThreadStoreMemory struct {
	snapshot []openai.ChatCompletionMessage
}

var _ ThreadStore = (*ThreadStoreMemory)(nil)

func NewThreadStoreMemory() *ThreadStoreMemory {
	return &ThreadStoreMemory{
		snapshot: []openai.ChatCompletionMessage{},
	}
}

func (t *ThreadStoreMemory) GetSnapshot() []openai.ChatCompletionMessage {
	return t.snapshot
}

func (t *ThreadStoreMemory) AddMessage(message openai.ChatCompletionMessage) {
	t.snapshot = append(t.snapshot, message)
}
