package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGeminiRequest(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "test-key" {
			t.Errorf("missing key header")
		}
		if r.URL.Path != "/v1beta/models/gemini-2.5-flash-lite:generateContent" {
			t.Errorf("wrong path: %s", r.URL.Path)
		}
		var body struct {
			SystemInstruction struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"systemInstruction"`
			Contents []struct {
				Role string `json:"role"`
			} `json:"contents"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body.Contents) != 2 || body.Contents[0].Role != "model" || body.Contents[1].Role != "user" {
			t.Errorf("bad history mapping: %+v", body.Contents)
		}
		if len(body.SystemInstruction.Parts) == 0 {
			t.Error("persona not supplied")
		} else {
			persona := body.SystemInstruction.Parts[0].Text
			if !strings.Contains(persona, "你现在是宋代大文学家苏轼") || !strings.Contains(persona, "自称“老夫”“子瞻”或“东坡”") {
				t.Errorf("persona does not establish Su Shi's first-person role: %s", persona)
			}
			if !strings.Contains(persona, "都只用简体中文回答") {
				t.Errorf("persona does not require simplified Chinese: %s", persona)
			}
			for _, required := range []string{"密州遭遇蝗灾与严重饥荒", "欲把西湖比西子", "未来之事老夫尚未经历", "每次只选当前最相关的一点"} {
				if !strings.Contains(persona, required) {
					t.Errorf("persona knowledge base missing %q", required)
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"但愿人长久。"}]}}]}`))
	}))
	defer mock.Close()
	client := newGeminiClient("test-key", "gemini-2.5-flash-lite")
	client.endpoint = mock.URL + "/v1beta/models/"
	reply, err := client.reply(context.Background(), []message{{Role: "assistant", Content: "你好"}, {Role: "user", Content: "解释这首词"}})
	if err != nil || reply != "但愿人长久。" {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
}
