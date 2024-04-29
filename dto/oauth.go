package dto

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

type RefreshReqFromClient struct {
	RefreshToken string `json:"refreshToken"`
}

type RefreshReqToOAuth struct {
	ClientInfo
	RefreshReqFromClient
}

type RespFromOAuth struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
