package apollostudio

import (
	"context"
	"fmt"
	"time"

	"github.com/hasura/go-graphql-client"
)

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

type LatestSchemaBuild struct {
	Result struct {
		BuildFailure struct {
			ErrorMessages []ApolloError
		} `graphql:"... on BuildFailure"`
	}
	Input struct {
		CompositionBuildInput struct {
			SubGraphs []struct {
				Name string
				Hash string
			} `graphql:"subgraphs"`
		} `graphql:"... on CompositionBuildInput"`
	}
}

func (b *LatestSchemaBuild) ContainsGraph(graphName string) bool {
	for _, v := range b.Input.CompositionBuildInput.SubGraphs {
		if v.Name == graphName {
			return true
		}
	}
	return false
}

func (b *LatestSchemaBuild) Errors() []ApolloError {
	return b.Result.BuildFailure.ErrorMessages
}

func (c *Client) GetSubGraph(ctx context.Context, name string) (*SubGraphResult, error) {
	type Query struct {
		Graph struct {
			Variant struct {
				Subgraph SubGraphResult `graphql:"subgraph(name: $subgraph)"`
			} `graphql:"variant(name: $variant)"`
		} `graphql:"graph(id: $graph_id)"`
	}

	var query Query

	vars := map[string]interface{}{
		"graph_id": graphql.ID(c.GraphRef.getGraphId()),
		"subgraph": graphql.ID(name),
		"variant":  c.GraphRef.getVariant(),
	}

	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to query apollo studio %w", err)
	}

	return &query.Graph.Variant.Subgraph, nil
}

func (c *Client) GetLatestSchemaBuild(ctx context.Context) (*LatestSchemaBuild, error) {
	type Query struct {
		Variant struct {
			GraphVariant struct {
				LatestLaunch struct {
					Build LatestSchemaBuild
				}
			} `graphql:"... on GraphVariant"`
		} `graphql:"variant(ref: $ref)"`
	}

	var query Query

	vars := map[string]interface{}{
		"ref": graphql.ID(c.GraphRef),
	}

	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to query apollo studio %w", err)
	}

	errors := query.Variant.GraphVariant.LatestLaunch.Build.Result.BuildFailure.ErrorMessages
	if len(errors) > 0 {
		return nil, &OperationError{
			fmt.Sprintf("latest schema build contains %d errors", len(errors)),
			errors,
		}
	}

	return &query.Variant.GraphVariant.LatestLaunch.Build, nil
}
