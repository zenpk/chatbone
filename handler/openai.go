package handler

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/zenpk/chatbone/dal"
	"github.com/zenpk/chatbone/dto"
)

func (h *Handler) chat(c echo.Context) error {
	const ChanSize = 1024
	req := new(dto.OpenAiReqFromClient) // TODO change to dto.ChatReqFromClient
	if err := c.Bind(req); err != nil {
		c.Set(KeyErrCode, dto.ErrInput)
		return err
	}
	uuid := c.Get(KeyUuid).(string)
	replyChan := make(chan any, ChanSize)
	errChan := make(chan error, 1)
	w := c.Response()
	w.Header().Set(echo.HeaderContentType, "text/event-stream; charset=utf-8")
	w.Header().Set(echo.HeaderCacheControl, "no-cache")
	w.Header().Set(echo.HeaderConnection, "keep-alive")
	// get and check model
	model, err := h.modelService.GetAndCheckModelById(req.ModelId)
	if err != nil {
		return err
	}
	switch model.Provider {
	case dal.ProviderOpenAi:
		go func() {
			errChan <- h.openAiService.Chat(uuid, model, req, replyChan)
		}()
		for {
			select {
			case reply := <-replyChan:
				converted := reply.(dto.OpenAiResp)
				event := Event{
					Data: []byte(converted.Choices[0].Delta.Content),
				}
				if err := event.MarshalTo(w); err != nil {
					return err
				}
				w.Flush()
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
