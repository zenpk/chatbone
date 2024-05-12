package service

import (
	"errors"

	"github.com/zenpk/chatbone/cal"
	"github.com/zenpk/chatbone/dal"
	"github.com/zenpk/chatbone/util"
)

type User struct {
	conf   *util.Configuration
	logger util.ILogger
	err    error

	model   *dal.Model
	history *dal.History
	user    dal.IUser
}

func NewUser(conf *util.Configuration, logger util.ILogger, db *dal.Database, cache *cal.Cache) (*User, error) {
	u := new(User)
	u.conf = conf
	u.logger = logger
	u.model = db.Model
	u.history = db.History
	u.user = cache.User
	u.err = errors.New("at User service")
	return u, nil
}

func (u *User) GetInfo(uuid string) (*dal.User, error) {
	return u.user.SelectByIdInsertIfNotExists(uuid)
}
