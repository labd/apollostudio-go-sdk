package apollostudio

import "fmt"

type ApolloError struct {
	Message string
	Code    string
}

// OperationError contains Apollo Studio errors for a certain
// Apollo Studio operation such as submitting the schema.
// TODO: we probably get {Message, Code} errors back from validating
// too? If not and submit specific, rename and move to submit.go.
type OperationError struct {
	Message      string
	ApolloErrors []ApolloError
}

func (e *OperationError) Error() string {
	return fmt.Sprintf("%s (%d apollo errors)", e.Message, len(e.ApolloErrors))
}

func (e *OperationError) Print() {
	fmt.Println(e.Error())
	fmt.Println("Apollo errors:")
	for i, error := range e.ApolloErrors {
		fmt.Printf("#%d: code: %s, message: %s\n", i, error.Code, error.Message)
	}
}
