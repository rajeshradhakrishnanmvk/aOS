package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type OllamaClient struct {
	baseURL    string
	model      string
	timeout    time.Duration
	http       *http.Client
	retries    int
	backoff    time.Duration
	maxBackoff time.Duration
}

func NewOllamaClient(baseURL, model string, timeoutSec int) *OllamaClient {
	// sensible defaults for retries/backoff
	const (
		defaultRetries    = 3
		defaultBackoffMs  = 500
		defaultMaxBackoff = 5000
	)

	// seed rand for jitter
	rand.Seed(time.Now().UnixNano())

	// Use no global HTTP client timeout so streaming responses can be consumed.
	// Per-request timeouts are applied with request contexts.
	return &OllamaClient{
		baseURL:    baseURL,
		model:      model,
		timeout:    time.Duration(timeoutSec) * time.Second,
		http:       &http.Client{Timeout: 0},
		retries:    defaultRetries,
		backoff:    time.Duration(defaultBackoffMs) * time.Millisecond,
		maxBackoff: time.Duration(defaultMaxBackoff) * time.Millisecond,
	}
}

// StreamChat sends the prompts with `stream: true` and calls `onChunk` for
// each partial chunk of content received. The provided ctx governs the
// overall request lifetime (use context.WithTimeout to bound it).
func (c *OllamaClient) StreamChat(ctx context.Context, systemPrompt, userPrompt string, onData func(string), onDone func()) error {
	reqBody := chatRequest{
		Model: c.model,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Stream: true,
	}

	resp, err := c.post("/api/chat", reqBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		var chunk chatResponse
		if err := decoder.Decode(&chunk); err != nil {
			return err
		}
		if chunk.Message.Content != "" {
			onData(chunk.Message.Content)
		}
		if chunk.Done {
			if onDone != nil {
				onDone()
			}
			break
		}
	}
	return nil
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Message message `json:"message"`
	Done    bool    `json:"done,omitempty"`
}

var ErrRequestFailed = errors.New("request failed")

// internal helper: marshal request and POST to path. Returns response on 200.
func (c *OllamaClient) post(path string, request interface{}) (*http.Response, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s%s", c.baseURL, path)
	resp, err := c.http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("%w: %s", ErrRequestFailed, resp.Status)
	}
	return resp, nil
}

func (c *OllamaClient) Chat(systemPrompt, userPrompt string) (string, error) {
	reqBody := chatRequest{
		Model: c.model,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Stream: false,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + "/api/chat"

	var lastErr error
	for attempt := 1; attempt <= c.retries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
		if err != nil {
			cancel()
			return "", fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.http.Do(req)
		cancel()
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: failed to call ollama: %w", attempt, err)
		} else {
			// got response, inspect
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				lastErr = fmt.Errorf("ollama returned HTTP %d: %s", resp.StatusCode, string(body))
			} else {
				var chatResp chatResponse
				if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
					lastErr = fmt.Errorf("failed to decode response: %w", err)
				} else {
					return chatResp.Message.Content, nil
				}
			}
		}

		// if not last attempt, sleep with backoff+jitter
		if attempt < c.retries {
			backoff := c.backoff * (1 << uint(attempt-1))
			if backoff > c.maxBackoff {
				backoff = c.maxBackoff
			}
			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			time.Sleep(backoff + jitter)
			continue
		}
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("failed to call ollama: unknown error")
}

// SetRetries sets the number of retry attempts for the client.
func (c *OllamaClient) SetRetries(n int) {
	if n > 0 {
		c.retries = n
	}
}

// SetBackoff sets the base backoff duration in milliseconds.
func (c *OllamaClient) SetBackoff(ms int) {
	if ms > 0 {
		c.backoff = time.Duration(ms) * time.Millisecond
	}
}

// SetMaxBackoff sets the maximum backoff duration in milliseconds.
func (c *OllamaClient) SetMaxBackoff(ms int) {
	if ms > 0 {
		c.maxBackoff = time.Duration(ms) * time.Millisecond
	}
}
