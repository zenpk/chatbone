package util

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	HttpAddress   string `json:"httpAddress"`
	TimeoutSecond int64  `json:"timeoutSecond"`
	OAuthEndPoint string `json:"oAuthEndPoint"`
	JwtIssuer     string `json:"jwtIssuer"`
	LogFilePath   string `json:"logFilePath"`
	MongoDbUri    string `json:"mongoDbUri"`
	MongoDbName   string `json:"mongoDbName"`
	OpenAiOrgId   string `json:"openAiOrgId"`
	OpenAiApiKey  string `json:"openAiApiKey"`
}

func (c *Configuration) Init(mode string) error {
	filename := "conf-" + mode + ".json"
	confJson, err := os.ReadFile("./" + filename)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(confJson, &c); err != nil {
		return err
	}
	return nil
}
