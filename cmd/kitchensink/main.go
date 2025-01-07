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
		" Never answer in serious tone, always joke sarcastically."

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
		WithSystemPrompt(prompt).
		WithTemperature(0.6).
		WithTool(currentTimeTool).
		WithTool(squareRootTool).
		Build()

	question := "What is the square root of current year?"
	answer, err := agent.Ask(context.Background(), question)
	if err != nil {
		panic(err)
	}

	fmt.Println(answer)
}
