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
- 因为不同语言之间对单词词性的定义不同，所以单词词性你必须使用原始语言。
- 你必须先调用searchWord来获取句子中每个单词的解释。
- 使用searchWord获取到的解析来完成内容的分析，并使用通俗易懂的解释来进行翻译。
- 你只能输出函数调用和解释，不能输出其他无关内容。

## 当你使用Tools时，你需要以以下格式进行调用

<function-name>
</function-name>

## 你可以使用以下工具

` + toolsString(searchWord)

func jsonStr(a any) string {
	data, _ := json.Marshal(a)

	return string(data)
}

func toolsString(tools ...openai.ChatCompletionToolParam) string {
	var str strings.Builder
	for _, v := range tools {
		fmt.Fprintf(&str, "<%s>\n", v.Function.Value.Name.Value)
		str.WriteString(jsonStr(v))
		fmt.Fprintf(&str, "\n</%s>\n", v.Function.Value.Name.Value)
	}

	return str.String()
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
					"type":        "array",
					"description": "Detailed explanation of each word in the text.",
					"item": map[string]any{
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

var searchWord = openai.ChatCompletionToolParam{
	Type: openai.F(openai.ChatCompletionToolTypeFunction),
	Function: openai.F(openai.FunctionDefinitionParam{
		Name:        openai.F("searchWord"),
		Description: openai.F("search one word explain"),
		Parameters: openai.F(shared.FunctionParameters{
			"type": "object",
			"properties": shared.FunctionParameters{
				"language": map[string]any{
					"type":        "string",
					"description": "The language of original text.",
					"enum":        []string{"en", "ja"},
				},
				"word": map[string]any{
					"type":        "string",
					"description": "Original word.",
				},
			},
			"required":             []string{"language", "word"},
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
	tools map[string][]string
}

func (t *toolsParser) parseTools(b string) string {
	// os.Stdout.WriteString(b)

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
				t.tools[t.current] = append(t.tools[t.current], t.body)
				t.status = needStartStart
				t.body = ""
			}

			continue
		}

		z += string(v)
	}

	return z
}
