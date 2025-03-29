package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
)

type LLM struct {
	client *openai.Client
	model  string
}

func NewLLM(baseurl string, apiKeys string, model string) *LLM {
	client := openai.NewClient(
		option.WithAPIKey(apiKeys),
		option.WithBaseURL(baseurl),
	)

	return &LLM{
		client: client,
		model:  model,
	}
}

func (llm *LLM) Chat(ctx context.Context, prompt string, w io.Writer) error {
	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(SystemZh),
		openai.UserMessage(prompt),
	}

	for {
		fmt.Println("--------- history ---------")
		json.NewEncoder(w).Encode(msgs)
		fmt.Println("--------- history ---------")

		ss := llm.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Messages: openai.F(msgs),
			Model:    openai.F(llm.model),
		})

		tp := toolsParser{
			tools: map[string][]string{},
		}

		for ss.Next() {
			_, _ = w.Write([]byte(tp.parseTools(ss.Current().Choices[0].Delta.Content)))
		}

		if len(tp.tools) == 0 {
			return nil
		}

		for k, v := range tp.tools {
			switch k {
			case "searchWord":
				searchWordTool := searchWordTool{}

				for _, one := range v {
					resp, err := searchWordTool.Call(ctx, one)
					if err != nil {
						resp = err.Error()
					}

					msgs = append(msgs, openai.ToolMessage("searchWord", one))
					msgs = append(msgs, openai.FunctionMessage("searchWord", resp))
				}

			case "printExplain":
				printExplainTool := printExplainTool{}

				for _, one := range v {
					_, _ = printExplainTool.Call(ctx, one)
				}

				return nil
			}
		}
	}
}

func init() {
	ssestream.RegisterDecoder("text/event-stream", func(rc io.ReadCloser) ssestream.Decoder {
		return &eventStreamDecoder{
			rc:  rc,
			scn: bufio.NewScanner(rc),
		}
	})
}

// A base implementation of a Decoder for text/event-stream.
type eventStreamDecoder struct {
	evt ssestream.Event
	rc  io.ReadCloser
	scn *bufio.Scanner
	err error
}

func (s *eventStreamDecoder) Next() bool {
	if s.err != nil {
		return false
	}

	event := ""
	data := bytes.NewBuffer(nil)

	for s.scn.Scan() {
		txt := s.scn.Bytes()

		// Dispatch event on an empty line
		if len(txt) == 0 {
			if len(data.Bytes()) == 0 {
				continue
			}

			s.evt = ssestream.Event{
				Type: event,
				Data: data.Bytes(),
			}

			return true
		}

		// Split a string like "event: bar" into name="event" and value=" bar".
		name, value, _ := bytes.Cut(txt, []byte(":"))

		// Consume an optional space after the colon if it exists.
		if len(value) > 0 && value[0] == ' ' {
			value = value[1:]
		}

		switch string(name) {
		case "":
			// An empty line in the for ": something" is a comment and should be ignored.
			continue
		case "event":
			event = string(value)
		case "data":
			_, s.err = data.Write(value)
			if s.err != nil {
				break
			}
			_, s.err = data.WriteRune('\n')
			if s.err != nil {
				break
			}
		}
	}

	if s.scn.Err() != nil {
		s.err = s.scn.Err()
	}

	return false
}

func (s *eventStreamDecoder) Event() ssestream.Event {
	return s.evt
}

func (s *eventStreamDecoder) Close() error {
	return s.rc.Close()
}

func (s *eventStreamDecoder) Err() error {
	return s.err
}
