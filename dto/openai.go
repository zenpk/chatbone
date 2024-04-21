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

type OpenAiResp struct{
{"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"gpt-3.5-turbo-0125", "system_fingerprint": "fp_44709d6fcb", "choices":[{"index":0,"delta":{"role":"assistant","content":""},"logprobs":null,"finish_reason":null}]}
}
