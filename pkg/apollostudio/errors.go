package apollostudio

import (
	"fmt"
	"strings"
)

type ApolloError struct {
	Message string
	Code    string
}

// OperationError contains Apollo Studio errors for a certain
// Apollo Studio operation such as submitting the schema.
type OperationError struct {
	Message      string
	ApolloErrors []ApolloError
}

func (e *OperationError) Error() string {
	msgs := make([]string, len(e.ApolloErrors))
	for i, err := range e.ApolloErrors {
		msgs = append(msgs, fmt.Sprintf("#%d: code: %s, message: %s\n", i, err.Code, err.Message))
	}
	return fmt.Sprintf("%s (%d apollo errors): %s", e.Message, len(e.ApolloErrors), strings.Join(msgs, ""))
}

// IsOperationError returns true if the error is an OperationError.
// This is useful because, for example, submitting the schema may return
// an error, but the subgraph will still be created. As a result, the
// federation schema remains unaffected, while the subgraph coexists with errors.
func IsOperationError(err error) bool {
	_, ok := err.(*OperationError)
	return ok
}
