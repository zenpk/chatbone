package dal

// import (
// 	"github.com/zenpk/chatbone/util"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// )

// type ModelUnused struct {
// 	Deleted      bool
// 	Name         string
// 	ProviderId   int     // e.g. openai: 1
// 	InRate       float64 // dollar per token
// 	OutRate      float64
// 	SupportImage bool

// 	conf           *util.Configuration
// 	logger         util.ILogger
// 	client         *mongo.Client
// 	collectionName string
// }

// func initModelUnused(conf *util.Configuration, client *mongo.Client, logger util.ILogger) (*ModelUnused, error) {
// 	m := new(ModelUnused)
// 	m.conf = conf
// 	m.logger = logger
// 	m.client = client
// 	m.collectionName = "model"
// 	return m, nil
// }

// func (m *ModelUnused) SelectByProviderAndName(provider int, name string) (*ModelUnused, error) {
// 	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
// 	filter := bson.M{"Deleted": false, "Provider": provider, "Name": name}
// 	result := new(ModelUnused)
// 	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
// 	defer cancel()
// 	if err := collection.FindOne(ctx, filter).Decode(result); err != nil {
// 		return nil, err
// 	}
// 	return result, nil
// }

// func (m *ModelUnused) SelectAll() ([]*ModelUnused, error) {
// 	collection := m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
// 	filter := bson.M{"Deleted": false}
// 	ctx, cancel := util.GetTimeoutContext(m.conf.TimeoutSecond)
// 	defer cancel()
// 	cursor, err := collection.Find(ctx, filter)
// 	if err != nil {
// 		return nil, err
// 	}
// 	result := make([]*ModelUnused, 0)
// 	if err := cursor.All(ctx, &result); err != nil {
// 		return nil, err
// 	}
// 	return result, nil
// }
