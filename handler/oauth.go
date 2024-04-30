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
		c.Set(ErrCodeKey, dto.ErrInput)
		return err
	}
	if reqBody.AuthorizationCode == "" || reqBody.CodeVerifier == "" {
		c.Set(ErrCodeKey, dto.ErrInput)
		return errors.New("invalid input")
	}
	resp, err := h.oAuthService.Authorization(reqBody)
	if err != nil {
		c.Set(ErrCodeKey, dto.ErrAuthFailed)
		return err
	}
	if err := h.setTokens(c, resp); err != nil {
		return err
	}
	return nil
}

func (h *Handler) Refresh(c echo.Context) error {
	refreshTokenCookie, err := c.Cookie(CookieRefreshToken)
	if err != nil {
		c.Set(ErrCodeKey, dto.ErrInput)
		return err
	}
	if refreshTokenCookie.Value == "" {
		c.Set(ErrCodeKey, dto.ErrInput)
		return errors.New("refresh token is empty")
	}
	resp, err := h.oAuthService.Refresh(refreshTokenCookie.Value)
	if err != nil {
		c.Set(ErrCodeKey, dto.ErrRefreshFailed)
		return err
	}
	if err := h.setTokens(c, resp); err != nil {
		return err
	}
	return nil
}

func (h *Handler) setTokens(c echo.Context, tokenResp *dto.RespFromOAuth) error {
	if tokenResp == nil || tokenResp.AccessToken == "" || tokenResp.RefreshToken == "" {
		return errors.New("invalid token response")
	}
	// no expires for cookies for simplicity
	// the OAuth will check for the expiration
	c.SetCookie(&http.Cookie{
		Name:     CookieAccessToken,
		Value:    tokenResp.AccessToken,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})
	c.SetCookie(&http.Cookie{
		Name:     CookieRefreshToken,
		Value:    tokenResp.RefreshToken,
		HttpOnly: true,
		Secure:   true,
		Path:     "/refresh",
	})
	return nil
}
