package service

import (
	"github.com/zenpk/chatbone/dal"
	"github.com/zenpk/chatbone/util"
)

type OpenAi struct {
	conf *util.Configuration
	db   *dal.Database
}

func InitOpenAi(conf *util.Configuration, db *dal.Database) (*OpenAi, error) {
	openAi := new(OpenAi)
	openAi.conf = conf
	openAi.db = db
	return openAi, nil
}
