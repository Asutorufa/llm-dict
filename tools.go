package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Asutorufa/hujiang_dictionary/en"
	"github.com/Asutorufa/hujiang_dictionary/jp"
)

type Tool interface {
	Call(ctx context.Context, input string) (string, error)
	Next() []Tool
}

type printExplainTool struct{}

type printExplainRequest struct {
	FullExplanation  string `json:"full_explanation,omitempty"`
	WordsExplanation []Word `json:"words_explanation,omitempty"`
}

func (p printExplainRequest) String() string {
	var str strings.Builder

	fmt.Fprint(&str, p.FullExplanation)
	str.WriteString("\n\nWords:\n")

	for _, v := range p.WordsExplanation {
		fmt.Fprintf(&str, "%s %s\n", v.Word, v.PartOfSpeech)
		fmt.Fprintf(&str, "  %s\n", v.Explanation)
	}

	return str.String()
}

type Word struct {
	Explanation  string `json:"explanation,omitempty"`
	PartOfSpeech string `json:"part_of_speech,omitempty"`
	Word         string `json:"word,omitempty"`
}

func (t *printExplainTool) Call(ctx context.Context, input string) (string, error) {
	var calls printExplainRequest
	_ = json.Unmarshal([]byte(input), &calls)

	fmt.Println(calls.String())
	return "The full explanation of original text already printed.", nil
}

func (t *printExplainTool) Next() []Tool {
	return []Tool{}
}

type searchWordTool struct{}

type searchWordRequest struct {
	Language string `json:"language,omitempty"`
	Word     string `json:"word,omitempty"`
}

func (t *searchWordTool) Call(ctx context.Context, input string) (string, error) {
	var req searchWordRequest
	_ = json.Unmarshal([]byte(input), &req)

	slog.Info("search word", "language", req.Language, "word", req.Word)

	switch req.Language {
	case "ja":
		return jp.FormatString(req.Word), nil
	case "en":
		return en.FormatString(req.Word), nil
	default:
		return fmt.Sprintf("unsupported language: %s", req.Language), nil
	}
}

func (t *searchWordTool) Next() []Tool {
	return []Tool{t}
}
