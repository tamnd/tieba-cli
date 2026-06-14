package tieba_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tamnd/tieba-cli/tieba"
)

const mockHotResponse = `{
  "data": {
    "bang_topic": {
      "topic_list": [
        {
          "topic_name": "拉锯战!五星巴西战平摩洛哥",
          "discuss_num": 2656410,
          "topic_url": "https://tieba.baidu.com/hottopic/browse/hottopic?topic_id=28355141&topic_name=%E6%8B%89%E9%94%AF%E6%88%98"
        },
        {
          "topic_name": "神舟十九号成功返回",
          "discuss_num": 1234567,
          "topic_url": "https://tieba.baidu.com/hottopic/browse/hottopic?topic_id=28355142&topic_name=%E7%A5%9E%E8%88%9F"
        },
        {
          "topic_name": "热门话题三",
          "discuss_num": 999999,
          "topic_url": "https://tieba.baidu.com/hottopic/browse/hottopic?topic_id=28355143&topic_name=%E7%83%AD%E9%97%A8"
        }
      ]
    }
  }
}`

func newTestClient(ts *httptest.Server) *tieba.Client {
	cfg := tieba.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return tieba.NewClient(cfg)
}

func TestHotSendsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte(mockHotResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Hot(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHotParsesResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockHotResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	topics, err := c.Hot(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(topics) != 3 {
		t.Fatalf("got %d topics, want 3", len(topics))
	}

	top := topics[0]
	if top.Rank != 1 {
		t.Errorf("rank = %d, want 1", top.Rank)
	}
	if top.Name != "拉锯战!五星巴西战平摩洛哥" {
		t.Errorf("name = %q", top.Name)
	}
	if top.Discussions != 2656410 {
		t.Errorf("discussions = %d, want 2656410", top.Discussions)
	}
	if top.URL == "" {
		t.Error("URL is empty")
	}

	second := topics[1]
	if second.Rank != 2 {
		t.Errorf("rank = %d, want 2", second.Rank)
	}
}

func TestHotLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockHotResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	topics, err := c.Hot(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(topics) != 2 {
		t.Fatalf("got %d topics with limit 2, want 2", len(topics))
	}
}

func TestHotRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(mockHotResponse))
	}))
	defer srv.Close()

	cfg := tieba.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := tieba.NewClient(cfg)

	start := time.Now()
	_, err := c.Hot(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestHotUsesBaseURL(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(mockHotResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Hot(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/hottopic/browse/topicList" {
		t.Errorf("path = %q, want /hottopic/browse/topicList", gotPath)
	}
}
