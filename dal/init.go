package dal

import (
	"context"

	"github.com/zenpk/chatbone/util"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	History *History
	Message *Message
	Model   *Model
	User    *User
}

func Init(conf *util.Configuration, logger util.ILogger) (*Database, error) {
	// use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(conf.MongoDbUri).SetServerAPIOptions(serverAPI)
	// create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return nil, err
	}
	history, err := initHistory(conf, client, logger)
	if err != nil {
		return nil, err
	}
	message, err := initMessage(conf, client, logger)
	if err != nil {
		return nil, err
	}
	// makes life easier to just hardcode the models
	model, err := initModel()
	if err != nil {
		return nil, err
	}
	user, err := initUser(conf, client, logger)
	if err != nil {
		return nil, err
	}
	return &Database{history, message, model, user}, nil
}
