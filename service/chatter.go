package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/zenpk/chatbone/dto"
)

type Chatter interface {
	ReadBody(*http.Response) (int, error)
	CanProcess() bool
	BodyEmpty() bool
	ParseJson() (any, error)
}

// chat must follow the correct processing order
// which is read body -> check data validity -> parse json
func chat(chatter Chatter, resp *http.Response, respChan chan<- any) ([]any, error) {
	responseArr := make([]any, 0)
	for {
		n, err := chatter.ReadBody(resp)
		if err != nil {
			return nil, err
		}
		if !chatter.CanProcess() {
			continue
		}
		if n <= 0 && chatter.BodyEmpty() {
			return responseArr, nil
		}
		parsed, err := chatter.ParseJson()
		if err != nil {
			if errors.Is(err, ErrIncompleteJson) {
				continue
			}
			return nil, err
		}
		if parsed != nil {
			respChan <- parsed
			responseArr = append(responseArr, parsed)
		}
	}
}

type OpenAiChatter struct {
	buffer     []byte
	bufferPos  int
	bufferSize int
	prefix     string
	suffix     string
}

func newOpenAiChatter(bufferSize int, prefix, suffix string) *OpenAiChatter {
	o := new(OpenAiChatter)
	o.buffer = make([]byte, bufferSize)
	o.bufferPos = 0
	o.bufferSize = bufferSize
	o.prefix = prefix
	o.suffix = suffix
	return o
}

func (o *OpenAiChatter) ReadBody(resp *http.Response) (int, error) {
	n, err := resp.Body.Read(o.buffer[o.bufferPos:])
	if err != nil {
		if err != io.EOF {
			return 0, fmt.Errorf("read body to buffer failed: %w", err)
		}
	}
	o.bufferPos = n
	return n, nil
}

func (o *OpenAiChatter) CanProcess() bool {
	startPos := bytes.Index(o.buffer, []byte(o.prefix))
	return startPos != -1
}

func (o *OpenAiChatter) BodyEmpty() bool {
	return o.bufferPos == 0
}

func (o *OpenAiChatter) ParseJson() (any, error) {
	startPos := bytes.Index(o.buffer, []byte(o.prefix))
	if startPos == -1 {
		return nil, errors.New("parse JSON failed, data invalid")
	}
	buf := bytes.NewBuffer(o.buffer[startPos+len(o.prefix) : o.bufferPos])
	dec := json.NewDecoder(buf)
	var message dto.OpenAiResp
	if err := dec.Decode(&message); err != nil && err != io.EOF {
		// malformed json, this often means only a part of JSON was received
		// and can be concatenated with the next read
		// ignore this time and wait for the next read
		return nil, ErrIncompleteJson
	}

	// decoded successfully, clear the buffer
	// we only need to clear the prefix part of the buffer
	// there's no need to clear the JSON part of the consumed buffer
	// because the next read will start from the next prefix
	newBuffer := make([]byte, o.bufferSize)
	copy(newBuffer, o.buffer[startPos+len(o.prefix):])
	o.buffer = newBuffer
	o.bufferPos -= startPos + len(o.prefix)

	// check the message
	if len(message.Choices) > 0 && message.Choices[0] != nil &&
		message.Choices[0].Delta != nil && message.Choices[0].Delta.Content != "" {
		return message, nil
	}
	return nil, nil
}
