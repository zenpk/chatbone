package service

import (
	"github.com/zenpk/chatbone/dal"
	"github.com/zenpk/chatbone/util"
)

type OpenAi struct{}

func InitOpenAi(conf *util.Configuration, db *dal.Database) (*OpenAi, error) {
	return nil, nil
}
