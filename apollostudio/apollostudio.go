package apollostudio

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hasura/go-graphql-client"
)

type Client struct {
	httpClient    *http.Client
	graphqlClient *graphql.Client
	key           string
	body          string
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

type Config struct {
	APIKey string
}

type headerTransport struct {
	APIKey string
}

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

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("x-api-key", t.APIKey)
	return http.DefaultTransport.RoundTrip(req)
}

func NewClient(options Config) *Client {
	httpClient := http.Client{Transport: &headerTransport{
		APIKey: options.APIKey,
	}}

	graphqlClient := graphql.NewClient("https://graphql.api.apollographql.com/api/graphql", &httpClient)

	return &Client{
		graphqlClient: graphqlClient,
		httpClient:    &httpClient,
	}

}

func (c *Client) ValidateSubGraph(ctx context.Context, opts *ValidateOptions) (bool, error) {
	var subgraphCheckMutation struct {
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
			} `graphql:"variant(name: $name)"`
		} `graphql:"graph(id: $graph_id)"`
	}

	subgraphCheckVariables := map[string]interface{}{
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

	checkErr := c.graphqlClient.Mutate(context.Background(), &subgraphCheckMutation, subgraphCheckVariables)
	if checkErr != nil {
		return false, fmt.Errorf("failed to query apollo studio %v", checkErr.Error())
	}

	workflowId := subgraphCheckMutation.Graph.Variant.SubmitSubgraphCheckAsync.CheckRequestSuccess.WorkflowID
	if workflowId == nil {
		return false, fmt.Errorf("could not create check workflow in apollo studio")
	}

	subgraphCheckWorkflowVariables := map[string]interface{}{
		"graph_id":   graphql.ID(opts.SchemaID),
		"workflowId": graphql.ID(*workflowId),
	}

	type SubgraphCheckWorkflowQuery struct {
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

	for {
		var subgraphCheckWorkflowQuery SubgraphCheckWorkflowQuery

		workflowErr := c.graphqlClient.Query(context.Background(), &subgraphCheckWorkflowQuery, subgraphCheckWorkflowVariables)

		if workflowErr != nil {
			return false, fmt.Errorf("failed to query apollo studio %v", workflowErr.Error())
		}

		time.Sleep(1 * time.Second)

		if subgraphCheckWorkflowQuery.Graph.CheckWorkFlow.Status == "FAILED" {
			return false, fmt.Errorf("Subgraph check failed %v", subgraphCheckWorkflowQuery.Graph.CheckWorkFlow)
		}

		if subgraphCheckWorkflowQuery.Graph.CheckWorkFlow.Status == "PASSED" {
			return true, nil
		}

		// if not FAILED or PASSED it is PENDING, so for loop is ran again
	}

}

func (c *Client) SubmitSubGraph(ctx context.Context, opts *SubmitOptions) (bool, error) {
	var uploadSchemaMutation struct {
		Graph struct {
			PublishSubgraph struct {
				CompositionConfig struct {
					SchemaHash string
				}
				LaunchUrl     string
				LaunchCliCopy string
				Errors        []struct {
					Message string
					Code    string
				}
				UpdatedGateway bool
				WasCreated     bool
			} `graphql:"publishSubgraph(name: $subgraph, graphVariant: $variant, revision: $revision, activePartialSchema: $schema, gitContext: $git_context)"`
		} `graphql:"graph(id: $graph_id)"`
	}

	uploadSchemaVariables := map[string]interface{}{
		"graph_id": graphql.ID(opts.SchemaID),
		"subgraph": graphql.String(opts.SubGraphName),
		"variant":  graphql.String(opts.SchemaVariant),
		"revision": graphql.String(""),
		"schema": PartialSchemaInput{
			Sdl: string(opts.SubGraphSchema),
		},
		"git_context": GitContextInput{},
	}

	checkErr := c.graphqlClient.Mutate(context.Background(), &uploadSchemaMutation, uploadSchemaVariables)
	if checkErr != nil {
		return false, fmt.Errorf("failed to query apollo studio %v", checkErr.Error())
	}

	submitErr := uploadSchemaMutation.Graph.PublishSubgraph.Errors
	hasErr := len(submitErr) > 0
	if hasErr {
		return false, fmt.Errorf("failed to submit schema %v", submitErr)
	}

	return true, nil
}
