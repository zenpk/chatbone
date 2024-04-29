package handler

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
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
	errCodeKey   string
	err          error
}

func New(conf *util.Configuration, logger util.ILogger,
	modelService *service.Model, oAuthService *service.OAuth, messageService *service.Message, openAiService *service.OpenAi,
) (*Handler, error) {
	h := new(Handler)
	h.conf = conf
	h.logger = logger
	h.errCodeKey = "errCode"
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
	h.setRoutes()
	return h, nil
}

func (h *Handler) jwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		accessTokenCookie, err := c.Cookie(CookieAccessToken)
		if err != nil {
			c.JSON(http.StatusOK, dto.CommonResp{Code: dto.ErrUnauthorized, Msg: fmt.Sprintf("get cookie failed: %v", err)})
			return nil
		}
		claims, err := util.VerifyAndParse(accessTokenCookie.Value, &h.rsaPublicKey)
		if err != nil {
			return errors.Join(err, h.err)
		}
		if claims.Issuer != h.conf.OAuthIssuer {
			c.JSON(http.StatusOK, dto.CommonResp{Code: dto.ErrUnauthorized, Msg: "wrong JWT issuer"})
			return nil
		}
		if !claims.IsValidAt(time.Now()) {
			c.JSON(http.StatusOK, dto.CommonResp{Code: dto.ErrUnauthorized, Msg: "JWT token expired"})
			return nil
		}
		// set user id to context
		c.Set("uuid", claims.Uuid)
		return next(c)
	}
}

func (h *Handler) setRoutes() {
	h.e.POST("/authorization", h.Authorization)
	h.e.POST("/refresh", h.Refresh)

	// auth needed
	h.e.Use(h.jwtMiddleware)
	h.e.POST("/chat", h.chat)
}

func (h *Handler) Shutdown(ctx context.Context) error {
	return h.e.Shutdown(ctx)
}

func (h *Handler) ListenAndServe() error {
	return h.e.Start(h.conf.HttpAddress)
}

// getTokenFromAuthorizationHeader not used
func (h *Handler) getTokenFromAuthorizationHeader(c echo.Context) (string, error) {
	authorization := c.Request().Header.Get("Authorization")
	if authorization == "" {
		return "", errors.New("missing Authorization header")
	}
	split := strings.Split(authorization, " ")
	if len(split) != 2 {
		return "", errors.New("invalid Authorization header")
	}
	token := split[1]
	return token, nil
}
