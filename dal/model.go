package dal

type Model struct {
	Name         string  `json:"name"`
	Encoding     string  `json:"encoding"`
	ProviderId   int     `json:"providerId"`
	InRate       float64 `json:"inRate"`
	OutRate      float64 `json:"outRate"`
	SupportImage bool    `json:"supportImage"`

	hardcoded []*Model
}

func initModel() (*Model, error) {
	m := new(Model)
	m.hardcoded = append(m.hardcoded, &Model{
		Name:         "gpt-4-turbo",
		Encoding:     "cl100k_base",
		ProviderId:   ProviderOpenAi,
		InRate:       0.00001,
		OutRate:      0.00003,
		SupportImage: false,
	}, &Model{
		Name:         "gpt-3.5-turbo",
		Encoding:     "cl100k_base",
		ProviderId:   ProviderOpenAi,
		InRate:       0.0000005,
		OutRate:      0.0000015,
		SupportImage: false,
	})
	return m, nil
}

func (m *Model) SelectAll() ([]*Model, error) {
	return m.hardcoded, nil
}

func (m *Model) SelectByProviderAndName(provider int, name string) (*Model, error) {
	for _, v := range m.hardcoded {
		if v.ProviderId == provider && v.Name == name {
			return v, nil
		}
	}
	return nil, nil
}
