package chatcompletionstream

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/ChrisFrontDev/fclx/chatservice/internal/domain/entity"
	"github.com/ChrisFrontDev/fclx/chatservice/internal/domain/gateway"
	openai "github.com/sashabaranov/go-openai"
)

type ChatCompletionConfigInputDTO struct {
	Model                string
	ModelMaxTokens       int
	Temperature          float32
	TopP                 float32
	N                    int
	Stop                 []string
	MaxTokens            int
	PresencePenalty      float32
	FrequencyPenalty     float32
	InitialSystemMessage string
}

type ChatCompletionInputDTO struct {
	ChatID      string
	UserID      string
	UserMessage string
	Config      ChatCompletionConfigInputDTO
}

type ChatCompletionOutputDTO struct {
	ChatID  string
	UserID  string
	Content string
}

type ChatCompletionUseCase struct {
	ChatGateway  gateway.ChatGateway
	OpenAIClient *openai.Client
	Stream       chan ChatCompletionOutputDTO
}

func NewChatCompletionUseCase(chatGateway gateway.ChatGateway, openAIClient *openai.Client, stream chan ChatCompletionOutputDTO) *ChatCompletionUseCase {
	return &ChatCompletionUseCase{
		ChatGateway:  chatGateway,
		OpenAIClient: openAIClient,
		Stream:       stream,
	}
}

func (uc *ChatCompletionUseCase) Execute(ctx context.Context, input ChatCompletionInputDTO) (*ChatCompletionOutputDTO, error) {
	chat, err := uc.ChatGateway.FindChatById(ctx, input.ChatID)
	if err != nil {
		if err.Error() == "chat not found" {
			// create new chat entity
			chat, err = createNewChat(input)
			if err != nil {
				return nil, errors.New("failed to create new chat" + err.Error())
			}
			// save on database
			err = uc.ChatGateway.CreateChat(ctx, chat)
			if err != nil {
				return nil, errors.New("failed to saving new chat" + err.Error())
			}
		} else {
			return nil, errors.New("failed to find existing chat" + err.Error())
		}
		return nil, err
	}

	// get user message
	userMessage, err := entity.NewMessage("user", input.UserMessage, chat.Config.AIModel)
	if err != nil {
		return nil, errors.New("failed to create user message" + err.Error())
	}

	err = chat.AddMessage(userMessage)
	if err != nil {
		return nil, errors.New("failed to add user message" + err.Error())
	}

	messages := []openai.ChatCompletionMessage{}
	for _, message := range chat.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.Content,
		})
	}

	resp, err := uc.OpenAIClient.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:            chat.Config.AIModel.Name,
		Messages:         messages,
		Temperature:      chat.Config.Temperature,
		TopP:             chat.Config.TopP,
		N:                chat.Config.N,
		Stop:             chat.Config.StopSequences,
		MaxTokens:        chat.Config.MaxTokens,
		PresencePenalty:  chat.Config.PresencePenalty,
		FrequencyPenalty: chat.Config.FrequencyPenalty,
		Stream:           true,
	})
	if err != nil {
		return nil, errors.New("failed to create chat completion stream" + err.Error())
	}

	var fullResponse strings.Builder
	for {
		response, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, errors.New("failed to receive chat completion stream" + err.Error())
		}
		fullResponse.WriteString(response.Choices[0].Delta.Content)
		r := ChatCompletionOutputDTO{
			ChatID:  chat.ChatId,
			UserID:  input.UserID,
			Content: fullResponse.String(),
		}
		uc.Stream <- r
	}

	assistant, err := entity.NewMessage("assistant", fullResponse.String(), chat.Config.AIModel)
	if err != nil {
		return nil, errors.New("failed to create assistant message" + err.Error())
	}
	err = chat.AddMessage(assistant)
	if err != nil {
		return nil, errors.New("failed to add assistant message" + err.Error())
	}

	err = uc.ChatGateway.SaveChat(ctx, chat)
	if err != nil {
		return nil, errors.New("failed to save chat" + err.Error())
	}

	return &ChatCompletionOutputDTO{
		ChatID:  chat.ChatId,
		UserID:  input.UserID,
		Content: fullResponse.String(),
	}, nil
}

func createNewChat(input ChatCompletionInputDTO) (*entity.Chat, error) {
	model := entity.NewAIModel(input.Config.Model, input.Config.ModelMaxTokens)
	chatConfig := &entity.ChatConfig{
		Temperature:      input.Config.Temperature,
		TopP:             input.Config.TopP,
		N:                input.Config.N,
		StopSequences:    input.Config.Stop,
		MaxTokens:        input.Config.MaxTokens,
		PresencePenalty:  input.Config.PresencePenalty,
		FrequencyPenalty: input.Config.FrequencyPenalty,
		AIModel:          model,
	}

	initalMessage, err := entity.NewMessage("system", input.Config.InitialSystemMessage, model)
	if err != nil {
		return nil, errors.New("failed to create initial message" + err.Error())
	}
	chat, err := entity.NewChat(input.UserID, initalMessage, chatConfig)
	if err != nil {
		return nil, errors.New("failed to create chat" + err.Error())
	}
	return chat, nil
}
