package dal

import (
	"github.com/zenpk/chatbone/util"
	"go.mongodb.org/mongo-driver/mongo"
)

type Message struct {
	Id        string // timestamp + userId
	Deleted   bool
	UserId    int64
	Timestamp int64
	Messages  string
	Model     string
	Persona   string
	Shared    bool

	conf           *util.Configuration
	client         *mongo.Client
	collectionName string
}

func (m *Message) Init(conf *util.Configuration, client *mongo.Client) {
	m.conf = conf
	m.client = client
	m.collectionName = "message"
}

func (m *Message) SelectById(id string) (*Message, error) {
	m.client.Database(m.conf.MongoDbName).Collection(m.collectionName)
	return nil, nil
}

func (m *Message) SelectByUserId(id int64) ([]*Message, error) {
	return nil, nil
}
