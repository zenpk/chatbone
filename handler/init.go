package handler

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zenpk/chatbone/dto"
	"github.com/zenpk/chatbone/service"
	"github.com/zenpk/chatbone/util"
)

type Handler struct {
	messageService *service.Message
	openAiService  *service.OpenAi

	e            *echo.Echo
	conf         *util.Configuration
	rsaPublicKey rsa.PublicKey
	errCodeKey   string
	err          error
}

func New(conf *util.Configuration, messageService *service.Message, openAiService *service.OpenAi) (*Handler, error) {
	h := new(Handler)
	h.conf = conf
	h.errCodeKey = "errCode"
	h.err = errors.New("at Handler")
	// get JWK from the OAuth 2.0 endpoint
	client := http.Client{
		Timeout: time.Duration(h.conf.TimeoutSecond) * time.Second,
	}
	resp, err := client.Get(h.conf.OAuthJwkPath)
	if err != nil {
		return nil, errors.Join(err, h.err)
	}
	var publicKey util.Jwk
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&publicKey); err != nil {
		return nil, errors.Join(err, h.err)
	}
	rsaN := new(big.Int)
	rsaN, ok := rsaN.SetString(publicKey.N, 10)
	if !ok {
		return nil, errors.Join(errors.New("invalid RSA public key field: N"), h.err)
	}
	rsaE, err := strconv.ParseInt(publicKey.E, 10, 64)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("invalid RSA public key field: E: %w", err), h.err)
	}
	h.rsaPublicKey = rsa.PublicKey{
		N: rsaN,
		E: int(rsaE),
	}

	h.e = echo.New()
	h.e.Use(middleware.Recover())
	h.e.HTTPErrorHandler = func(err error, c echo.Context) {
		errCode, ok := c.Get(h.errCodeKey).(int)
		if !ok {
			errCode = dto.ErrUnknown
		}
		if err := c.JSON(http.StatusOK, dto.CommonResp{Code: errCode, Msg: err.Error()}); err != nil {
			h.e.Logger.Print(err)
		}
	}
	h.e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"*"},
		AllowHeaders: []string{"*"},
	}))
	h.e.Use(middleware.BodyLimit("2M"))
	h.e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
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
	h.e.Use(h.jwtMiddleware)
	h.setRoutes()
	return h, nil
}

func (h *Handler) jwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authorization := c.Request().Header.Get("Authorization")
		if authorization == "" {
			c.JSON(http.StatusOK, dto.CommonResp{Code: dto.ErrUnauthorized, Msg: "missing Authorization header"})
			return nil
		}
		split := strings.Split(authorization, " ")
		if len(split) != 2 {
			c.JSON(http.StatusOK, dto.CommonResp{Code: dto.ErrUnauthorized, Msg: "invalid Authorization header"})
			return nil
		}
		token := split[1]
		claims, err := util.VerifyAndParse(token, &h.rsaPublicKey)
		if err != nil {
			return errors.Join(err, h.err)
		}
		if claims.Issuer != h.conf.JwtIssuer {
			c.JSON(http.StatusOK, dto.CommonResp{Code: dto.ErrUnauthorized, Msg: "wrong JWT issuer"})
			return nil
		}
		if !claims.IsValidAt(time.Now()) {
			c.JSON(http.StatusOK, dto.CommonResp{Code: dto.ErrUnauthorized, Msg: "JWT token expired"})
			return nil
		}
		return next(c)
	}
}

func (h *Handler) setRoutes() {
	h.e.POST("/chat", h.chat)
}

func (h *Handler) Shutdown(ctx context.Context) error {
	return h.e.Shutdown(ctx)
}

func (h *Handler) ListenAndServe() error {
	return h.e.Start(h.conf.HttpAddress)
}
