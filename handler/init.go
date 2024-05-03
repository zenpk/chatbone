package handler

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zenpk/chatbone/dto"
	"github.com/zenpk/chatbone/service"
	"github.com/zenpk/chatbone/util"
)

type Handler struct {
	modelService   *service.Model
	oAuthService   *service.OAuth
	messageService *service.Message
	openAiService  *service.OpenAi

	e            *echo.Echo
	conf         *util.Configuration
	logger       util.ILogger
	rsaPublicKey rsa.PublicKey
	err          error
}

func New(conf *util.Configuration, logger util.ILogger,
	modelService *service.Model, oAuthService *service.OAuth, messageService *service.Message, openAiService *service.OpenAi,
) (*Handler, error) {
	h := new(Handler)
	h.conf = conf
	h.logger = logger
	h.err = errors.New("at Handler")

	h.modelService = modelService
	h.oAuthService = oAuthService
	h.messageService = messageService
	h.openAiService = openAiService

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
	nBytes, err := base64.RawURLEncoding.DecodeString(publicKey.N)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("invalid RSA public key field: N: %w", err), h.err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(publicKey.E)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("invalid RSA public key field: E: %w", err), h.err)
	}
	h.rsaPublicKey = rsa.PublicKey{
		N: big.NewInt(0).SetBytes(nBytes),
		E: int(big.NewInt(0).SetBytes(eBytes).Int64()),
	}

	h.e = echo.New()
	h.e.Use(middleware.Recover())
	h.e.HTTPErrorHandler = func(err error, c echo.Context) {
		errCode, ok := c.Get(KeyErrCode).(int)
		if !ok {
			errCode = dto.ErrUnknown
		}
		if err := c.JSON(http.StatusOK, dto.CommonResp{Code: errCode, Msg: err.Error()}); err != nil {
			h.e.Logger.Print(err)
		}
	}
	h.e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     h.conf.AllowOrigins,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"Origin", "X-Requested-With", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))
	h.e.Use(middleware.BodyLimit("2M"))
	h.e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(KeyUsername, "unknown user")
			return next(c)
		}
	})
	h.e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogError:    true,
		LogStatus:   true,
		LogMethod:   true,
		LogURIPath:  true,
		LogRemoteIP: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			errCode, ok := c.Get(KeyErrCode).(int)
			if !ok {
				errCode = dto.ErrUnknown
			}
			h.logger.Printf("%v | %v | %-7s | %v | %v | %v\n", v.Status, errCode, v.Method, v.URIPath, v.RemoteIP, c.Get(KeyUsername))
			return v.Error
		},
	}))
	h.setRoutes()
	return h, nil
}

func (h *Handler) setRoutes() {
	h.e.POST("/authorize", h.Authorize)
	h.e.POST("/refresh", h.Refresh)

	// auth group
	g := h.e.Group("/")
	g.Use(h.jwtMiddleware)
	g.POST("/chat", h.chat)
}

func (h *Handler) jwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, err := h.jwtCheck(c)
		if err != nil {
			return c.JSON(http.StatusOK, dto.CommonResp{Code: dto.ErrAuthFailed, Msg: err.Error()})
		}
		// set user id to context
		c.Set(KeyUuid, claims.Uuid)
		c.Set(KeyUsername, claims.Username)
		return next(c)
	}
}

func (h *Handler) jwtCheck(c echo.Context) (*util.Claims, error) {
	accessTokenCookie, err := c.Cookie(CookieAccessToken)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("get cookie failed: %w", err), h.err)
	}
	claims, err := util.VerifyAndParse(accessTokenCookie.Value, &h.rsaPublicKey)
	if err != nil {
		return nil, errors.Join(err, h.err)
	}
	if claims.Issuer != h.conf.OAuthIssuer {
		return nil, errors.Join(errors.New("wrong JWT issuer"), h.err)
	}
	if !claims.IsValidAt(time.Now()) {
		return nil, errors.Join(errors.New("JWT token expired"), h.err)
	}
	return claims, nil
}

func (h *Handler) success(c echo.Context) error {
	return c.JSON(http.StatusOK, dto.CommonResp{Code: dto.ErrOk, Msg: "success"})
}

func (h *Handler) Shutdown(ctx context.Context) error {
	return h.e.Shutdown(ctx)
}

func (h *Handler) ListenAndServe() error {
	return h.e.Start(h.conf.HttpAddress)
}
