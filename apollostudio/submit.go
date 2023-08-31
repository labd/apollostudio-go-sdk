package apollostudio

import (
	"context"
	"fmt"
	"github.com/hasura/go-graphql-client"
)

type SubmitOptions struct {
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

func (c *Client) SubmitSubGraph(ctx context.Context, opts *SubmitOptions) (*SubmitSubgraphResult, error) {
	type Mutation struct {
		Graph struct {
			PublishSubgraph SubmitSubgraphResult `graphql:"publishSubgraph(name: $subgraph, graphVariant: $variant, revision: $revision, activePartialSchema: $schema, gitContext: $git_context, url: $url)"`
		} `graphql:"graph(id: $graph_id)"`
	}

	var mutation Mutation

	vars := map[string]interface{}{
		"graph_id": graphql.ID(c.GraphRef.getGraphId()),
		"subgraph": opts.SubGraphName,
		"variant":  c.GraphRef.getVariant(),
		"revision": "",
		"schema": PartialSchemaInput{
			Sdl: string(opts.SubGraphSchema),
		},
		"git_context": GitContextInput{},
		// URL is necessary if sub graph does not exist and is created
		// during the submission
		"url": opts.SubGraphURL,
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
