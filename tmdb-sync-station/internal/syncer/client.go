package syncer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxResponseBytes = 16 << 20

type APIError struct {
	Status   int
	Message  string
	NotFound bool
}

func (e *APIError) Error() string { return fmt.Sprintf("TMDB API 返回 %d: %s", e.Status, e.Message) }

type client struct {
	config      Config
	httpClient  *http.Client
	mu          sync.Mutex
	nextRequest time.Time
	interval    time.Duration
}

func newClient(config Config) *client {
	return &client{config: config, httpClient: &http.Client{Timeout: 90 * time.Second}, interval: time.Duration(float64(time.Second) / config.QPS)}
}

func (c *client) get(ctx context.Context, path string, query url.Values, destination any) error {
	for attempt := 0; attempt < 6; attempt++ {
		if err := c.wait(ctx); err != nil {
			return err
		}
		requestURL := strings.TrimRight(c.config.APIBase, "/") + "/" + strings.TrimLeft(path, "/")
		parameters := url.Values{}
		for key, values := range query {
			for _, value := range values {
				parameters.Add(key, value)
			}
		}
		if c.config.APIKey != "" {
			parameters.Set("api_key", c.config.APIKey)
		}
		if encoded := parameters.Encode(); encoded != "" {
			requestURL += "?" + encoded
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return err
		}
		request.Header.Set("Accept", "application/json")
		request.Header.Set("User-Agent", "TMDB-Sync-Station/1.0")
		if c.config.ReadAccessToken != "" {
			request.Header.Set("Authorization", "Bearer "+c.config.ReadAccessToken)
		}
		response, err := c.httpClient.Do(request)
		if err != nil {
			if attempt == 5 || !isTemporary(err) {
				return fmt.Errorf("请求 TMDB 失败: %w", err)
			}
			if err := pause(ctx, retryDelay(attempt, "")); err != nil {
				return err
			}
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
		response.Body.Close()
		if readErr != nil {
			if ctx.Err() == nil && attempt < 5 && isRetryableResponseRead(readErr) {
				if err := pause(ctx, retryDelay(attempt, "")); err != nil {
					return err
				}
				continue
			}
			return fmt.Errorf("读取 TMDB 响应失败: %w", readErr)
		}
		if len(body) > maxResponseBytes {
			return errors.New("TMDB 响应超过 16 MiB，已停止读取")
		}
		if response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500 {
			if attempt == 5 {
				return responseError(response.StatusCode, body)
			}
			if err := pause(ctx, retryDelay(attempt, response.Header.Get("Retry-After"))); err != nil {
				return err
			}
			continue
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return responseError(response.StatusCode, body)
		}
		if err := json.Unmarshal(body, destination); err != nil {
			return fmt.Errorf("解析 TMDB JSON 失败: %w", err)
		}
		return nil
	}
	return errors.New("TMDB 请求重试次数已用尽")
}

func (c *client) wait(ctx context.Context) error {
	c.mu.Lock()
	wait := time.Until(c.nextRequest)
	if wait < 0 {
		wait = 0
	}
	c.nextRequest = time.Now().Add(wait + c.interval)
	c.mu.Unlock()
	return pause(ctx, wait)
}

func retryDelay(attempt int, retryAfter string) time.Duration {
	if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
		return min(time.Duration(seconds)*time.Second, time.Minute)
	}
	if date, err := http.ParseTime(retryAfter); err == nil && time.Until(date) > 0 {
		return min(time.Until(date), time.Minute)
	}
	return min(time.Duration(1<<attempt)*time.Second, 30*time.Second)
}

func responseError(status int, body []byte) error {
	var message struct {
		StatusMessage string `json:"status_message"`
		StatusCode    int    `json:"status_code"`
	}
	_ = json.Unmarshal(body, &message)
	if message.StatusMessage == "" {
		message.StatusMessage = strings.TrimSpace(string(body))
	}
	if len(message.StatusMessage) > 500 {
		message.StatusMessage = message.StatusMessage[:500] + "…"
	}
	if message.StatusMessage == "" {
		message.StatusMessage = http.StatusText(status)
	}
	return &APIError{Status: status, Message: message.StatusMessage, NotFound: status == http.StatusNotFound && message.StatusCode == 34}
}

func isTemporary(err error) bool {
	if errors.Is(err, context.Canceled) {
		return false
	}
	var urlError *url.Error
	if errors.As(err, &urlError) {
		return true
	}
	var networkError net.Error
	return errors.As(err, &networkError) && (networkError.Timeout() || networkError.Temporary())
}

func isRetryableResponseRead(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || isTemporary(err)
}

func pause(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
