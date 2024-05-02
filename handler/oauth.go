package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/zenpk/chatbone/dto"
)

func (h *Handler) Authorization(c echo.Context) error {
	reqBody := new(dto.AuthorizeReqFromClient)
	if err := c.Bind(reqBody); err != nil {
		c.Set(KeyErrCode, dto.ErrInput)
		return err
	}
	if reqBody.AuthorizationCode == "" || reqBody.CodeVerifier == "" {
		c.Set(KeyErrCode, dto.ErrInput)
		return errors.New("invalid input")
	}
	resp, err := h.oAuthService.Authorization(reqBody)
	if err != nil {
		c.Set(KeyErrCode, dto.ErrAuthFailed)
		return err
	}
	if err := h.setTokens(c, resp); err != nil {
		return err
	}
	return h.verifyResp(c)
}

func (h *Handler) Refresh(c echo.Context) error {
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
	if err := h.setTokens(c, resp); err != nil {
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

// Verify will first verify the access token, if not valid, it will try to refresh the token
func (h *Handler) Verify(c echo.Context) error {
	checkJwt := h.jwtMiddleware(func(c echo.Context) error { return nil })
	if err := checkJwt(c); err != nil {
		// access token is invalid
		return h.Refresh(c)
	}
	return h.verifyResp(c)
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

func (h *Handler) setTokens(c echo.Context, tokenResp *dto.RespFromOAuth) error {
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
			Path:     h.conf.CookiePathPrefix + "/refresh",
		})
	}
	return h.success(c)
}
