package dto

type OpenAiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAiReq struct {
	Model    string          `json:"model"`
	Messages []OpenAiMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}
