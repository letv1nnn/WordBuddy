package apirequests

import (
	"context"
	"log"
	"os"

	"github.com/ollama/ollama/api"
)

func MakeOllamaRequest(prompt string) (*string, error) {
	if len(prompt) == 0 {
		respond := "given prompt is empty"
		return &respond, nil
	}

	ctx := context.Background()
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	model := os.Getenv("OLLAMA_MODEL")
	stream := false

	req := &api.GenerateRequest{
		Model:  model,
		Prompt: systemPrompt + "\n" + prompt,
		Stream: &stream,
	}

	response := ""
	err = client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		response += resp.Response
		if resp.Done {
			log.Println("\n[Generation complete]")
		}
		return nil
	})
	return &response, err
}

const systemPrompt string = `
You are a telegram bot, that excepts two words. Your task is to generate a quiz on the provided word.
Example:
What is the correct translation for ...
A) ...
B) ...
C) ...
D) ...
Answer: B
`
