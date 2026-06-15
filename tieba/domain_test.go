package tieba

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

// These tests are offline: they exercise the URI driver's pure string functions
// and the host wiring (mint, resolve), which need no network. The client's
// HTTP behaviour is covered in tieba_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "tieba" {
		t.Errorf("Scheme = %q, want tieba", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "tieba" {
		t.Errorf("Identity.Binary = %q, want tieba", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct{ in, typ, id string }{
		{"足球", "topic", "足球"},
		{"NBA", "topic", "NBA"},
		{"https://tieba.baidu.com/hottopic/browse/hottopic?topic_id=123&topic_name=%E8%B6%B3%E7%90%83", "topic", "足球"},
	}
	for _, tc := range cases {
		typ, id, err := Domain{}.Classify(tc.in)
		if err != nil || typ != tc.typ || id != tc.id {
			t.Errorf("Classify(%q) = (%q, %q, %v), want (%q, %q, nil)",
				tc.in, typ, id, err, tc.typ, tc.id)
		}
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("topic", "足球")
	want := "https://tieba.baidu.com/f?kw=%E8%B6%B3%E7%90%83"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

// TestHostWiring mounts the driver in a kit Host and checks the round trip:
// a record mints to its URI and a bare name resolves back to the same URI.
func TestHostWiring(t *testing.T) {
	h, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}

	topic := &Topic{Rank: 1, Name: "足球", URL: "https://tieba.baidu.com/f?kw=%E8%B6%B3%E7%90%83"}
	u, err := h.Mint(topic)
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	// kit percent-encodes the id in the URI; both forms are equivalent.
	uStr := u.String()
	if uStr != "tieba://topic/足球" && uStr != "tieba://topic/%E8%B6%B3%E7%90%83" {
		t.Errorf("Mint = %q, want tieba://topic/足球 or percent-encoded form", uStr)
	}

	got, err := h.ResolveOn("tieba", "足球")
	if err != nil {
		t.Fatalf("ResolveOn error: %v", err)
	}
	gStr := got.String()
	if gStr != "tieba://topic/足球" && gStr != "tieba://topic/%E8%B6%B3%E7%90%83" {
		t.Errorf("ResolveOn = %q, want tieba://topic/足球 or percent-encoded form", gStr)
	}
}
