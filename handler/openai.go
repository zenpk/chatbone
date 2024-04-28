package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/zenpk/chatbone/dto"
)

func (h *Handler) chat(c echo.Context) error {
	const ChanSize = 1024
	req := new(dto.OpenAiReqFromClient)
	if err := c.Bind(req); err != nil {
		return err
	}
	uuid := c.Get("uuid").(string)
	replyChan := make(chan string, ChanSize)
	errChan := make(chan error, 1)
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
	c.Response().Header().Set(echo.HeaderConnection, "keep-alive")
	go func() {
		if err := h.openAiService.Chat(uuid, req, replyChan); err != nil {
			errChan <- err
		}
	}()
	for {
		select {
		case reply := <-replyChan:
			if _, err := c.Response().Write([]byte(reply)); err != nil {
				return err
			}
			if reply == dto.MessageEnding {
				return nil
			}
		case err := <-errChan:
			return err
		}
	}
}
