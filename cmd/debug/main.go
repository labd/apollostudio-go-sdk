package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/labd/go-apollostudio-sdk/apollostudio"
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

	_, err := client.ValidateSubGraph(ctx, &apollostudio.ValidateOptions{
		SchemaID:       schemaId,
		SchemaVariant:  schemaVariant,
		SubGraphSchema: []byte(subGraphSchema),
		SubGraphName:   subGraphName,
	})

	// fmt.Println(validates, err)

	submits, err := client.SubmitSubGraph(ctx, &apollostudio.SubmitOptions{
		SchemaID:       schemaId,
		SchemaVariant:  schemaVariant,
		SubGraphName:   subGraphName,
		SubGraphSchema: []byte(subGraphSchema),
	})

	fmt.Println(submits, err)

}
