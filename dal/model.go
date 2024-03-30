package dal

import (
	"github.com/zenpk/chatbone/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Model struct {
	Deleted      bool
	Name         string
	Provider     string  // e.g. openai
	InRate       float64 // dollar per token
	OutRate      float64
	SupportImage bool

	conf           *util.Configuration
	client         *mongo.Client
	collectionName string
}

func initModel(conf *util.Configuration, client *mongo.Client) (*Model, error) {
	m := new(Model)
	m.conf = conf
	m.client = client
	m.collectionName = "message"
	return m, nil
}

func (m *Model) SelectByProviderAndName(provider, name string) (*Model, error) {
	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	filter := bson.M{"Deleted": false, "Provider": provider, "Name": name}
	result := new(Model)
	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
	defer cancel()
	if err := collection.FindOne(ctx, filter).Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *Model) SelectAll() ([]*Model, error) {
	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	filter := bson.M{"Deleted": false}
	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
	defer cancel()
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	result := make([]*Model, 0)
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}
