package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/zenpk/chatbone/dal"
	"github.com/zenpk/chatbone/dto"
)

func (h *Handler) setStreamHeaders(c echo.Context) {
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream; charset=utf-8")
	c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
	c.Response().Header().Set(echo.HeaderConnection, "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no") // necessary for disabling NGINX buffer
	c.Response().WriteHeader(http.StatusOK)
}

func (h *Handler) chat(c echo.Context) error {
	const ChanSize = 1024
	req := new(dto.ChatReqFromClient)
	if err := c.Bind(req); err != nil {
		c.Set(KeyErrCode, dto.ErrInput)
		return err
	}
	uuid := c.Get(KeyUuid).(string)
	replyChan := make(chan any, ChanSize)
	errChan := make(chan error, 1)
	// get and check model
	model, err := h.modelService.GetAndCheckModelById(req.ModelId)
	if err != nil {
		return err
	}

	switch model.Provider {
	case dal.ProviderOpenAi:
		convertedMessages, ok := req.Messages.([]dto.OpenAiMessage)
		if !ok {
			c.Set(KeyErrCode, dto.ErrInput)
			return errors.Join(errors.New("input messages format error"), h.err)
		}
		convertedReq := &dto.OpenAiReqFromClient{
			ModelId:   req.ModelId,
			SessionId: req.SessionId,
			Messages:  convertedMessages,
		}
		go func() {
			errChan <- h.openAiService.Chat(uuid, model, convertedReq, replyChan)
		}()
		h.setStreamHeaders(c)
		for {
			select {
			case reply := <-replyChan:
				convertedResp := reply.(dto.OpenAiResp)
				event := Event{
					Data: []byte(convertedResp.Choices[0].Delta.Content),
				}
				if err := event.MarshalTo(c.Response()); err != nil {
					return err
				}
				c.Response().Flush()
			case err := <-errChan:
				if err != nil {
					h.logger.Errorf("chat error: %v", err)
				}
				return err
			}
		}
	default:
		return errors.Join(errors.New("model provider not supported"), h.err)
	}
}
