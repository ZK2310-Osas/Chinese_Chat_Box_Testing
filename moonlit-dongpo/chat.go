package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

type chatRequest struct {
	Messages []message `json:"messages"`
}
type chatResponse struct {
	Reply    string `json:"reply"`
	Provider string `json:"provider,omitempty"`
}

func respond(w http.ResponseWriter, status int, reply, provider string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(chatResponse{Reply: reply, Provider: provider})
}
func chatHandler(client *modelClient, l *limiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Session-Id")
		if len(id) < 16 || len(id) > 80 {
			respond(w, 400, "会话标识无效，请刷新页面重试。", "")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 32*1024)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var payload chatRequest
		if err := decoder.Decode(&payload); err != nil {
			respond(w, 400, "请求格式有误，请重新发送。", "")
			return
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			respond(w, 400, "请求内容无效。", "")
			return
		}
		if n := len(payload.Messages); n < 1 || n > 12 || payload.Messages[n-1].Role != "user" {
			respond(w, 400, "对话记录格式不正确。", "")
			return
		}
		total := 0
		for _, m := range payload.Messages {
			if m.Role != "user" && m.Role != "assistant" || !utf8.ValidString(m.Content) || strings.TrimSpace(m.Content) == "" {
				respond(w, 400, "对话记录包含无效消息。", "")
				return
			}
			size := utf8.RuneCountInString(m.Content)
			total += size
			if m.Role == "user" && size > 300 {
				respond(w, 400, "每条消息最多 300 字。", "")
				return
			}
		}
		if total > 4000 {
			respond(w, 400, "对话记录过长，请刷新页面重新开始。", "")
			return
		}
		pacing, ok := l.start(id)
		if !ok {
			respond(w, 429, "上一条消息仍在处理，请稍候再试。", "")
			return
		}
		defer l.finish(id)
		ctx, cancel := context.WithTimeout(r.Context(), 40*time.Second)
		defer cancel()
		begin := time.Now()
		if pacing > 0 {
			select {
			case <-time.After(pacing):
			case <-ctx.Done():
				respond(w, 503, "今夜月色正浓，请稍后再问。", "")
				return
			}
		}
		if !l.acquire(ctx) {
			respond(w, 503, "访客较多，请稍后再试。", "")
			return
		}
		reply, err := client.reply(ctx, payload.Messages)
		l.release()
		if err != nil {
			log.Printf("Gemini request failed: %v", err)
			respond(w, 503, "月下雅集暂时繁忙，请稍后再来。请联系网站管理员检查服务器日志。", "")
			return
		}
		remaining := 5*time.Second - time.Since(begin)
		if remaining > 0 {
			select {
			case <-time.After(remaining):
			case <-ctx.Done():
				respond(w, 503, "请求超时，请稍后再试。", "")
				return
			}
		}
		respond(w, 200, reply, "gemini")
	})
}
