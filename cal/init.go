package cal

import (
	"github.com/zenpk/chatbone/dal"
	"github.com/zenpk/chatbone/util"
)

type Cache struct {
	User *User
}

func New(conf *util.Configuration, logger util.ILogger, db *dal.Database) (*Cache, error) {
	user, err := newUser(conf, logger, db)
	if err != nil {
		return nil, err
	}
	return &Cache{User: user}, nil
}
