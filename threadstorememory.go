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

func (t *ThreadStoreMemory) GetSnapshot() ([]openai.ChatCompletionMessage, error) {
	return t.snapshot, nil
}

func (t *ThreadStoreMemory) AddMessage(message openai.ChatCompletionMessage) error {
	t.snapshot = append(t.snapshot, message)
	return nil
}
