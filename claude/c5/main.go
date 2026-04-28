package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseUrl := os.Getenv("OPENROUTER_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		panic("Env variable OPENROUTER_API_KEY not found")
	}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(prompt),
	}

	tools := []openai.ChatCompletionToolUnionParam{
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "Read",
			Description: openai.String("Read and return the contents of a file"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"file_path": map[string]any{
						"type":        "string",
						"description": "The path to the file to read",
					},
				},
				"required": []string{"file_path"},
			},
		}),

		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "Write",
			Description: openai.String("Write content to a file"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"file_path": map[string]any{
						"type":        "string",
						"description": "The path to the file to write to",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "The content to write to the file",
					},
				},
				"required": []string{"file_path", "content"},
			},
		}),
	}

	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl))
	for {
		resp, err := client.Chat.Completions.New(context.Background(),
			openai.ChatCompletionNewParams{
				Model:    "anthropic/claude-haiku-4.5",
				Messages: messages,
				Tools:    tools,
			},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if len(resp.Choices) == 0 {
			panic("No choices in response")
		}

		messages = append(messages, resp.Choices[0].Message.ToParam())

		if len(resp.Choices[0].Message.ToolCalls) == 0 {
			fmt.Print(resp.Choices[0].Message.Content)
			break
		}

		for _, toolCall := range resp.Choices[0].Message.ToolCalls {
			switch toolCall.Function.Name {
			case "Read":
				var args struct {
					FilePath string `json:"file_path"`
				}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
					messages = append(messages, openai.ToolMessage(fmt.Sprintf("error parsing arguments: %v", err), toolCall.ID))
					continue
				}

				content, err := os.ReadFile(args.FilePath)
				if err != nil {
					messages = append(messages, openai.ToolMessage(fmt.Sprintf("error reading file: %v", err), toolCall.ID))
					continue
				}
				messages = append(messages, openai.ToolMessage(string(content), toolCall.ID))
			case "Write":
				var args struct {
					FilePath string `json:"file_path"`
					Content  string `json:"content"`
				}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
					messages = append(messages, openai.ToolMessage(fmt.Sprintf("error parsing arguments: %v", err), toolCall.ID))
					continue
				}

				file, err := os.Create(args.FilePath)
				if err != nil {
					messages = append(messages, openai.ToolMessage(fmt.Sprintf("error creating or truncating file: %v", err), toolCall.ID))
					continue
				}

				if _, err := file.WriteString(args.Content); err != nil {
					file.Close()
					messages = append(messages, openai.ToolMessage(fmt.Sprintf("error writing to file: %v", err), toolCall.ID))
					continue
				}
				if err := file.Close(); err != nil {
					messages = append(messages, openai.ToolMessage(fmt.Sprintf("error closing file: %v", err), toolCall.ID))
					continue
				}

				messages = append(messages, openai.ToolMessage(args.Content, toolCall.ID))
			default:
				messages = append(messages, openai.ToolMessage(fmt.Sprintf("unknown tool: %s", toolCall.Function.Name), toolCall.ID))
			}
		}
	}
}
