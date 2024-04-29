package dal

import (
	"errors"

	"github.com/zenpk/chatbone/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Message struct {
	Deleted   bool
	SessionId string
	UserId    string // uuid
	Timestamp int64
	Messages  string // json string of messages, might include persona (role: system)
	ModelId   int
	Shared    bool
	Saved     bool // if false, it means the message is automatically saved (last)

	conf           *util.Configuration
	logger         util.ILogger
	client         *mongo.Client
	collectionName string
	err            error
}

func newMessage(conf *util.Configuration, client *mongo.Client, logger util.ILogger) (*Message, error) {
	m := new(Message)
	m.conf = conf
	m.logger = logger
	m.client = client
	m.collectionName = "message"
	m.err = errors.New("at Message table")
	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
	defer cancel()
	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	mod := mongo.IndexModel{
		Keys: bson.M{"userId": "hashed"},
	}
	_, err := collection.Indexes().CreateOne(ctx, mod)
	if err != nil {
		return nil, errors.Join(err, m.err)
	}
	mod = mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}
	_, err = collection.Indexes().CreateOne(ctx, mod)
	if err != nil {
		return nil, errors.Join(err, m.err)
	}
	mod = mongo.IndexModel{
		Keys: bson.M{"sessionId": "hashed"},
	}
	_, err = collection.Indexes().CreateOne(ctx, mod)
	if err != nil {
		return nil, errors.Join(err, m.err)
	}
	return m, nil
}

func (m *Message) SelectBySessionId(id string) (*Message, error) {
	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	filter := bson.M{"deleted": false, "sessionId": id}
	result := new(Message)
	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
	defer cancel()
	if err := collection.FindOne(ctx, filter).Decode(result); err != nil {
		return nil, errors.Join(err, m.err)
	}
	return result, nil
}

func (m *Message) SelectByUserId(userId string) ([]*Message, error) {
	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	filter := bson.M{"deleted": false, "userId": userId}
	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
	defer cancel()
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.Join(err, m.err)
	}
	result := make([]*Message, 0)
	if err := cursor.All(ctx, &result); err != nil {
		return nil, errors.Join(err, m.err)
	}
	return result, nil
}

func (m *Message) SelectLastByUserId(userId string) (*Message, error) {
	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	filter := bson.M{"deleted": false, "userId": userId, "saved": false}
	result := new(Message)
	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
	defer cancel()
	if err := collection.FindOne(ctx, filter).Decode(result); err != nil {
		return nil, errors.Join(err, m.err)
	}
	return result, nil
}

func (m *Message) Insert(message *Message) error {
	if err := m.checkInput(message); err != nil {
		return err
	}
	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
	defer cancel()
	_, err := collection.InsertOne(ctx, message)
	return errors.Join(err, m.err)
}

func (m *Message) ReplaceBySessionId(message *Message) error {
	if err := m.checkInput(message); err != nil {
		return err
	}
	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	filter := bson.M{"deleted": false, "sessionId": message.SessionId}
	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
	defer cancel()
	_, err := collection.ReplaceOne(ctx, filter, message)
	return errors.Join(err, m.err)
}

func (m *Message) checkInput(message *Message) error {
	if message == nil || message.UserId == "" || message.SessionId == "" || message.Messages == "" ||
		message.Timestamp <= 0 || message.ModelId <= 0 {
		return errors.Join(errors.New("insert invalid input"), m.err)
	}
	return nil
}
