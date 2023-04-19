package apollostudio

import (
	"context"
	"fmt"
	"time"

	"github.com/hasura/go-graphql-client"
)

type ReadOptions struct {
	SchemaID      string
	SchemaVariant string
	SubGraphName  string
}

type SubGraphResult struct {
	GraphID             string `graphql:"graphID"`
	ActivePartialSchema struct {
		Sdl string
	}
	CreatedAt    time.Time
	GraphVariant string
	Name         string
	Revision     string
	UpdatedAt    time.Time
	URL          string `graphql:"url"`
}

func (c *Client) ReadSubGraph(ctx context.Context, opts *ReadOptions) (*SubGraphResult, error) {
	type Query struct {
		Graph struct {
			Variant struct {
				Subgraph SubGraphResult `graphql:"subgraph(name: $subgraph)"`
			} `graphql:"variant(name: $variant)"`
		} `graphql:"graph(id: $graph_id)"`
	}

	var query Query

	vars := map[string]interface{}{
		"graph_id": graphql.ID(opts.SchemaID),
		"subgraph": graphql.ID(opts.SubGraphName),
		"variant":  graphql.String(opts.SchemaVariant),
	}

	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to query apollo studio %w", err)
	}

	return &query.Graph.Variant.Subgraph, nil
}
