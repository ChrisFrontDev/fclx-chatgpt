package entity

type AIModel struct {
	Name      string
	MaxTokens int
}

func NewAIModel(name string, maxTokens int) *AIModel {
	return &AIModel{
		Name:      name,
		MaxTokens: maxTokens,
	}
}

func (m *AIModel) GetMaxTokens() int {
	return m.MaxTokens
}

func (m *AIModel) getModelName() string {
	return m.Name
}
