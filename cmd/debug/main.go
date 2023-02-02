package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labd/go-apollostudio-sdk/pkg/apollostudio"
)

func main() {
	godotenv.Load()

	apiKey := os.Getenv("API_KEY")
	schemaId := os.Getenv("SCHEMA_ID")
	schemaVariant := os.Getenv("SCHEMA_VARIANT")
	subGraphSchema := os.Getenv("SUB_GRAPH_SCHEMA")
	subGraphName := os.Getenv("SUB_GRAPH_NAME")

	ctx := context.Background()

	client := apollostudio.NewClient(apollostudio.ClientOpts{APIKey: apiKey})

	isValid, err := client.ValidateSubGraph(ctx, &apollostudio.ValidateOptions{
		SchemaID:       schemaId,
		SchemaVariant:  schemaVariant,
		SubGraphSchema: []byte(subGraphSchema),
		SubGraphName:   subGraphName,
	})

	if err != nil {
		log.Fatal(err)
	}

	if !isValid {
		// TODO: add more info, also see docs ValidateSubGraph.
		fmt.Println("schema validation failed")
		return
	}

	submits, err := client.SubmitSubGraph(ctx, &apollostudio.SubmitOptions{
		SchemaID:       schemaId,
		SchemaVariant:  schemaVariant,
		SubGraphName:   subGraphName,
		SubGraphSchema: []byte(subGraphSchema),
	})

	fmt.Println(submits, err)

}
