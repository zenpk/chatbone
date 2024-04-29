package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

func (o *OpenAi) Chat(uuid string, model *dal.Model, reqBody *dto.OpenAiReqFromClient, responseChan chan<- string) error {
	if uuid == "" || reqBody == nil || responseChan == nil {
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
	const bufferSize = 8192
	bodyRead := make([]byte, bufferSize)
	// sometimes OpenAI will return an incomplete JSON, we save it to last
	last := make([]byte, bufferSize/4)
	lastLen := 0

	for {
		var buf bytes.Buffer
		n, err := resp.Body.Read(bodyRead)
		if err != nil {
			if err != io.EOF {
				return errors.Join(fmt.Errorf("read body to buffer failed: %w", err), o.err)
			}
			break
		}

		if n > 0 {
			// start from the prefix "data: " to remove the malformed JSON
			startPos := bytes.Index(bodyRead, []byte("data: "))
			if startPos == -1 {
				continue
			}
			dataToRead := bodyRead[startPos+6 : n]
			// check ending
			if find := bytes.Index(dataToRead, []byte(dto.MessageEnding)); find == 0 {
				responseChan <- dto.MessageEnding
				break
			}
			// if there is data from the last read, append the current data to it
			if lastLen > 0 {
				dataToRead = append(last[:lastLen], dataToRead...)
				lastLen = 0
			}

			// append read bytes to the buffer
			buf.Write(dataToRead)

			for {
				if buf.Len() <= 0 {
					break
				}
				// check ending
				if find := bytes.Index(buf.Bytes(), []byte(dto.MessageEnding)); find == 0 {
					responseChan <- dto.MessageEnding
					break
				}
				// try to decode the object
				var message dto.OpenAiResp
				dec := json.NewDecoder(&buf)
				if err := dec.Decode(&message); err != nil && err != io.EOF {
					o.logger.Warnf("decode message failed: %s, saving the buffer to last\n", err)
					// malformed json, this often means only a part of JSON was received
					// and can be concatenated with the next read
					// save the data to last and break
					decoderLen, err := dec.Buffered().Read(last)
					if err != nil && err != io.EOF {
						return errors.Join(fmt.Errorf("read remainder from decoder to last failed: %w", err), o.err)
					}
					bufferLen, err := buf.Read(last[decoderLen:])
					if err != nil && err != io.EOF {
						return errors.Join(fmt.Errorf("read remainder from buffer to last failed: %w", err), o.err)
					}
					lastLen = decoderLen + bufferLen
					break
				}
				// save the decoded object and send the message to the channel
				if len(message.Choices) > 0 && message.Choices[0] != nil &&
					message.Choices[0].Delta != nil && message.Choices[0].Delta.Content != "" {
					responseMessages = append(responseMessages, *message.Choices[0].Delta)
					responseChan <- message.Choices[0].Delta.Content
				}
				// read the remainder to next buffer
				next := make([]byte, bufferSize/4)
				decoderLen, err := dec.Buffered().Read(next)
				if err != nil && err != io.EOF {
					return errors.Join(fmt.Errorf("read remainder from decoder to next failed: %w", err), o.err)
				}
				bufferLen, err := buf.Read(next[decoderLen:])
				if err != nil && err != io.EOF {
					return errors.Join(fmt.Errorf("read remainder from buffer to next failed: %w", err), o.err)
				}
				// find the prefix
				startPos := bytes.Index(next, []byte("data: "))
				if startPos == -1 {
					// if there is no prefix, save the data to last and break
					// because at this point, the next might only contain "da" rather than a complete "data: "
					last = next
					lastLen = decoderLen + bufferLen
					break
				}
				buf = *bytes.NewBuffer(next[startPos+6 : decoderLen+bufferLen])
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
