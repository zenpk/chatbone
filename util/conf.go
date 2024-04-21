package util

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	TimeoutSecond    int    `json:"timeoutSecond"`
	HttpAddress      string `json:"httpAddress"`
	LogFilePath      string `json:"logFilePath"`
	OAuthRefreshPath string `json:"oAuthRefreshPath"`
	OAuthJwkPath     string `json:"oAuthJwkPath"`
	JwtIssuer        string `json:"jwtIssuer"`
	JwtClientId      string `json:"jwtClientId"`
	MongoDbUri       string `json:"mongoDbUri"`
	MongoDbName      string `json:"mongoDbName"`
	OpenAiOrgId      string `json:"openAiOrgId"`
	OpenAiApiKey     string `json:"openAiApiKey"`
	ReqLengthLimit   int    `json:"reqLengthLimit"`
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
