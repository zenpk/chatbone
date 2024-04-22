package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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

func (o *OpenAi) Chat(sessionId string, user *dal.User, model *dal.Model, reqBody *dto.OpenAiReqFromClient, responseChan chan<- string) error {
	if sessionId == "" || user == nil || model == nil || reqBody == nil || responseChan == nil {
		return errors.Join(errors.New("chat invalid input"), o.err)
	}
	if err := o.checkRequestBody(reqBody); err != nil {
		return errors.Join(err, o.err)
	}
	reqByte, err := json.Marshal(dto.OpenAiReqToOpenAi{
		OpenAiReqFromClient: *reqBody,
		Stream:              true, // always stream
	})
	if err != nil {
		return errors.Join(err, o.err)
	}
	req, err := http.NewRequest("POST", util.OpenAiEndPoint, bytes.NewBuffer(reqByte))
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

	responseMessages := make([]dto.OpenAiMessage, 0)
	const bufferSize = 4096
	temp := make([]byte, bufferSize)
	lastLen := 0

	for {
		var buf bytes.Buffer
		n, err := resp.Body.Read(temp[lastLen:])
		if err != nil {
			if err != io.EOF {
				return errors.Join(err, o.err)
			}
			break
		}
		if n > 0 {
			// append read bytes to the buffer
			buf.Write(temp[:lastLen+n])
			lastLen = 0
			var message dto.OpenAiMessage

			for {
				// try to decode the object
				dec := json.NewDecoder(&buf)
				if err := dec.Decode(&message); err != nil {
					if err == io.EOF {
						// not enough data to decode, save the data to temp and break
						lastLen = buf.Len()
						if _, err := buf.Read(temp[:lastLen]); err != nil {
							return errors.Join(err, o.err)
						}
						break
					}
					// handle other errors (malformed JSON, etc.)
					return errors.Join(err, o.err)
				}
				// save the decoded object and send the message to the channel
				responseMessages = append(responseMessages, message)
				responseChan <- message.Content
				// remove the decoded object from the buffer
				next := make([]byte, bufferSize)
				_, _ = dec.Buffered().Read(next)
				buf = *bytes.NewBuffer(next)
			}
		}
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
	if err := o.history.Insert(&dal.History{
		SessionId:     sessionId,
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

func (o *OpenAi) checkRequestBody(req *dto.OpenAiReqFromClient) error {
	// check model
	model, err := o.model.SelectById(req.ModelId)
	if err != nil {
		return err
	}
	if model == nil {
		return errors.New("unsupported OpenAI model")
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
