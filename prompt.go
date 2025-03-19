package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

var SystemZh = `
你是专业的翻译官，你需要根据用户的输出及要求，给出详细的解析。

## 你需要遵守以下规则

- 请使用用户所使用的语言作为目标语言，除非用户指定了目标语言。如果用户没有任何表达，那么默认目标语言为英文。
- 如果输入为句子，那么你需要解析句子的结构，并且尽可能多的解析句子中的词的定义。
- 如果用户指定了只需要解释句子中的某一部分，那么可以不遵守上一条规则。
- 如果用户输入的为单个词，那么只需要对这个词进行详细解析即可。
- 尽量注明翻译的来源，如某词典，并使用那些更具有权威的来源。
- 在可能的情况下，还需要使用原始语言来进行解释。
- 当你完成解释后，你需要调用printExplain来对解释进行打印。
- 你只需要输出printExplain的函数调用即可。

## 当你使用函数调用时，你需要以以下格式返回结果

<function-name>
</function-name>

如:

<printExplain>
{"full_explanation":"xxx","words_explanation":[{"word":"aaa","explanation":"a","part_of_speech":"名词"}]}
</printExplain>

## 你可以使用以下工具

` + jsonStr(printExplain)

func jsonStr(a any) string {
	data, _ := json.Marshal(a)

	return string(data)
}

type printExplainResponse struct {
	FullExplanation  string `json:"full_explanation,omitempty"`
	WordsExplanation []Word `json:"words_explanation,omitempty"`
}

func (p printExplainResponse) String() string {
	var str strings.Builder

	fmt.Fprint(&str, p.FullExplanation)
	str.WriteString("\n")

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

var printExplain = openai.ChatCompletionToolParam{
	Type: openai.F(openai.ChatCompletionToolTypeFunction),
	Function: openai.F(openai.FunctionDefinitionParam{
		Name:        openai.F("printExplain"),
		Description: openai.F("print words explains"),
		Parameters: openai.F(shared.FunctionParameters{
			"type": "object",
			"properties": shared.FunctionParameters{
				"full_explanation": map[string]any{
					"type":        "string",
					"description": "The full explanation of original text.",
				},
				"words_explanation": map[string]any{
					"type":        "array<object>",
					"description": "Detailed explanation of each word in the text.",
					"properties": map[string]any{
						"type":     "object",
						"required": []string{"word", "explanation"},
						"word": map[string]any{
							"type":        "string",
							"description": "Original word.",
						},
						"part_of_speech": map[string]any{
							"type":        "string",
							"description": "The word's part of speech.",
						},
						"explanation": map[string]any{
							"type":        "string",
							"description": "Detailed explanation of the word.",
						},
						"example": map[string]any{
							"type":        "array<string>",
							"Description": "example sentence",
						},
					},
				},
			},
			"required":             []string{"full_explanation"},
			"additionalProperties": false,
			"strict":               true,
		}),
	}),
}

type parseStatus int

var (
	needStartStart parseStatus = 0
	needStartEnd   parseStatus = 1
	needEndStart   parseStatus = 2
	needEndEnd     parseStatus = 3
)

type toolsParser struct {
	status  parseStatus
	current string

	body  string
	tools map[string]string
}

func (t *toolsParser) parseTools(b string) string {
	z := ""
	for _, v := range b {
		if t.status == needStartStart {
			if v == '<' {
				t.status = needStartEnd
				continue
			}
		}

		if t.status == needStartEnd {
			if v == '>' {
				t.current = t.body
				t.body = ""
				t.status = needEndStart
				continue
			}

			t.body += string(v)
			continue
		}

		if t.status == needEndStart {
			if v == '<' {
				t.status = needEndEnd
				continue
			}

			t.body += string(v)
			continue
		}

		if t.status == needEndEnd {
			if v == '>' {
				t.tools[t.current] = t.body
				t.status = needStartStart
			}

			continue
		}

		z += string(v)
	}

	return z
}
