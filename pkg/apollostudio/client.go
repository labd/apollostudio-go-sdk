package apollostudio

import (
	"errors"
	"net/http"
	"strings"

	"github.com/hasura/go-graphql-client"
)

type Client struct {
	httpClient *http.Client
	gqlClient  *graphql.Client
	key        string
	GraphRef   string
	GraphID    string
	Variant    string
}

type ClientOpts struct {
	APIKey   string
	GraphRef string
}

func NewClient(opts ClientOpts) (*Client, error) {
	httpClient := http.Client{
		Transport: &headerTransport{
			APIKey: opts.APIKey,
		},
	}

	gqlClient := graphql.NewClient(
		"https://graphql.api.apollographql.com/api/graphql",
		&httpClient,
	)

	if opts.GraphRef == "" {
		return nil, errors.New("graph ref is required")
	}

	p := strings.Split(opts.GraphRef, "@")
	if len(p) < 2 {
		return nil, errors.New("missing variant in graph ref")
	}

	return &Client{
		gqlClient:  gqlClient,
		httpClient: &httpClient,
		GraphRef:   opts.GraphRef,
		GraphID:    p[0],
		Variant:    p[1],
	}, nil
}
