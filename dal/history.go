package dal

import (
	"github.com/zenpk/chatbone/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type History struct {
	Timestamp int64
	UserId    string
	Model     string
	Provider  string
	Cost      float64

	conf           *util.Configuration
	logger         util.ILogger
	client         *mongo.Client
	collectionName string
}

func initHistory(conf *util.Configuration, client *mongo.Client, logger util.ILogger) (*History, error) {
	h := new(History)
	h.conf = conf
	h.logger = logger
	h.client = client
	h.collectionName = "history"
	ctx, cancel := util.GetTimeoutContext(h.conf.TimeoutSecond)
	defer cancel()
	collection := h.client.Database(h.conf.MongoDbName).Collection(h.collectionName)
	mod := mongo.IndexModel{
		Keys: bson.D{{"Timestamp", 1}, {"UserId", 1}},
	}
	_, err := collection.Indexes().CreateOne(ctx, mod)
	return h, err
}

func (h *History) SelectByUserIdAfter(userId string, timestamp int64) ([]*History, error) {
	collection := h.client.Database(h.conf.MongoDbName).Collection(h.collectionName)
	filter := bson.M{"UserId": userId, "Timestamp": bson.M{"$gt": timestamp}}
	ctx, cancel := util.GetTimeoutContext(h.conf.TimeoutSecond)
	defer cancel()
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	result := make([]*History, 0)
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (h *History) Insert(history *History) error {
	collection := h.client.Database(h.conf.MongoDbName).Collection(h.collectionName)
	ctx, cancel := util.GetTimeoutContext(h.conf.TimeoutSecond)
	defer cancel()
	if _, err := collection.InsertOne(ctx, history); err != nil {
		return err
	}
	return nil
}
