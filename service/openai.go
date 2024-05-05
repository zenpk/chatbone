package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/pkoukk/tiktoken-go"
	"github.com/zenpk/chatbone/cal"
	"github.com/zenpk/chatbone/dal"
	"github.com/zenpk/chatbone/dto"
	"github.com/zenpk/chatbone/util"
)

type OpenAi struct {
	conf   *util.Configuration
	logger util.ILogger
	err    error

	model   *dal.Model
	history *dal.History
	user    dal.IUser
}

func NewOpenAi(conf *util.Configuration, logger util.ILogger, db *dal.Database, cache *cal.Cache) (*OpenAi, error) {
	o := new(OpenAi)
	o.conf = conf
	o.logger = logger
	o.model = db.Model
	o.history = db.History
	o.user = cache.User
	o.err = errors.New("at OpenAi service")
	return o, nil
}

func (o *OpenAi) Chat(uuid string, model *dal.Model, reqBody *dto.OpenAiReqFromClient, respChan chan<- any) error {
	if uuid == "" || reqBody == nil || respChan == nil {
		return errors.Join(errors.New("chat invalid input"), o.err)
	}
	user, err := o.user.SelectByIdInsertIfNotExists(uuid)
	if err != nil {
		return errors.Join(err, o.err)
	}
	if user.Balance <= 0 {
		return errors.Join(errors.New("user doesn't have enough balance"), o.err)
	}
	if err := o.checkChatRequestBody(reqBody); err != nil {
		return errors.Join(err, o.err)
	}
	reqByte, err := json.Marshal(dto.OpenAiReqToOpenAi{
		Model:    model.Name,
		Messages: reqBody.Messages,
		Stream:   true, // always stream
	})
	if err != nil {
		return errors.Join(err, o.err)
	}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqByte))
	if err != nil {
		return errors.Join(err, o.err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.conf.OpenAiApiKey)
	client := http.Client{
		Timeout: time.Duration(o.conf.TimeoutSecond) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Join(err, o.err)
	}
	defer resp.Body.Close()
	openAiChatter := newOpenAiChatter(8192, "data: ", dto.OpenAiMessageEnding)
	responseAny, nil := chat(openAiChatter, resp, respChan)
	responseMessages := make([]dto.OpenAiMessage, len(responseAny))
	for i, message := range responseAny {
		responseMessages[i] = message.(dto.OpenAiMessage)
	}
	// update the history
	inToken, err := o.countTokensFromMessages(reqBody.Messages, model)
	if err != nil {
		return errors.Join(err, o.err)
	}
	outToken, err := o.countTokensFromMessages(responseMessages, model)
	if err != nil {
		return errors.Join(err, o.err)
	}
	o.user.ReduceBalance(user.Id, int64(inToken)*int64(model.InRate*dal.BalanceMultipleFactor)+
		int64(outToken)*int64(model.OutRate*dal.BalanceMultipleFactor))
	if err := o.history.Insert(&dal.History{
		SessionId:     reqBody.SessionId,
		Timestamp:     util.GetTimestamp(),
		UserId:        user.Id,
		ModelId:       reqBody.ModelId,
		InTokenCount:  inToken,
		OutTokenCount: outToken,
	}); err != nil {
		return errors.Join(err, o.err)
	}
	return nil
}

func (o *OpenAi) countTokensFromMessages(messages []dto.OpenAiMessage, model *dal.Model) (int, error) {
	tke, err := tiktoken.GetEncoding(model.Encoding)
	if err != nil {
		return 0, err
	}

	tokensPerMessage := 0
	switch model.Name {
	case "gpt-3.5-turbo", "gpt-4-turbo":
		tokensPerMessage = 3
	default:
		o.logger.Warnf("countTokensFromMessage unknown model: %s", model)
	}

	numTokens := 0
	for _, message := range messages {
		numTokens += tokensPerMessage
		numTokens += len(tke.Encode(message.Content, nil, nil))
		numTokens += len(tke.Encode(message.Role, nil, nil))
	}
	numTokens += 3 // every reply is primed with <|start|>assistant<|message|>
	return numTokens, nil
}

func (o *OpenAi) checkChatRequestBody(req *dto.OpenAiReqFromClient) error {
	if req == nil {
		return errors.New("request body should not be nil")
	}
	// check messages
	messageLen := 0
	for _, message := range req.Messages {
		if message.Role != "user" && message.Role != "assistant" && message.Role != "system" {
			return errors.New("unsupported message role")
		}
		if message.Content == "" {
			return errors.New("message content should not be empty")
		}
		messageLen += len(message.Content)
		if messageLen > o.conf.MessageLengthLimit {
			return errors.New("message content too long")
		}
	}
	return nil
}
