package cal

import (
	"errors"

	"github.com/zenpk/chatbone/dal"
	"github.com/zenpk/chatbone/util"
)

type User struct {
	*dal.User

	conf   *util.Configuration
	logger util.ILogger
	cached map[string]*dal.User // id -> User
	err    error
}

func newUser(conf *util.Configuration, logger util.ILogger, db *dal.Database) (*User, error) {
	u := new(User)
	u.User = db.User
	u.conf = conf
	u.logger = logger
	u.err = errors.New("at User cache")
	cached, err := db.User.SelectAll()
	if err != nil {
		return nil, errors.Join(err, u.err)
	}
	for _, user := range cached {
		if user == nil || user.Id == "" {
			return nil, errors.Join(errors.New("found malformed User data"), u.err)
		}
		u.cached[user.Id] = user
	}
	return u, nil
}

func (u *User) SelectByIdInsertIfNotExists(uuid string) (*dal.User, error) {
	// cache first
	if user, ok := u.cached[uuid]; ok {
		return user, nil
	}
	user, err := u.User.SelectByIdInsertIfNotExists(uuid)
	if err != nil {
		return nil, err
	}
	if user == nil || user.Id == "" {
		return nil, errors.Join(errors.New("found malformed User data"), u.err)
	}
	u.cached[user.Id] = user
	return user, nil
}

func (u *User) ReduceBalance(id string, amount int64) error {
	// TODO
	user, err := u.SelectByIdInsertIfNotExists(id)
	if err != nil {
		return err
	}
	if user.Balance < amount {
		return errors.Join(errors.New("balance is not enough"), u.err)
	}
	user.Balance -= amount
	return u.User.ReduceBalance(id, amount)
}
