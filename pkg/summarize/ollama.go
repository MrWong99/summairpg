package summarize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/MrWong99/summairpg/pkg/transcribe"
	"github.com/sashabaranov/go-openai"
)

// OllamaClient to use when addressing the Ollama API.
type OllamaClient struct {
	// Address is the combination of host:port for the Ollama endpoint.
	Address string
	// Model of AI to use.
	Model string
	// ContextLength to set (num_ctx).
	ContextLength int
	// HttpClient to use when making requests.
	HttpClient *http.Client
}

// NewOllamaClient creates a new OllamaClient using the http.DefaultClient and a context length determined by the model.
func NewOllamaClient(address, model string) *OllamaClient {
	oc := OllamaClient{
		Address:       address,
		Model:         model,
		ContextLength: contextLengthByModel(model),
		HttpClient:    http.DefaultClient,
	}
	return &oc
}

func contextLengthByModel(model string) int {
	modelWithoutTag := strings.Split(model, ":")[0]
	switch modelWithoutTag {
	case "llama2":
		return 4096
	case "llama3", "mistral":
		return 8192
	case "llama3-gradient":
		return 32768 // just one value, would be better to calculate based on input
	case "phi3":
		return 128000
	case "mixtral":
		return 64000
	default:
		return 2048
	}
}

// OllamaChatRequest HTTP body to send for the Summarize method.
type OllamaChatRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
		//Images  []string `json:"images"`
	} `json:"messages"`
	Stream  bool           `json:"stream"`
	Options map[string]any `json:"options"`
}

func (cr *OllamaChatRequest) toOpenAIMessages() []openai.ChatCompletionMessage {
	res := make([]openai.ChatCompletionMessage, len(cr.Messages))
	for i, msg := range cr.Messages {
		res[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return res
}

// OllamaChatResponse HTTP body returned by Ollama
type OllamaChatResponse struct {
	Model   string `json:"model"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

// Summarize the given lines of text using a special system prompt.
func (c *OllamaClient) Summarize(lines []transcribe.Line) (string, error) {
	// TODO: split by tokens so not to overload the AI
	fullText := ""
	for _, line := range lines {
		fullText += "\n" + line.String()
	}
	fullText = strings.TrimPrefix(fullText, "\n")
	chatReq := OllamaChatRequest{
		Model: c.Model,
		Messages: []struct {
			Role    string "json:\"role\""
			Content string "json:\"content\""
		}{
			{
				Role:    "system",
				Content: summarySystemPrompt,
			},
			{
				Role:    "user",
				Content: fullText,
			},
		},
		Stream: false,
		Options: map[string]any{
			"num_ctx": c.ContextLength,
		},
	}
	tokenCount := NumTokensFromMessages(chatReq.toOpenAIMessages())
	if tokenCount > c.ContextLength-500 { // leave some tokens spare so AI can answer
		slog.Warn("input token count is very close or bigger than context length", "tokens", tokenCount, "context-length", c.ContextLength)
		if strings.Split(c.Model, ":")[0] == "llama3-gradient" {
			slog.Info("since you are using the llama3-gradient model the context length will be automatically adjusted for the bigger token count")
			chatReq.Options["num_ctx"] = tokenCount + 500
		}
	}
	res, err := json.Marshal(&chatReq)
	if err != nil {
		return "", fmt.Errorf("could not encode request JSON body: %w", err)
	}
	httpReq, err := http.NewRequest("POST", "http://"+c.Address+"/api/chat", bytes.NewReader(res))
	if err != nil {
		return "", fmt.Errorf("could not create Ollama HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("could request summary via Ollama HTTP API: %w", err)
	}
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", fmt.Errorf("error while reading response from Ollama: %w", err)
	}
	if httpResp.StatusCode >= 400 {
		return "", fmt.Errorf("ollama returned an error code %d with body\n%s", httpResp.StatusCode, body)
	}
	var chatResponse OllamaChatResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return "", fmt.Errorf("error while decoding response from Ollama: %w", err)
	}
	return chatResponse.Message.Content, nil
}

// OllamaPullRequest HTTP body to send for the UpdateModel method.
type OllamaPullRequest struct {
	Name     string `json:"name"`
	Insecure bool   `json:"insecure"`
	Stream   bool   `json:"stream"`
}

// UpdateModel updates the Ollama model by pulling it.
func (c *OllamaClient) UpdateModel() error {
	body, err := json.Marshal(&OllamaPullRequest{
		Name:     c.Model,
		Insecure: false,
		Stream:   false,
	})
	if err != nil {
		return fmt.Errorf("could not encode request JSON body: %w", err)
	}
	httpReq, err := http.NewRequest("POST", "http://"+c.Address+"/api/pull", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("could not create Ollama HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("could request model update via Ollama HTTP API: %w", err)
	}
	body, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("error while reading response from Ollama: %w", err)
	}
	if httpResp.StatusCode >= 400 {
		return fmt.Errorf("ollama returned an error code %d with body\n%s", httpResp.StatusCode, body)
	}
	return nil
}
