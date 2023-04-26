package apollostudio

import (
	"context"
	"fmt"
	"time"

	"github.com/hasura/go-graphql-client"
)

const (
	ValidationStatusFailed  ValidationStatus = "FAILED"
	ValidationStatusPassed  ValidationStatus = "PASSED"
	ValidationStatusPending ValidationStatus = "PENDING"
	ValidationStatusBlocked ValidationStatus = "BLOCKED"
)

type ValidationStatus string

type ValidateOptions struct {
	SubGraphName   string
	SubGraphSchema []byte
}

type ValidationResult struct {
	Status ValidationStatus
	Tasks  []ValidationTask
}

type ValidationTask interface {
	GetStatus() ValidationStatus
	Changes() []string
	Errors() []ApolloError
}

type ValidationOperationsCheckTask struct {
	Status ValidationStatus
	Result struct {
		Changes []struct {
			Severity    string
			Code        string
			Description string
		}
		NumberOfCheckedOperations int
	}
}

type ValidationCompositionCheckTask struct {
	Status ValidationStatus
	Result struct {
		GraphCompositionID string `graphql:"graphCompositionID"`
		Errors             []struct {
			Message   string
			Code      string
			Locations []struct {
				Column int
				Line   int
			}
		}
	}
}

// NewValidationResult creates a new ValidationResult with the given status
func NewValidationResult(s ValidationStatus) *ValidationResult {
	return &ValidationResult{
		Status: s,
		Tasks:  make([]ValidationTask, 0),
	}
}

// IsValid returns true if the proposed schema is valid
func (vr ValidationResult) IsValid() bool {
	return vr.Status == ValidationStatusPassed
}

// Errors returns a list of errors from all failed tasks
func (vr ValidationResult) Errors() []ApolloError {
	var errors []ApolloError
	for _, task := range vr.Tasks {
		if task.GetStatus() == ValidationStatusFailed {
			errors = append(errors, task.Errors()...)
		}
	}
	return errors
}

func (vr ValidationResult) Changes() []string {
	var changes []string
	for _, task := range vr.Tasks {
		changes = append(changes, task.Changes()...)
	}
	return changes
}

func (t ValidationOperationsCheckTask) Errors() []ApolloError {
	var errors []ApolloError
	for _, c := range t.Result.Changes {
		errors = append(
			errors, ApolloError{
				Message: fmt.Sprintf("[%s]: %s", c.Severity, c.Description),
				Code:    c.Code,
			},
		)
	}
	return errors
}

func (t ValidationOperationsCheckTask) Changes() []string {
	var changes []string
	for _, c := range t.Result.Changes {
		changes = append(changes, fmt.Sprintf("[%s]: %s", c.Severity, c.Description))
	}

	return changes
}

func (t ValidationOperationsCheckTask) GetStatus() ValidationStatus {
	return t.Status
}

func (t ValidationCompositionCheckTask) Errors() []ApolloError {
	var errors []ApolloError
	for _, e := range t.Result.Errors {
		locations := ""
		for _, loc := range e.Locations {
			locations += fmt.Sprintf("line %d, column %d", loc.Line, loc.Column)
		}
		msg := e.Message
		if len(locations) > 0 {
			msg = fmt.Sprintf("%s at %s", e.Message, locations)
		}
		errors = append(
			errors, ApolloError{
				Message: msg,
				Code:    e.Code,
			},
		)
	}
	return errors
}

func (t ValidationCompositionCheckTask) Changes() []string {
	return []string{}
}

func (t ValidationCompositionCheckTask) GetStatus() ValidationStatus {
	return t.Status
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
			} "graphql:\"variant(name: $variant)\""
		} `graphql:"graph(id: $graph_id)"`
	}

	var mutation Mutation

	vars := map[string]interface{}{
		"graph_id": graphql.ID(c.GraphID),
		"variant":  graphql.String(c.Variant),
		"input": SubgraphCheckAsyncInput{
			Config:         HistoricQueryParametersInput{},
			GitContext:     GitContextInput{},
			GraphRef:       c.GraphRef,
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
		// TODO: pass some more details using smth like OperationError?
		return "", fmt.Errorf("could not create check workflow in apollo studio")
	}
	return *workflowId, nil
}

// checkWorkflow polls the status of the workflow and returns the result when failed or passed.
func (c *Client) checkWorkflow(ctx context.Context, workflowId string) (
	*ValidationResult, error,
) {
	type Query struct {
		Graph struct {
			CheckWorkFlow struct {
				Status    string
				CreatedAt string
				Tasks     []struct {
					Typename             string                         `graphql:"__typename"`
					OperationsCheckTask  ValidationOperationsCheckTask  `graphql:"... on OperationsCheckTask"`
					CompositionCheckTask ValidationCompositionCheckTask `graphql:"... on CompositionCheckTask"`
				}
			} `graphql:"checkWorkflow(id: $workflowId)"`
		} `graphql:"graph(id: $graph_id)"`
	}

	vars := map[string]interface{}{
		"graph_id":   graphql.ID(c.GraphID),
		"workflowId": graphql.ID(workflowId),
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		var query Query

		err := c.gqlClient.Query(ctx, &query, vars)
		if err != nil {
			return nil, fmt.Errorf("failed to query apollo studio %w", err)
		}

		workflow := query.Graph.CheckWorkFlow
		status := ValidationStatus(workflow.Status)

		// we are setting task results not only when validation fails but also when it passes
		// because we want to show the user what is going to change in the graph
		switch status {
		case ValidationStatusBlocked:
			fallthrough
		case ValidationStatusFailed:
			fallthrough
		case ValidationStatusPassed:
			vr := NewValidationResult(status)
			for _, task := range workflow.Tasks {
				switch task.Typename {
				case "OperationsCheckTask":
					vr.Tasks = append(vr.Tasks, task.OperationsCheckTask)
				case "CompositionCheckTask":
					vr.Tasks = append(vr.Tasks, task.CompositionCheckTask)
				}
			}
			return vr, nil
		case ValidationStatusPending:
		default:
			fmt.Printf("WARNING: unknown workflow state %s\n", status)
		}

		time.Sleep(1 * time.Second)
	}
}

// ValidateSubGraph submits the proposed schema and returns the result of the async workflow.
func (c *Client) ValidateSubGraph(ctx context.Context, opts *ValidateOptions) (*ValidationResult, error) {
	id, err := c.submitSubgraphCheck(ctx, opts)
	if err != nil {
		return nil, err
	}

	return c.checkWorkflow(ctx, id)
}
