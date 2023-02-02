package apollostudio

import (
	"net/http"

	"github.com/hasura/go-graphql-client"
)

type Client struct {
	httpClient *http.Client
	gqlClient  *graphql.Client
	key        string
	body       string
}

type ValidateOptions struct {
	SchemaID       string
	SchemaVariant  string
	APIKey         string
	SubGraphName   string
	SubGraphSchema []byte
}

type SubmitOptions struct {
	SchemaID       string
	SchemaVariant  string
	APIKey         string
	SubGraphName   string
	SubGraphSchema []byte
}

type ClientOpts struct {
	APIKey string
}

func NewClient(opts ClientOpts) *Client {
	httpClient := http.Client{Transport: &headerTransport{
		APIKey: opts.APIKey,
	}}

	gqlClient := graphql.NewClient("https://graphql.api.apollographql.com/api/graphql", &httpClient)

	return &Client{
		gqlClient:  gqlClient,
		httpClient: &httpClient,
	}
}
