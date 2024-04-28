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
	Model    string          `json:"model"`
	Messages []OpenAiMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type OpenAiResp struct {
	Id                string          `json:"id"`
	Object            string          `json:"object"`
	Created           int64           `json:"created"`
	Model             string          `json:"model"`
	SystemFingerprint string          `json:"system_fingerprint"`
	Choices           []*OpenAiChoice `json:"choices"`
}

type OpenAiChoice struct {
	Index        int            `json:"index"`
	Delta        *OpenAiMessage `json:"delta"`
	Logprobs     interface{}    `json:"logprobs"`
	FinishReason string         `json:"finish_reason"`
}
