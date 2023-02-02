package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labd/go-apollostudio-sdk/pkg/apollostudio"
)

func handleErr(err error) {
	var opErr *apollostudio.OperationError
	if errors.As(err, &opErr) {
		opErr.Print()
		os.Exit(1)
	}
	log.Fatal(err)
}

func main() {
	godotenv.Load()

	apiKey := os.Getenv("API_KEY")
	schemaId := os.Getenv("SCHEMA_ID")
	schemaVariant := os.Getenv("SCHEMA_VARIANT")
	subGraphSchema := os.Getenv("SUB_GRAPH_SCHEMA")
	subGraphName := os.Getenv("SUB_GRAPH_NAME")

	ctx := context.Background()

	client := apollostudio.NewClient(apollostudio.ClientOpts{APIKey: apiKey})

	valid, err := client.ValidateSubGraph(ctx, &apollostudio.ValidateOptions{
		SchemaID:       schemaId,
		SchemaVariant:  schemaVariant,
		SubGraphSchema: []byte(subGraphSchema),
		SubGraphName:   subGraphName,
	})

	if err != nil {
		handleErr(err)
	}

	if !valid {
		// TODO: add more info, also see docs ValidateSubGraph.
		fmt.Println("schema validation failed")
		return
	}

	if err := client.SubmitSubGraph(ctx, &apollostudio.SubmitOptions{
		SchemaID:       schemaId,
		SchemaVariant:  schemaVariant,
		SubGraphName:   subGraphName,
		SubGraphSchema: []byte(subGraphSchema),
	}); err != nil {
		handleErr(err)
	}

	fmt.Println("schema submitted")
}
