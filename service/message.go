package service

import (
	"github.com/zenpk/chatbone/dal"
	"github.com/zenpk/chatbone/util"
)

type Message struct {
	conf   *util.Configuration
	logger util.ILogger
	db     *dal.Database
}

func NewMessage(conf *util.Configuration, logger util.ILogger, db *dal.Database) (*Message, error) {
	m := new(Message)
	m.conf = conf
	m.logger = logger
	m.db = db
	return m, nil
}

func (m *Message) GetMessages(userId int64) error {
	return nil
}
