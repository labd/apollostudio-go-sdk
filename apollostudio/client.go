package apollostudio

import (
	"errors"
	"net/http"
	"strings"

	"github.com/hasura/go-graphql-client"
)

var defaultUrl = "https://graphql.api.apollographql.com/api/graphql"

type clientSettings struct {
	httpClient      *http.Client
	debug           bool
	url             string
	requestModifier graphql.RequestModifier
}

type Client struct {
	gqlClient *graphql.Client
	GraphRef  GraphRef
}

type ClientOpt func(client *clientSettings)

// WithHttpClient allows you to set a custom http client for fine-grained control of traffic
func WithHttpClient(httpClient *http.Client) ClientOpt {
	return func(settings *clientSettings) {
		settings.httpClient = httpClient
	}
}

// WithDebug allows you to run the client in debug mode, which will return internal error details
func WithDebug(debug bool) ClientOpt {
	return func(settings *clientSettings) {
		settings.debug = debug
	}
}

// WithUrl allows you to overrule the default url
func WithUrl(url string) ClientOpt {
	return func(settings *clientSettings) {
		settings.url = url
	}
}

// WithRequestModifier allows you to modify the request. Note that the API key is also set by ApiKeyRequestModifier,
// so make sure to also call this method if you want to further modify the request
func WithRequestModifier(modifier graphql.RequestModifier) ClientOpt {
	return func(settings *clientSettings) {
		settings.requestModifier = modifier
	}
}

type GraphRef string

func ValidateGraphRef(graphRef string) error {
	if graphRef == "" {
		return errors.New("graph ref is required")
	}

	p := strings.Split(graphRef, "@")

	if len(p) < 2 {
		return errors.New("missing variant in graph ref")
	}

	return nil
}

func (g *GraphRef) getGraphId() string {
	p := strings.Split(string(*g), "@")

	return p[0]
}

func (g *GraphRef) getVariant() string {

	p := strings.Split(string(*g), "@")

	return p[1]
}

func NewClient(apiKey string, graphRef string, opts ...ClientOpt) (*Client, error) {
	settings := &clientSettings{
		httpClient:      http.DefaultClient,
		debug:           false,
		url:             defaultUrl,
		requestModifier: ApiKeyRequestModifier(apiKey),
	}

	for _, opt := range opts {
		opt(settings)
	}

	err := ValidateGraphRef(graphRef)
	if err != nil {
		return nil, err
	}

	gqlClient := graphql.NewClient(settings.url, settings.httpClient)
	gqlClient = gqlClient.WithDebug(settings.debug)
	gqlClient = gqlClient.WithRequestModifier(settings.requestModifier)

	return &Client{
		gqlClient: gqlClient,
		GraphRef:  GraphRef(graphRef),
	}, nil
}

func ApiKeyRequestModifier(apiKey string) graphql.RequestModifier {
	return func(req *http.Request) {
		req.Header.Add("x-api-key", apiKey)
	}
}
