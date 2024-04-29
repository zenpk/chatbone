package handler

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/zenpk/chatbone/dto"
)

func (h *Handler) chat(c echo.Context) error {
	const ChanSize = 1024
	req := new(dto.OpenAiReqFromClient)
	if err := c.Bind(req); err != nil {
		c.Set(h.errCodeKey, dto.ErrInput)
		return err
	}
	uuid := c.Get("uuid").(string)
	replyChan := make(chan string, ChanSize)
	errChan := make(chan error, 1)
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
	c.Response().Header().Set(echo.HeaderConnection, "keep-alive")
	enc := json.NewEncoder(c.Response())
	go func() {
		errChan <- h.openAiService.Chat(uuid, req, replyChan)
	}()
	for {
		select {
		case reply := <-replyChan:
			if err := enc.Encode(dto.CommonResp{
				Code: dto.ErrOk,
				Msg:  reply,
			}); err != nil {
				return err
			}
			c.Response().Flush()
			if reply == dto.MessageEnding {
				c.Response().WriteHeader(http.StatusOK)
			}
		case err := <-errChan:
			if err != nil {
				h.logger.Errorf("chat error: %v", err)
			}
			return err
		}
	}
}
