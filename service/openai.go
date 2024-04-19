package service

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/zenpk/chatbone/dal"
	"github.com/zenpk/chatbone/dto"
	"github.com/zenpk/chatbone/util"
)

type OpenAi struct {
	conf   *util.Configuration
	logger util.ILogger
	db     *dal.Database
}

func InitOpenAi(conf *util.Configuration, logger util.ILogger, db *dal.Database) (*OpenAi, error) {
	o := new(OpenAi)
	o.conf = conf
	o.logger = logger
	o.db = db
	return o, nil
}

func (o *OpenAi) Chat(model string, messages []dto.OpenAiMessage) error {
	data, err := json.Marshal(messages)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", util.OpenAiEndPoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.conf.OpenAiApiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	temp := make([]byte, 4096)

	for {
		n, err := resp.Body.Read(temp)
		if n > 0 {
			// append read bytes to the buffer
			buf.Write(temp[:n])

			for {
				// try to decode the object
				dec := json.NewDecoder(&buf)
				if err := dec.Decode(&t); err != nil {
					if err == io.EOF {
						// not enough data to decode, break the inner loop
						break
					}
					// handle other errors (malformed JSON, etc.)
					return err
				}

				// send the decoded object to the channel

				// remove the decoded object from the buffer
				next := make([]byte, 4096)
				_, _ = dec.Buffered().Read(next)
				buf = *bytes.NewBuffer(next)
			}
		}

		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
	}

	return nil
}
