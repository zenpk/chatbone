package service

import (
	"github.com/zenpk/chatbone/dal"
	"github.com/zenpk/chatbone/util"
)

type Message struct{}

func InitMessage(conf *util.Configuration, db *dal.Database) (*Message, error) {
	return nil, nil
}

func (m *Message) GetMessages(userId int64) error {
	return nil
}
