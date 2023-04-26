package apollostudio

import (
	"context"
	"fmt"
	"github.com/hasura/go-graphql-client"
)

func (c *Client) RemoveSubGraph(ctx context.Context, name string) error {
	type Mutation struct {
		Graph struct {
			RemoveImplementingServiceAndTriggerComposition struct {
				Errors         []ApolloError
				UpdatedGateway bool
			} `graphql:"removeImplementingServiceAndTriggerComposition(graphVariant: $variant, name: $subgraph)"`
		} `graphql:"graph(id: $graph_id)"`
	}

	var mutation Mutation

	vars := map[string]interface{}{
		"graph_id": graphql.ID(c.GraphID),
		"subgraph": graphql.String(name),
		"variant":  graphql.String(c.Variant),
	}

	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return fmt.Errorf("failed to query apollo studio %v", err.Error())
	}

	errors := mutation.Graph.RemoveImplementingServiceAndTriggerComposition.Errors
	if len(errors) > 0 {
		return &OperationError{"failed to remove schema", errors}
	}

	return nil
}
