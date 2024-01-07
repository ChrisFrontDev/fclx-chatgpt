package entity

import "errors"

type ChatConfig struct {
	AIModel          *AIModel
	Temperature      float32
	TopP             float32
	N                int
	StopSequences    []string
	MaxTokens        int
	FrequencyPenalty float32
	PresencePenalty  float32
}

type Chat struct {
	Name                 string
	UserId               string
	InitialSystemMessage *Message
	Messages             []*Message
	OldMessages          []*Message
	Status               string
	TokenUsage           int
	Config               *ChatConfig
}

func NewChat(name string, userId string, initialSystemMessage *Message, config *ChatConfig) (*Chat, error) {
	chat := &Chat{
		Name:                 name,
		UserId:               userId,
		InitialSystemMessage: initialSystemMessage,
		Messages:             []*Message{},
		OldMessages:          []*Message{},
		Status:               "running",
		TokenUsage:           0,
		Config:               config,
	}

	chat.addMessage(initialSystemMessage)

	if err := chat.Validate(); err != nil {
		return nil, err
	}

	return chat, nil

}

func (c *Chat) addMessage(m *Message) error {
	if c.Status == "ended" {

		return errors.New("chat is ended")
	}
	for {
		if c.TokenUsage+m.GetQtdTokens() <= c.Config.MaxTokens {
			c.Messages = append(c.Messages, m)
			c.RefreshTokenUsage()
			break
		}
		c.OldMessages = append(c.OldMessages, c.Messages[0])
		c.Messages = c.Messages[1:]
		c.RefreshTokenUsage()
	}
	return nil
}

func (c *Chat) GetMessages() []*Message {
	return c.Messages
}

func (c *Chat) GetCountMessages() int {
	return len(c.Messages)
}

func (c *Chat) EndChat() {
	c.Status = "ended"
}

func (c *Chat) RefreshTokenUsage() {
	c.TokenUsage = 0
	for m := range c.Messages {
		c.TokenUsage += c.Messages[m].GetQtdTokens()
	}
}

func (c *Chat) Validate() error {
	if c.UserId == "" {
		return errors.New("user id is required")
	}
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Config == nil {
		return errors.New("config is required")
	}
	if c.Config.AIModel == nil {
		return errors.New("ai model is required")
	}
	if c.Status != "active" && c.Status != "ended" {
		return errors.New("invalid status")
	}
	if condition := c.Config.Temperature < 0 || c.Config.Temperature > 2; condition {
		return errors.New("invalid temperature")
		
	}
	return nil
}
