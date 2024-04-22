package dal

import (
	"errors"

	"github.com/zenpk/chatbone/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type History struct {
	SessionId     string
	Timestamp     int64
	UserId        string
	ModelId       int
	InTokenCount  int
	OutTokenCount int

	conf           *util.Configuration
	logger         util.ILogger
	client         *mongo.Client
	collectionName string
	err            error
}

func newHistory(conf *util.Configuration, client *mongo.Client, logger util.ILogger) (*History, error) {
	h := new(History)
	h.conf = conf
	h.logger = logger
	h.client = client
	h.collectionName = "history"
	h.err = errors.New("at History table")
	ctx, cancel := util.GetTimeoutContext(h.conf.TimeoutSecond)
	defer cancel()
	collection := h.client.Database(h.conf.MongoDbName).Collection(h.collectionName)
	mod := mongo.IndexModel{
		Keys: bson.D{{"Timestamp", 1}, {"UserId", 1}},
	}
	_, err := collection.Indexes().CreateOne(ctx, mod)
	if err != nil {
		return nil, errors.Join(err, h.err)
	}
	return h, nil
}

func (h *History) SelectBySessionId(sessionId string) ([]*History, error) {
	collection := h.client.Database(h.conf.MongoDbName).Collection(h.collectionName)
	filter := bson.M{"SessionId": sessionId}
	ctx, cancel := util.GetTimeoutContext(h.conf.TimeoutSecond)
	defer cancel()
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, errors.Join(err, h.err)
	}
	result := make([]*History, 0)
	if err := cursor.All(ctx, &result); err != nil {
		return nil, errors.Join(err, h.err)
	}
	return result, nil
}

func (h *History) SelectByUserIdAfter(userId string, timestamp int64) ([]*History, error) {
	collection := h.client.Database(h.conf.MongoDbName).Collection(h.collectionName)
	filter := bson.M{"UserId": userId, "Timestamp": bson.M{"$gt": timestamp}}
	ctx, cancel := util.GetTimeoutContext(h.conf.TimeoutSecond)
	defer cancel()
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, errors.Join(err, h.err)
	}
	result := make([]*History, 0)
	if err := cursor.All(ctx, &result); err != nil {
		return nil, errors.Join(err, h.err)
	}
	return result, nil
}

func (h *History) Insert(history *History) error {
	if history == nil || history.SessionId == "" || history.UserId == "" || history.ModelId <= 0 ||
		history.Timestamp <= 0 || history.InTokenCount <= 0 || history.OutTokenCount <= 0 {
		return errors.Join(errors.New("insert invalid input"), h.err)
	}
	collection := h.client.Database(h.conf.MongoDbName).Collection(h.collectionName)
	ctx, cancel := util.GetTimeoutContext(h.conf.TimeoutSecond)
	defer cancel()
	if _, err := collection.InsertOne(ctx, history); err != nil {
		return errors.Join(err, h.err)
	}
	return nil
}
