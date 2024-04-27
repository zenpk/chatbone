package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/zenpk/chatbone/dto"
)

func (h *Handler) chat(c echo.Context) error {
	req := new(dto.OpenAiReqFromClient)
	if err := c.Bind(req); err != nil {
		return err
	}
	uuid := c.Get("uuid").(string)
	// TODO
	replyChan := make(chan string)
	h.openAiService.Chat(uuid, req, replyChan)
	return nil
}
