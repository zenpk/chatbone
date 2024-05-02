package dto

import "github.com/zenpk/chatbone/dal"

type ClientInfo struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type AuthorizeReqFromClient struct {
	AuthorizationCode string `json:"authorizationCode"`
	CodeVerifier      string `json:"codeVerifier"`
}

type AuthorizeReqToOAuth struct {
	ClientInfo
	AuthorizeReqFromClient
}

type QuickRefreshReq struct {
	Action string `json:"action"`
}

type RefreshReqToOAuth struct {
	ClientInfo
	RefreshToken string `json:"refreshToken"`
}

type RespFromOAuth struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type VerifyResp struct {
	CommonResp
	Models []*dal.Model `json:"models"`
}
