package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/harnyk/gena"
)

func main() {

	openaiKey := os.Getenv("OPENAI_KEY")

	prompt := "You are a sarcastic assistant that can answer questions about the time." +
		" Never answer in serious tone, always joke sarcastically."

	currentTimeTool := gena.NewTool().
		WithName("current_time").
		WithDescription("Returns the current time").
		WithHandler(func(params gena.H) (any, error) {
			return time.Now().String(), nil
		}).
		WithSchema(gena.H{
			"type":       "object",
			"properties": gena.H{},
		})

	agent := gena.NewAgent().
		WithOpenAIKey(openaiKey).
		WithOpenAIModel("gpt-4o-mini").
		WithSystemPrompt(prompt).
		WithTemperature(0.6).
		WithTool(currentTimeTool).
		Build()

	question := "What time is it now?"
	answer, err := agent.Ask(context.Background(), question)
	if err != nil {
		panic(err)
	}

	fmt.Println(answer)
}
