package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// modelClient calls Gemini from the server; the browser never receives the API key.
type modelClient struct {
	key      string
	model    string
	http     *http.Client
	endpoint string
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func newGeminiClient(key, model string) *modelClient {
	return &modelClient{
		key:      strings.TrimSpace(key),
		model:    model,
		http:     &http.Client{Timeout: 36 * time.Second},
		endpoint: "https://generativelanguage.googleapis.com/v1beta/models/",
	}
}

func (c *modelClient) reply(ctx context.Context, history []message) (string, error) {
	if c.key == "" {
		return "", errors.New("GEMINI_API_KEY is not configured")
	}

	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Role  string `json:"role,omitempty"`
		Parts []part `json:"parts"`
	}
	contents := make([]content, 0, len(history))
	for _, m := range history {
		role := "user"
		if m.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, content{Role: role, Parts: []part{{Text: m.Content}}})
	}
	requestBody := struct {
		SystemInstruction content   `json:"systemInstruction"`
		Contents          []content `json:"contents"`
		GenerationConfig  struct {
			MaxOutputTokens int `json:"maxOutputTokens"`
		} `json:"generationConfig"`
	}{
		SystemInstruction: content{Parts: []part{{Text: personaPrompt}}},
		Contents:          contents,
	}
	requestBody.GenerationConfig.MaxOutputTokens = 2048
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	endpoint := c.endpoint + url.PathEscape(c.model) + ":generateContent"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.key)
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("Gemini network error: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiError struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Status  string `json:"status"`
			} `json:"error"`
		}
	
		if err := json.Unmarshal(data, &apiError); err == nil &&
			apiError.Error.Status != "" {
			return "", fmt.Errorf(
				"Gemini API HTTP %d (%s)",
				resp.StatusCode,
				apiError.Error.Status,
			)
		}
	
		return "", fmt.Errorf("Gemini API HTTP %d", resp.StatusCode)
	}
	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []part `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("invalid Gemini response: %w", err)
	}
	var texts []string
	for _, candidate := range parsed.Candidates {
		for _, p := range candidate.Content.Parts {
			if strings.TrimSpace(p.Text) != "" {
				texts = append(texts, p.Text)
			}
		}
		if len(texts) > 0 {
			break
		}
	}
	answer := strings.TrimSpace(strings.Join(texts, "\n"))
	if answer == "" {
		return "", errors.New("Gemini returned no text; the response may have been blocked")
	}
	return answer, nil
}
