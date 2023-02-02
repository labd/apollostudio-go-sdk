package apollostudio

import (
	"context"
	"fmt"

	"github.com/hasura/go-graphql-client"
)

func (c *Client) SubmitSubGraph(ctx context.Context, opts *SubmitOptions) (bool, error) {
	type Mutation struct {
		Graph struct {
			PublishSubgraph struct {
				CompositionConfig struct {
					SchemaHash string
				}
				LaunchUrl     string
				LaunchCliCopy string
				// TODO: make error a type so we can re-use when returning collected errors
				Errors []struct {
					Message string
					Code    string
				}
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
		return false, fmt.Errorf("failed to query apollo studio %v", err.Error())
	}

	errors := mutation.Graph.PublishSubgraph.Errors
	if len(errors) > 0 {
		// TODO: make new error struct that contians "gqlErrors" or so
		return false, fmt.Errorf("failed to submit schema %v", errors)
	}

	return true, nil
}
