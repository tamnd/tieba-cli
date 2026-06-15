package tieba

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes Tieba as a kit Domain so a multi-domain host can
// blank-import it:
//
//	import _ "github.com/tamnd/tieba-cli/tieba"
//
// The same Domain also builds the standalone tieba binary (cmd/tieba).
func init() { kit.Register(Domain{}) }

// Domain is the Tieba driver.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "tieba",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "tieba",
			Short:  "A command line for Baidu Tieba.",
			Long: `A command line for Baidu Tieba.

tieba reads public Baidu Tieba data over plain HTTPS, shapes it into clean
records, and prints output that pipes into the rest of your tools. No API key,
nothing to run alongside it.`,
			Site: Host,
			Repo: "https://github.com/tamnd/tieba-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// topic: resolver op so that kit can mint Topic records and answer
	// `tieba topic <name>` and `ant get tieba://topic/<name>`.
	kit.Handle(app, kit.OpMeta{Name: "topic", Group: "read", Single: true,
		URIType: "topic", Resolver: true,
		Summary: "Resolve a Tieba topic name to its URL",
		Args:    []kit.Arg{{Name: "name", Help: "topic name or URL"}}}, getTopic)

	// hot: current hot topics list.
	kit.Handle(app, kit.OpMeta{Name: "hot", Group: "read", List: true,
		URIType: "topic",
		Summary: "List the current hot topics on Baidu Tieba"}, getHot)
}

// newClient builds the HTTP client from the kit host config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type topicInput struct {
	Name   string  `kit:"arg"   help:"topic name or URL"`
	Client *Client `kit:"inject"`
}

type hotInput struct {
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func getTopic(_ context.Context, in topicInput, emit func(*Topic) error) error {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return errs.Usage("topic name is required")
	}
	// If passed a full URL, extract the topic_name query param.
	if strings.HasPrefix(name, "http") {
		if u, err := url.Parse(name); err == nil {
			if n := u.Query().Get("topic_name"); n != "" {
				name = n
			}
		}
	}
	return emit(&Topic{
		Name: name,
		URL:  "https://" + Host + "/f?kw=" + url.QueryEscape(name),
	})
}

func getHot(ctx context.Context, in hotInput, emit func(*Topic) error) error {
	limit := in.Limit
	topics, err := in.Client.Hot(ctx, limit)
	if err != nil {
		return err
	}
	for i := range topics {
		if err := emit(&topics[i]); err != nil {
			return err
		}
	}
	return nil
}

// --- Resolver ---

// topicURLRE matches a Tieba topic URL and captures the topic_name param.
var topicURLRE = regexp.MustCompile(`(?:https?://)?tieba\.baidu\.com/.*[?&]topic_name=([^&]+)`)

// Classify turns a Tieba URL or topic name into (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", errs.Usage("empty Tieba reference")
	}
	if m := topicURLRE.FindStringSubmatch(input); len(m) >= 2 {
		name, _ := url.QueryUnescape(m[1])
		if name == "" {
			name = m[1]
		}
		return "topic", name, nil
	}
	// Bare topic name.
	return "topic", input, nil
}

// Locate is the inverse: the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	if uriType != "topic" {
		return "", errs.Usage("tieba has no resource type %q", uriType)
	}
	return "https://" + Host + "/f?kw=" + url.QueryEscape(id), nil
}
