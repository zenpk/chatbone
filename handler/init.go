package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zenpk/chatbone/dto"
	"github.com/zenpk/chatbone/service"
	"github.com/zenpk/chatbone/util"
)

type Handler struct {
	messageService *service.Message
	openAiService  *service.OpenAi

	e    *echo.Echo
	conf *util.Configuration
}

func Init(conf *util.Configuration, messageService *service.Message, openAiService *service.OpenAi) (*Handler, error) {
	h := new(Handler)
	h.conf = conf
	e := echo.New()
	h.e = e
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		errCode, ok := c.Get("code").(int)
		if !ok {
			errCode = dto.ErrUnknown
		}
		if err := c.JSON(http.StatusOK, dto.CommonResp{Code: errCode, Msg: err.Error()}); err != nil {
			e.Logger.Print(err)
		}
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"*"},
		AllowHeaders: []string{"*"},
	}))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogError:    true,
		LogStatus:   true,
		LogMethod:   true,
		LogURIPath:  true,
		LogRemoteIP: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Printf("| %v | %-7s | %v | %v\n", v.Status, v.Method, v.URIPath, v.RemoteIP)
			return v.Error
		},
	}))

	return h, nil
}

func (h *Handler) Shutdown(ctx context.Context) error {
	return h.e.Shutdown(ctx)
}

func (h *Handler) ListenAndServe() error {
	return h.e.Start(h.conf.HttpAddress)
}
