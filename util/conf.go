package util

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	TimeoutSecond      int      `json:"timeoutSecond"`
	HttpAddress        string   `json:"httpAddress"`
	LogFilePath        string   `json:"logFilePath"`
	Domain             string   `json:"domain"`
	AllowOrigins       []string `json:"allowOrigins"`
	AuthEnabled        bool     `json:"authEnabled"`
	CookiePathPrefix   string   `json:"cookiePathPrefix"`
	OAuthJwkPath       string   `json:"oAuthJwkPath"`
	OAuthAuthPath      string   `json:"oAuthAuthPath"`
	OAuthRefreshPath   string   `json:"oAuthRefreshPath"`
	OAuthClientId      string   `json:"oAuthClientId"`
	OAuthClientSecret  string   `json:"oAuthClientSecret"`
	OAuthIssuer        string   `json:"oAuthIssuer"`
	MongoDbUri         string   `json:"mongoDbUri"`
	MongoDbName        string   `json:"mongoDbName"`
	OpenAiOrgId        string   `json:"openAiOrgId"`
	OpenAiApiKey       string   `json:"openAiApiKey"`
	MessageLengthLimit int      `json:"messageLengthLimit"`
}

func NewConf(mode string) (*Configuration, error) {
	c := new(Configuration)
	filename := "conf-" + mode + ".json"
	confJson, err := os.ReadFile("./" + filename)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(confJson, &c); err != nil {
		return nil, err
	}
	return c, nil
}
