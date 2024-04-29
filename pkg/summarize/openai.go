package summarize

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/MrWong99/summairpg/pkg/transcribe"
	"github.com/sashabaranov/go-openai"
)

// OpenAIClient to use when addressing the OpenAI API.
type OpenAIClient struct {
	// Client to use.
	Client *openai.Client
	// Model of AI to use.
	Model string
}

// NewOpenAIClient creates a new OpenAIClient using the http.DefaultClient
func NewOpenAIClient(baseUrl, model, orgId string, apiType openai.APIType, apiVersion string) *OpenAIClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseUrl
	config.OrgID = orgId
	config.APIType = apiType
	config.APIVersion = apiVersion
	config.HTTPClient = http.DefaultClient
	return &OpenAIClient{
		Client: openai.NewClientWithConfig(config),
		Model:  model,
	}
}

// Summarize the given lines of text using a special system prompt.
func (c *OpenAIClient) Summarize(lines []transcribe.Line) (string, error) {
	// TODO: split by tokens so not to overload the AI
	fullText := ""
	for _, line := range lines {
		fullText += "\n" + line.String()
	}
	fullText = strings.TrimPrefix(fullText, "\n")
	ctx := context.Background()
	resp, err := c.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: summarySystemPrompt,
			},
			{
				Role:    "user",
				Content: fullText,
			},
		},
	})
	if err != nil {
		return "", err
	}
	allResponses := ""
	for i, choice := range resp.Choices {
		if i == 0 {
			allResponses = choice.Message.Content
		} else {
			allResponses = fmt.Sprintf("%s\n\n%s", allResponses, choice.Message.Content)
		}
	}
	if allResponses == "" {
		return "", errors.New("no summary returned by ChatGPT")
	}
	return allResponses, nil
}
