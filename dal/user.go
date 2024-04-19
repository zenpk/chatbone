package dal

import (
	"errors"

	"github.com/zenpk/chatbone/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	Id            string // uuid
	LastTopUpTime int64
	Clipboard     string

	conf           *util.Configuration
	logger         util.ILogger
	client         *mongo.Client
	collectionName string
}

func initUser(conf *util.Configuration, client *mongo.Client, logger util.ILogger) (*User, error) {
	u := new(User)
	u.conf = conf
	u.logger = logger
	u.client = client
	u.collectionName = "user"
	ctx, cancel := util.GetTimeoutContext(u.conf.TimeoutSecond)
	defer cancel()
	collection := u.client.Database(u.conf.MongoDbName).Collection(u.collectionName)
	mod := mongo.IndexModel{
		Keys: bson.M{"Id": 1},
	}
	_, err := collection.Indexes().CreateOne(ctx, mod)
	return u, err
}

func (u *User) SelectByIdInsertIfNotExists(id string) (*User, error) {
	collection := u.client.Database(u.conf.MongoDbName).Collection(u.collectionName)
	filter := bson.M{"Deleted": false, "Id": id}
	result := new(User)
	ctx, cancel := util.GetTimeoutContext(u.conf.TimeoutSecond)
	defer cancel()
	if err := collection.FindOne(ctx, filter).Decode(result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			result.Id = id
			result.LastTopUpTime = -1
			result.Clipboard = ""
			if _, err := collection.InsertOne(ctx, result); err != nil {
				return nil, err
			}
			return result, nil
		}
		return nil, err
	}
	return result, nil
}

func (u *User) UpdateClipboard(id, clipboard string) error {
	collection := u.client.Database(u.conf.MongoDbName).Collection(u.collectionName)
	filter := bson.M{"Deleted": false, "Id": id}
	update := bson.M{"$set": bson.M{"Clipboard": clipboard}}
	ctx, cancel := util.GetTimeoutContext(u.conf.TimeoutSecond)
	defer cancel()
	if _, err := collection.UpdateOne(ctx, filter, update); err != nil {
		return err
	}
	return nil
}
