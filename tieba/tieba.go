// Package tieba is the library behind the tieba command: the HTTP client,
// request shaping, and the typed data models for Baidu Tieba (百度贴吧).
//
// The client fetches hot topic data from the public Tieba mobile endpoint
// at https://tieba.baidu.com/hottopic/browse/topicList. No authentication is
// required. It sets a mobile User-Agent, paces requests, and retries transient
// 429/5xx errors with exponential backoff.
package tieba

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// DefaultUserAgent is the mobile User-Agent that the Tieba endpoint expects.
const DefaultUserAgent = "Mozilla/5.0 (compatible; MSIE 10.0; Windows Phone 8.0)"

// Host is the canonical site hostname.
const Host = "tieba.baidu.com"

// Config holds constructor parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns sensible defaults for talking to Baidu Tieba.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://tieba.baidu.com",
		UserAgent: DefaultUserAgent,
		Rate:      500 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
	}
}

// Client talks to the Baidu Tieba public API.
type Client struct {
	cfg        Config
	httpClient *http.Client
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

// Hot fetches the current hot topic list from Tieba. It returns up to limit
// topics; pass 0 for all topics returned by the API.
func (c *Client) Hot(ctx context.Context, limit int) ([]Topic, error) {
	url := c.cfg.BaseURL + "/hottopic/browse/topicList"
	raw, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}
	var resp apiResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode hot topics: %w", err)
	}
	list := resp.Data.BangTopic.TopicList
	if limit > 0 && limit < len(list) {
		list = list[:limit]
	}
	out := make([]Topic, 0, len(list))
	for i, t := range list {
		out = append(out, Topic{
			Rank:        i + 1,
			Name:        t.TopicName,
			Discussions: t.DiscussNum,
			URL:         t.TopicURL,
		})
	}
	return out, nil
}

// get performs an HTTP GET with retry/pace and returns the body bytes.
func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		b, retry, err := c.do(ctx, url)
		if err == nil {
			return b, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
