package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/zenpk/chatbone/dto"
	"github.com/zenpk/chatbone/util"
)

type OAuth struct {
	conf   *util.Configuration
	logger util.ILogger
	err    error
}

func NewOAuth(conf *util.Configuration, logger util.ILogger) (*OAuth, error) {
	o := new(OAuth)
	o.conf = conf
	o.logger = logger
	o.err = errors.New("at OAuth service")
	return o, nil
}

func (o *OAuth) Authorize(reqBody *dto.AuthorizeReqFromClient) (*dto.RespFromOAuth, error) {
	reqByte, err := json.Marshal(dto.AuthorizeReqToOAuth{
		ClientInfo: dto.ClientInfo{
			ClientId:     o.conf.OAuthClientId,
			ClientSecret: o.conf.OAuthClientSecret,
		},
		AuthorizeReqFromClient: *reqBody,
	})
	if err != nil {
		return nil, errors.Join(err, o.err)
	}
	req, err := http.NewRequest("POST", o.conf.OAuthAuthPath, bytes.NewBuffer(reqByte))
	if err != nil {
		return nil, errors.Join(err, o.err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: time.Duration(o.conf.TimeoutSecond) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Join(err, o.err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Join(errors.New("OAuth authorization failed, error code: "+fmt.Sprintf("%v", resp.StatusCode)), o.err)
	}
	respBody := new(dto.RespFromOAuth)
	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return nil, errors.Join(err, o.err)
	}
	return respBody, nil
}

func (o *OAuth) Refresh(refreshToken string) (*dto.RespFromOAuth, error) {
	reqByte, err := json.Marshal(dto.RefreshReqToOAuth{
		ClientInfo: dto.ClientInfo{
			ClientId:     o.conf.OAuthClientId,
			ClientSecret: o.conf.OAuthClientSecret,
		},
		RefreshToken: refreshToken,
	})
	if err != nil {
		return nil, errors.Join(err, o.err)
	}
	req, err := http.NewRequest("POST", o.conf.OAuthRefreshPath, bytes.NewBuffer(reqByte))
	if err != nil {
		return nil, errors.Join(err, o.err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: time.Duration(o.conf.TimeoutSecond) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Join(err, o.err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Join(errors.New("OAuth refresh failed, error code: "+fmt.Sprintf("%v", resp.StatusCode)), o.err)
	}
	respBody := new(dto.RespFromOAuth)
	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return nil, errors.Join(err, o.err)
	}
	return respBody, nil
}
