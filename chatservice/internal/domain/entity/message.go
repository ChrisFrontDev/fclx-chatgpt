package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        string
	Role      string
	Content   string
	Model     *AIModel
	CreatedAt time.Time
	Tokens    int
}

func NewMessage(role string, content string, model *AIModel) (*Message, error) {
	msg := &Message{
		ID:        uuid.New().String(),
		Role:      role,
		Content:   content,
		Model:     model,
		CreatedAt: time.Now(),
	}

	err := msg.Validate()
	if err != nil {
		return nil, err
	}

	return msg, nil

}

func (m *Message) Validate() error {
	if m.Role != "user" && m.Role != "system" {
		return errors.New("invalid role")
	}

	if m.Content == "" {
		return errors.New("content is empty")
	}

	if m.CreatedAt.IsZero() {
		return errors.New("created_at is empty, problem with time.Now()")
	}

	return nil
}

func (m *Message) GetQtdTokens() int {
	return m.Tokens
}
