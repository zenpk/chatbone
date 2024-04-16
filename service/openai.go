package service

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/zenpk/chatbone/dal"
	"github.com/zenpk/chatbone/dto"
	"github.com/zenpk/chatbone/util"
)

type OpenAi struct {
	conf *util.Configuration
	db   *dal.Database
}

func InitOpenAi(conf *util.Configuration, db *dal.Database) (*OpenAi, error) {
	o := new(OpenAi)
	o.conf = conf
	o.db = db
	return o, nil
}

func (o *OpenAi) Chat(model string, messages []dto.OpenAiMessage) (string, error) {
	data, err := json.Marshal(messages)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", util.OpenAiEndPoint, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.conf.OpenAiApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
