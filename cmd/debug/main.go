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
	subGraphURL := os.Getenv("SUB_GRAPH_URL")

	ctx := context.Background()
	client := apollostudio.NewClient(apollostudio.ClientOpts{APIKey: apiKey})

	vr, err := client.ValidateSubGraph(
		ctx, &apollostudio.ValidateOptions{
			SchemaID:       schemaId,
			SchemaVariant:  schemaVariant,
			SubGraphSchema: []byte(subGraphSchema),
			SubGraphName:   subGraphName,
		},
	)

	if err != nil {
		handleErr(err)
	}

	if !vr.IsValid() {
		fmt.Println("schema is not valid")
		fmt.Println(vr.Errors())
		return
	}

	fmt.Println("schema validated")
	fmt.Println(vr.Changes())

	rr, err := client.ReadSubGraph(
		ctx, &apollostudio.ReadOptions{
			SchemaID:      schemaId,
			SchemaVariant: schemaVariant,
			SubGraphName:  subGraphName,
		},
	)

	if err != nil {
		handleErr(err)
	}

	fmt.Println("schema read")
	fmt.Println(rr)

	sr, err := client.SubmitSubGraph(
		ctx, &apollostudio.SubmitOptions{
			SchemaID:       schemaId,
			SchemaVariant:  schemaVariant,
			SubGraphName:   subGraphName,
			SubGraphSchema: []byte(subGraphSchema),
			SubGraphURL:    subGraphURL,
		},
	)

	if err != nil {
		handleErr(err)
	}

	fmt.Println("schema submitted")
	fmt.Println(sr)

	err = client.RemoveSubGraph(
		ctx, &apollostudio.RemoveOptions{
			SchemaID:      schemaId,
			SchemaVariant: schemaVariant,
			SubGraphName:  subGraphName,
		},
	)

	if err != nil {
		handleErr(err)
	}

	fmt.Println("schema deleted")
}
