package apollostudio

import "fmt"

type ApolloError struct {
	Message string
	Code    string
}

type OperationError struct {
	Message      string
	ApolloErrors []ApolloError
}

func (e *OperationError) Error() string {
	return fmt.Sprintf("%s (%d errors)", e.Message, len(e.ApolloErrors))
}
