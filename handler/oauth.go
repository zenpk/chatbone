package handler

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/zenpk/chatbone/dto"
)

func (h *Handler) Authorize(c echo.Context) error {
	reqBody := new(dto.AuthorizeReqFromClient)
	if err := c.Bind(reqBody); err != nil {
		c.Set(KeyErrCode, dto.ErrInput)
		return err
	}
	if reqBody.AuthorizationCode == "" || reqBody.CodeVerifier == "" {
		c.Set(KeyErrCode, dto.ErrInput)
		return errors.New("invalid input")
	}
	resp, err := h.oAuthService.Authorize(reqBody)
	if err != nil {
		c.Set(KeyErrCode, dto.ErrAuthFailed)
		return err
	}
	if err := h.setCookies(c, resp); err != nil {
		return err
	}
	return h.verifyResp(c)
}

// Refresh will first verify the access token, if not valid, it will try to refresh the token
func (h *Handler) Refresh(c echo.Context) error {
	if _, err := h.jwtCheck(c); err == nil {
		// access token is still valid
		return h.verifyResp(c)
	}
	refreshTokenCookie, err := c.Cookie(CookieRefreshToken)
	if err != nil {
		c.Set(KeyErrCode, dto.ErrInput)
		return err
	}
	if refreshTokenCookie.Value == "" {
		c.Set(KeyErrCode, dto.ErrInput)
		return errors.New("refresh token is empty")
	}
	resp, err := h.oAuthService.Refresh(refreshTokenCookie.Value)
	if err != nil {
		c.Set(KeyErrCode, dto.ErrRefreshFailed)
		return err
	}
	if err := h.setCookies(c, resp); err != nil {
		return err
	}
	// if the refresh request comes with a body
	// then it is a quick refresh request, proceed with the corresponding action
	req := new(dto.QuickRefreshReq)
	if err := c.Bind(req); err != nil {
		c.Set(KeyErrCode, dto.ErrInput)
		return errors.New("invalid refresh request body, the request should at least contain empty body")
	}
	switch req.Action {
	case ActionChat:
		return h.chat(c)
	default:
		return h.verifyResp(c)
	}
}

// verifyResp returns some initial data (e.g. models) to the client
// this can reduce a round-trip on initial loading
func (h *Handler) verifyResp(c echo.Context) error {
	models, err := h.modelService.GetAll()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.VerifyResp{
		CommonResp: dto.CommonResp{Code: dto.ErrOk, Msg: "success"},
		Models:     models,
	})
}

func (h *Handler) setCookies(c echo.Context, tokenResp *dto.RespFromOAuth) error {
	if tokenResp == nil || tokenResp.AccessToken == "" {
		return errors.New("invalid token response")
	}
	// no expires for cookies for simplicity
	// the OAuth will check for the expiration
	c.SetCookie(&http.Cookie{
		Name:     CookieAccessToken,
		Value:    tokenResp.AccessToken,
		HttpOnly: true,
		Secure:   true,
		Domain:   h.conf.Domain,
		Path:     h.conf.CookiePathPrefix + "/",
	})
	// refresh token could be nil if it's returned from refresh request
	if tokenResp.RefreshToken != "" {
		c.SetCookie(&http.Cookie{
			Name:     CookieRefreshToken,
			Value:    tokenResp.RefreshToken,
			HttpOnly: true,
			Secure:   true,
			Domain:   h.conf.Domain,
			Path:     h.conf.CookiePathPrefix + "/refresh",
		})
	}
	// info token, basically the same as access token payload
	split := strings.Split(tokenResp.AccessToken, ".")
	if len(split) != 3 {
		return errors.New("invalid access token")
	}
	infoToken := make([]byte, 1024)
	n, err := base64.RawStdEncoding.Decode(infoToken, []byte(split[1]))
	if err != nil {
		return err
	}
	c.SetCookie(&http.Cookie{
		Name:     CookieInfoToken,
		Value:    string(infoToken[:n]),
		HttpOnly: false,
		Secure:   true,
		Domain:   h.conf.Domain,
		Path:     h.conf.CookiePathPrefix + "/",
	})
	return nil
}
