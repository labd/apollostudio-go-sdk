package apollostudio

import (
	"context"
	"fmt"
	"github.com/hasura/go-graphql-client"
)

type RemoveOptions struct {
	SchemaID      string
	SchemaVariant string
	SubGraphName  string
}

func (c *Client) RemoveSubGraph(ctx context.Context, opts *RemoveOptions) error {
	type Mutation struct {
		Graph struct {
			RemoveImplementingServiceAndTriggerComposition struct {
				Errors []ApolloError
			} `graphql:"removeImplementingServiceAndTriggerComposition(graphVariant: $variant, name: $subgraph)"`
		} `graphql:"graph(id: $graph_id)"`
	}

	var mutation Mutation

	vars := map[string]interface{}{
		"graph_id": graphql.ID(opts.SchemaID),
		"subgraph": graphql.String(opts.SubGraphName),
		"variant":  graphql.String(opts.SchemaVariant),
	}

	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return fmt.Errorf("failed to query apollo studio %v", err.Error())
	}

	errors := mutation.Graph.RemoveImplementingServiceAndTriggerComposition.Errors
	if len(errors) > 0 {
		return &OperationError{"failed to submit schema", errors}
	}

	return nil
}
