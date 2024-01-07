package chatcompletionstream

import (
	"github.com/ChrisFrontDev/fclx/chatservice/internal/domain/gateway"
	openai "github.com/sashabaranov/go-openai"
)


type ChatCpletionUseCase struct{
	ChatGateway *gateway.ChatGateway
	OpenAIClient *openai.Client
}

func NewChatCompletionUseCase(chatGateway *gateway.ChatGateway, openAIClient *openai.Client) *ChatCpletionUseCase {
	return &ChatCpletionUseCase{
		ChatGateway: chatGateway,
		OpenAIClient: openAIClient,
	}
}