package apollostudio

import (
	"context"
	"fmt"

	"github.com/hasura/go-graphql-client"
)

type SubmitOptions struct {
	SchemaID       string
	SchemaVariant  string
	APIKey         string
	SubGraphName   string
	SubGraphSchema []byte
}

func (c *Client) SubmitSubGraph(ctx context.Context, opts *SubmitOptions) error {
	type Mutation struct {
		Graph struct {
			PublishSubgraph struct {
				CompositionConfig struct {
					SchemaHash string
				}
				LaunchUrl      string
				LaunchCliCopy  string
				Errors         []ApolloError
				UpdatedGateway bool
				WasCreated     bool
			} `graphql:"publishSubgraph(name: $subgraph, graphVariant: $variant, revision: $revision, activePartialSchema: $schema, gitContext: $git_context)"`
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
	}

	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return fmt.Errorf("failed to query apollo studio %v", err.Error())
	}

	errors := mutation.Graph.PublishSubgraph.Errors
	if len(errors) > 0 {
		return &OperationError{"failed to submit schema", errors}
	}

	return nil
}
