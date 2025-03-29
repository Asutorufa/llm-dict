package main

import "testing"

func TestParseTools(t *testing.T) {
	var data = `
<searchWord>
{"language": "ja", "word": "ナイフ"}
</searchWord>

<searchWord>
{"language": "ja", "word": "鉛筆"}
</searchWord>

<searchWord>
{"language": "ja", "word": "削る"}
</searchWord>

<printExplain>
{"full_explanation": "这句话的意思是用刀削铅笔。\n\n* ナイフ (naifu): 刀子，小刀 (noun)\n* 鉛筆 (enpitsu): 铅笔 (noun)\n* 削る (kezur): 削，刨，削尖 (verb)\n\n这句话的语法结构是：主语 (ナイフ) + 宾语 (鉛筆) + 谓语 (削る)。", "words_explanation": [{"word": "ナイフ", "explanation": "刀子，小刀。通常指用于切割的工具。", "part_of_speech": "名詞"}, {"word": "鉛筆", "explanation": "铅笔。用于书写或绘画的工具。", "part_of_speech": "名詞"}, {"word": "削る", "explanation": "削，刨，削尖。去除物体表面的一部分使其变薄或变尖。", "part_of_speech": "動詞"}]}
</printExplain>`

	tp := toolsParser{
		tools: map[string][]string{},
	}

	tp.parseTools(data)

	for k, v := range tp.tools {
		t.Log(k, v)
	}
}
