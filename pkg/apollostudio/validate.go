package apollostudio

import (
	"context"
	"fmt"
	"time"

	"github.com/hasura/go-graphql-client"
)

type ValidateOptions struct {
	SchemaID       string
	SchemaVariant  string
	APIKey         string
	SubGraphName   string
	SubGraphSchema []byte
}

// submitSubgraphCheck submits the proposed schema and returns a workflow id.
func (c *Client) submitSubgraphCheck(ctx context.Context, opts *ValidateOptions) (string, error) {
	type Mutation struct {
		Graph struct {
			Variant struct {
				SubmitSubgraphCheckAsync struct {
					CheckRequestSuccess struct {
						TargetURL  string  `graphql:"targetURL"`
						WorkflowID *string `graphql:"workflowID"`
					} `graphql:"... on CheckRequestSuccess"`
					InvalidInputError struct {
						message string
					} `graphql:"... on InvalidInputError"`
					PermissionError struct {
						message string
					} `graphql:"... on PermissionError"`
					PlanError struct {
						message string
					} `graphql:"... on PlanError"`
				} `graphql:"submitSubgraphCheckAsync(input: $input)"`
			} "graphql:\"variant(name: $name)\""
		} `graphql:"graph(id: $graph_id)"`
	}

	var mutation Mutation

	vars := map[string]interface{}{
		"graph_id": graphql.ID(opts.SchemaID),
		"name":     graphql.String(opts.SchemaVariant),
		"input": SubgraphCheckAsyncInput{
			Config:         HistoricQueryParametersInput{},
			GitContext:     GitContextInput{},
			GraphRef:       fmt.Sprintf("%s@%s", opts.SchemaID, opts.SchemaVariant),
			IsSandbox:      false,
			ProposedSchema: string(opts.SubGraphSchema),
			SubgraphName:   opts.SubGraphName,
		},
	}

	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return "", fmt.Errorf("failed to query apollo studio: %w", err)
	}

	workflowId := mutation.Graph.Variant.SubmitSubgraphCheckAsync.CheckRequestSuccess.WorkflowID
	if workflowId == nil {
		return "", fmt.Errorf("could not create check workflow in apollo studio")
	}
	return *workflowId, nil
}

// checkWorkflow polls the status of the of the workflow and returns the result when failed or passed.
func (c *Client) checkWorkflow(ctx context.Context, opts *ValidateOptions, workflowId string) (bool, error) {
	type Query struct {
		Graph struct {
			CheckWorkFlow struct {
				Status    string
				CreatedAt string
				Tasks     []struct {
					Typename            string `graphql:"__typename"`
					OperationsCheckTask struct {
						Status string
						Result struct {
							Changes []struct {
								Severity    string
								Code        string
								Description string
							}
							NumberOfCheckedOperations int
						}
					} `graphql:"... on OperationsCheckTask"`
				}
			} `graphql:"checkWorkflow(id: $workflowId)"`
		} `graphql:"graph(id: $graph_id)"`
	}

	vars := map[string]interface{}{
		"graph_id":   graphql.ID(opts.SchemaID),
		"workflowId": graphql.ID(workflowId),
	}

	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		var query Query

		err := c.gqlClient.Query(ctx, &query, vars)
		if err != nil {
			return false, fmt.Errorf("failed to query apollo studio %w", err)
		}

		workflow := query.Graph.CheckWorkFlow
		status := workflow.Status

		switch status {
		case "FAILED":
			// TODO: pretty print (JSON maybe)
			return false, fmt.Errorf("Subgraph check failed %v", workflow)
		case "PASSED":
			return true, nil
		case "PENDING":
		default:
			fmt.Printf("WARNING: unknown workflow state %s\n", status)
		}

		time.Sleep(1 * time.Second)
	}
}

// ValidateSubGraph submits the proposed schema and returns the result of the async workflow.
func (c *Client) ValidateSubGraph(ctx context.Context, opts *ValidateOptions) (bool, error) {
	workflowId, err := c.submitSubgraphCheck(ctx, opts)
	if err != nil {
		return false, err
	}

	return c.checkWorkflow(ctx, opts, workflowId)
}
