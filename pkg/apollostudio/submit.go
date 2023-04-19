package apollostudio

import (
	"context"
	"fmt"

	"github.com/hasura/go-graphql-client"
)

type SubmitOptions struct {
	SchemaID       string
	SchemaVariant  string
	SubGraphName   string
	SubGraphSchema []byte
	SubGraphURL    string
}

type SubmitSubgraphResult struct {
	CompositionConfig struct {
		SchemaHash string
	}
	LaunchUrl      string
	LaunchCliCopy  string
	Errors         []ApolloError
	UpdatedGateway bool
	WasCreated     bool
}

// OperationError contains Apollo Studio errors for a certain
// Apollo Studio operation such as submitting the schema.
type OperationError struct {
	Message      string
	ApolloErrors []ApolloError
}

func (e *OperationError) Error() string {
	return fmt.Sprintf("%s (%d apollo errors)", e.Message, len(e.ApolloErrors))
}

func (e *OperationError) Print() {
	fmt.Println(e.Error())
	fmt.Println("Apollo errors:")
	for i, err := range e.ApolloErrors {
		fmt.Printf("#%d: code: %s, message: %s\n", i, err.Code, err.Message)
	}
}

func (c *Client) SubmitSubGraph(ctx context.Context, opts *SubmitOptions) (*SubmitSubgraphResult, error) {
	type Mutation struct {
		Graph struct {
			PublishSubgraph SubmitSubgraphResult `graphql:"publishSubgraph(name: $subgraph, graphVariant: $variant, revision: $revision, activePartialSchema: $schema, gitContext: $git_context, url: $url)"`
		} `graphql:"graph(id: $graph_id)"`
	}

	var mutation Mutation

	vars := map[string]interface{}{
		"graph_id": graphql.ID(opts.SchemaID),
		"subgraph": graphql.String(opts.SubGraphName),
		"variant":  graphql.String(opts.SchemaVariant),
		"revision": graphql.String(""),
		"schema": PartialSchemaInput{
			Sdl: string(opts.SubGraphSchema),
		},
		"git_context": GitContextInput{},
		// URL is necessary if sub graph does not exist and is created
		// during the submission
		"url": graphql.String(opts.SubGraphURL),
	}

	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to query apollo studio %v", err.Error())
	}

	errors := mutation.Graph.PublishSubgraph.Errors
	if len(errors) > 0 {
		return nil, &OperationError{"failed to submit schema", errors}
	}

	return &mutation.Graph.PublishSubgraph, nil
}
