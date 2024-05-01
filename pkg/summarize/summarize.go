package summarize

import (
	_ "embed"

	"github.com/pkoukk/tiktoken-go"
	tokenLoader "github.com/pkoukk/tiktoken-go-loader"
	"github.com/sashabaranov/go-openai"
)

func init() {
	tiktoken.SetBpeLoader(tokenLoader.NewOfflineLoader())
}

//go:embed summary_system_prompt.txt
var summarySystemPrompt string

// NumTokensFromMessages gives a rough token count estimate.
func NumTokensFromMessages(messages []openai.ChatCompletionMessage) (numTokens int) {
	tkm, _ := tiktoken.GetEncoding("cl100k_base")

	tokensPerMessage := 3
	tokensPerName := 1

	for _, message := range messages {
		numTokens += tokensPerMessage
		numTokens += len(tkm.Encode(message.Content, nil, nil))
		numTokens += len(tkm.Encode(message.Role, nil, nil))
		numTokens += len(tkm.Encode(message.Name, nil, nil))
		if message.Name != "" {
			numTokens += tokensPerName
		}
	}
	numTokens += 3 // every reply is primed with <|start|>assistant<|message|>
	return numTokens
}
