package dto

type CommonResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type ChatReqFromClient struct {
	ModelId   int    `json:"modelId"`
	SessionId string `json:"sessionId"`
	Messages  any    `json:"messages"`
}
