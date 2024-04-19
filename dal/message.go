package dal

import (
	"github.com/zenpk/chatbone/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Message struct {
	Deleted   bool
	Id        string // UserId + Timestamp
	UserId    string // uuid
	Timestamp int64
	Messages  string
	Model     string
	Provider  string // e.g. openai
	Persona   string
	Shared    bool
	Last      bool // if is the automatically saved last message

	conf           *util.Configuration
	logger         util.ILogger
	client         *mongo.Client
	collectionName string
}

func initMessage(conf *util.Configuration, client *mongo.Client, logger util.ILogger) (*Message, error) {
	m := new(Message)
	m.conf = conf
	m.logger = logger
	m.client = client
	m.collectionName = "message"
	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
	defer cancel()
	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	mod := mongo.IndexModel{
		Keys: bson.M{"UserId": 1},
	}
	_, err := collection.Indexes().CreateOne(ctx, mod)
	if err != nil {
		return nil, err
	}
	mod = mongo.IndexModel{
		Keys: bson.M{"Id": 1},
	}
	_, err = collection.Indexes().CreateOne(ctx, mod)
	return m, err
}

func (m *Message) SelectById(id string) (*Message, error) {
	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	filter := bson.M{"Deleted": false, "Id": id}
	result := new(Message)
	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
	defer cancel()
	if err := collection.FindOne(ctx, filter).Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *Message) SelectByUserId(userId string) ([]*Message, error) {
	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	filter := bson.M{"Deleted": false, "UserId": userId}
	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
	defer cancel()
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	result := make([]*Message, 0)
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *Message) SelectLastByUserId(userId string) (*Message, error) {
	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	filter := bson.M{"Deleted": false, "UserId": userId, "Last": true}
	result := new(Message)
	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
	defer cancel()
	if err := collection.FindOne(ctx, filter).Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *Message) Insert(message *Message) error {
	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
	defer cancel()
	_, err := collection.InsertOne(ctx, message)
	return err
}

func (m *Message) ReplaceById(message *Message) error {
	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	filter := bson.M{"Deleted": false, "Id": message.Id}
	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
	defer cancel()
	_, err := collection.ReplaceOne(ctx, filter, message)
	return err
}
