package main

import (
	"context"
	"fmt"
	"os"

	"github.com/harnyk/gena"
)

type H = gena.H

func main() {

	openaiKey := os.Getenv("OPENAI_KEY")

	prompt := "You are a sarcastic assistant." +
		" Never answer in serious tone, always joke sarcastically." +
		" Think step-by-step. Never call multiple tools at once." +
		" If you need to call multiple tools, base your decision on the previous tool call results."

	currentTimeTool := gena.NewTool().
		WithName("current_time").
		WithDescription("Returns the current time").
		WithHandler(NewCurrentTimeHandler()).
		WithSchema(H{
			"type":       "object",
			"properties": H{},
		})

	squareRootTool := gena.NewTool().
		WithName("square_root").
		WithDescription("Returns the square root of a number").
		WithHandler(NewSquareRootHandler()).
		WithSchema(H{
			"type":       "object",
			"properties": H{"x": H{"type": "number"}},
		})

	agent := gena.NewAgent().
		WithOpenAIKey(openaiKey).
		WithOpenAIModel("gpt-4o-mini").
		// WithAPIURL("https://api.mistral.ai/v1").
		// WithOpenAIModel("mistral-large-latest").
		WithSystemPrompt(prompt).
		WithTemperature(0.5).
		WithTool(currentTimeTool).
		WithTool(squareRootTool).
		Build()

	question := "What is the square root of the current year?"
	answer, err := agent.Ask(context.Background(), question)
	if err != nil {
		panic(err)
	}

	fmt.Println(answer)
}
