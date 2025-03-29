package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDict(t *testing.T) {
	t.Log(SystemZh)

	keys, err := os.ReadFile(".apikey")
	require.NoError(t, err)

	ll := NewLLM("https://openrouter.ai/api/v1/",
		string(keys),
		// "google/gemma-3-27b-it:free",
		// "google/gemini-2.0-pro-exp-02-05:free",
		// "google/gemini-2.0-flash-lite-preview-02-05:free",
		// "google/gemma-3-12b-it:free",
		// "qwen/qwq-32b:free",
		// "google/gemini-2.0-pro-exp-02-05:free",
		"deepseek/deepseek-r1:free",
		// "deepseek/deepseek-r1-distill-qwen-32b:free",
	)

	err = ll.Chat(t.Context(), `
飲み込みがはやい。

上面的是什么意思
`, os.Stdout)
	require.NoError(t, err)
}
