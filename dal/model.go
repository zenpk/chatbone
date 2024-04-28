package dal

type Model struct {
	Id           int     `json:"id"`
	Name         string  `json:"name"`
	Encoding     string  `json:"encoding"`
	Provider     string  `json:"provider"`
	InRate       float64 `json:"inRate"`
	OutRate      float64 `json:"outRate"`
	SupportImage bool    `json:"supportImage"`

	hardcoded []*Model
}

func newModel() (*Model, error) {
	m := new(Model)
	m.hardcoded = append(m.hardcoded, &Model{
		Id:           ModelIdOpenAiGpt4,
		Name:         "gpt-4-turbo",
		Encoding:     "cl100k_base",
		Provider:     ProviderOpenAi,
		InRate:       0.00001,
		OutRate:      0.00003,
		SupportImage: false,
	}, &Model{
		Id:           ModelIdOpenAiGpt35,
		Name:         "gpt-3.5-turbo",
		Encoding:     "cl100k_base",
		Provider:     ProviderOpenAi,
		InRate:       0.0000005,
		OutRate:      0.0000015,
		SupportImage: false,
	})
	return m, nil
}

func (m *Model) SelectAll() ([]*Model, error) {
	return m.hardcoded, nil
}

func (m *Model) SelectById(id int) (*Model, error) {
	for _, v := range m.hardcoded {
		if v.Id == id {
			return v, nil
		}
	}
	return nil, nil
}
