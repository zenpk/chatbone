package dto

type OpenAiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAiReqFromClient struct {
	ModelId   int             `json:"modelId"`
	SessionId string          `json:"sessionId"`
	Messages  []OpenAiMessage `json:"messages"`
}

type OpenAiReqToOpenAi struct {
	OpenAiReqFromClient
	Stream bool `json:"stream"`
}
