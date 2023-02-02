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

	cfg := apollostudio.Config{
		APIKey: apiKey,
	}

	client := apollostudio.NewClient(cfg)

	validates, err := client.ValidateSubGraph(ctx, &apollostudio.ValidateOptions{
		SchemaID:       schemaId,
		SchemaVariant:  schemaVariant,
		SubGraphSchema: []byte(subGraphSchema),
		SubGraphName:   subGraphName,
	})

	fmt.Println(validates, err)

	// validates, err := client.SubmitGraph(ctx, &apollo.SubmitOptions{
	// 	SchemaID:      "1",
	// 	SchemaVariant: "2",
	// 	SchemaBody:    []byte("foo"),
	// })

}
