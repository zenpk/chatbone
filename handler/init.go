package handler

import(
	"github.com/zenpk/chatbone/service"
)

type Handler struct{
	messageService *service.Message
	openAiService *service.OpenAi

}

func (h *Handler) Init() {
}
