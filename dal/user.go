package dal

import (
	"errors"
	"fmt"
	"sync"

	"github.com/zenpk/chatbone/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type IUser interface {
	SelectByIdInsertIfNotExists(id string) (*User, error)
	ReduceBalance(id string, amount int64) error
}

type User struct {
	Id        string // uuid
	Balance   int64  // dollar * multiple factor (1000000)
	Clipboard string

	conf           *util.Configuration
	logger         util.ILogger
	client         *mongo.Client
	mutex          *sync.Mutex // for charging the balance
	collectionName string
	err            error
}

func newUser(conf *util.Configuration, client *mongo.Client, logger util.ILogger) (*User, error) {
	u := new(User)
	u.conf = conf
	u.logger = logger
	u.client = client
	u.collectionName = "user"
	u.err = errors.New("at User table")
	u.mutex = new(sync.Mutex)
	ctx, cancel := util.GetTimeoutContext(u.conf.TimeoutSecond)
	defer cancel()
	collection := u.client.Database(u.conf.MongoDbName).Collection(u.collectionName)
	mod := mongo.IndexModel{
		Keys: bson.D{{"Id", 1}},
	}
	_, err := collection.Indexes().CreateOne(ctx, mod)
	if err != nil {
		return nil, errors.Join(err, u.err)
	}
	return u, nil
}

func (u *User) SelectByIdInsertIfNotExists(uuid string) (*User, error) {
	collection := u.client.Database(u.conf.MongoDbName).Collection(u.collectionName)
	filter := bson.M{"Id": uuid}
	result := new(User)
	ctx, cancel := util.GetTimeoutContext(u.conf.TimeoutSecond)
	defer cancel()
	// get mutex, this is for ensuring the user we get has the latest balance value
	u.mutex.Lock()
	defer u.mutex.Unlock()
	if err := collection.FindOne(ctx, filter).Decode(result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			result.Id = uuid
			result.Balance = 1 * BalanceMultipleFactor // 1 dollar
			result.Clipboard = ""
			if _, err := collection.InsertOne(ctx, result); err != nil {
				return nil, errors.Join(err, u.err)
			}
			return result, nil
		}
		return nil, errors.Join(err, u.err)
	}
	return result, nil
}

func (u *User) ReduceBalance(id string, amount int64) (*User, error) {
	if amount <= 0 {
		return nil, errors.Join(errors.New("balance reduce amount must be positive"), u.err)
	}
	// ensure the user exists
	user, err := u.SelectByIdInsertIfNotExists(id)
	if err != nil {
		return nil, fmt.Errorf("reduce balance: select user failed: %w", err)
	}
	collection := u.client.Database(u.conf.MongoDbName).Collection(u.collectionName)
	filter := bson.M{"Id": id}
	update := bson.M{"$set": bson.M{"Balance": user.Balance - amount}}
	ctx, cancel := util.GetTimeoutContext(u.conf.TimeoutSecond)
	defer cancel()
	// update with mutex
	u.mutex.Lock()
	defer u.mutex.Unlock()
	if _, err := collection.UpdateOne(ctx, filter, update); err != nil {
		return nil, errors.Join(err, u.err)
	}
	return user, nil
}

func (u *User) SelectAll() ([]*User, error) {
	collection := u.client.Database(u.conf.MongoDbName).Collection(u.collectionName)
	ctx, cancel := util.GetTimeoutContext(u.conf.TimeoutSecond)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Join(err, u.err)
	}
	results := make([]*User, 0)
	for cursor.Next(ctx) {
		var result User
		if err := cursor.Decode(&result); err != nil {
			return nil, errors.Join(err, u.err)
		}
		results = append(results, &result)
	}
	return results, nil
}

func (u *User) UpdateClipboard(id, clipboard string) error {
	collection := u.client.Database(u.conf.MongoDbName).Collection(u.collectionName)
	filter := bson.M{"Id": id}
	update := bson.M{"$set": bson.M{"Clipboard": clipboard}}
	ctx, cancel := util.GetTimeoutContext(u.conf.TimeoutSecond)
	defer cancel()
	if _, err := collection.UpdateOne(ctx, filter, update); err != nil {
		return errors.Join(err, u.err)
	}
	return nil
}
