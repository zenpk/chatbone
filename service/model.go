package service

import (
	"errors"

	"github.com/zenpk/chatbone/dal"
	"github.com/zenpk/chatbone/util"
)

type Model struct {
	conf   *util.Configuration
	logger util.ILogger
	err    error

	model *dal.Model
}

func NewModel(conf *util.Configuration, logger util.ILogger, db *dal.Database) (*Model, error) {
	m := new(Model)
	m.conf = conf
	m.logger = logger
	m.model = db.Model
	m.err = errors.New("at Model service")
	return m, nil
}

func (m *Model) GetAndCheckModelById(id int) (*dal.Model, error) {
	model, err := m.model.SelectById(id)
	if err != nil {
		return nil, errors.Join(err, m.err)
	}
	if model == nil {
		return nil, errors.Join(errors.New("model not found"), m.err)
	}
	return model, nil
}
