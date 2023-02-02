package apollostudio

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

// XXX: why not <Something>Config
type HistoricQueryParametersInput struct {
	ExcludedClients               *string `json:"excludedClients"`
	ExcludedOperationNames        *string `json:"excludedOperationNames"`
	From                          *string `json:"from"`
	IgnoredOperations             *string `json:"ignoredOperations"`
	IncludedVariants              *string `json:"includedVariants"`
	QueryCountThreshold           *string `json:"queryCountThreshold"`
	QueryCountThresholdPercentage *string `json:"queryCountThresholdPercentage"`
	To                            *string `json:"to"`
}

type GitContextInput struct {
	Branch    *string `json:"branch"`
	Commit    *string `json:"commit"`
	Committer *string `json:"committer"`
	Message   *string `json:"message"`
	RemoteUrl *string `json:"remoteUrl"`
}

type PartialSchemaInput struct {
	Sdl string `json:"sdl"`
}

type SubgraphCheckAsyncInput struct {
	Config                HistoricQueryParametersInput `json:"config"`
	GitContext            GitContextInput              `json:"gitContext"`
	GraphRef              string                       `json:"graphRef"`
	IntrospectionEndpoint *string                      `json:"introspectionEndpoint"`
	IsSandbox             bool                         `json:"isSandbox"`
	ProposedSchema        string                       `json:"proposedSchema"`
	SubgraphName          string                       `json:"subgraphName"`
}

type HeaderTransport struct {
	APIKey string
}

func (t *HeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("x-api-key", t.APIKey)
	return http.DefaultTransport.RoundTrip(req)
}

func NewClient(opts ClientOpts) *Client {
	httpClient := http.Client{Transport: &HeaderTransport{
		APIKey: opts.APIKey,
	}}

	gqlClient := graphql.NewClient("https://graphql.api.apollographql.com/api/graphql", &httpClient)

	return &Client{
		gqlClient:  gqlClient,
		httpClient: &httpClient,
	}
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
