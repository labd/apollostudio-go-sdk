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
		fmt.Println(opErr.Error())
		os.Exit(1)
	}
	log.Fatal(err)
}

func main() {
	godotenv.Load()

	apiKey := os.Getenv("APOLLO_API_KEY")
	graphRef := os.Getenv("APOLLO_GRAPH_REF")
	subGraphSchema := os.Getenv("APOLLO_SUB_GRAPH_SCHEMA")
	subGraphName := os.Getenv("APOLLO_SUB_GRAPH_NAME")
	subGraphURL := os.Getenv("APOLLO_SUB_GRAPH_URL")

	ctx := context.Background()
	client, err := apollostudio.NewClient(
		apollostudio.ClientOpts{
			APIKey:   apiKey,
			GraphRef: graphRef,
		},
	)

	if err != nil {
		handleErr(err)
	}

	b, err := client.GetLatestSchemaBuild(ctx)

	if err != nil {
		handleErr(err)
	}

	for _, v := range b.Result.BuildFailure.ErrorMessages {
		fmt.Println(v.Code, v.Message)
	}

	vr, err := client.ValidateSubGraph(
		ctx, &apollostudio.ValidateOptions{
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

	sr, err := client.SubmitSubGraph(
		ctx, &apollostudio.SubmitOptions{
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

	rr, err := client.GetSubGraph(ctx, subGraphName)

	if err != nil {
		handleErr(err)
	}

	fmt.Println("schema read")
	fmt.Println(rr.Revision)

	err = client.RemoveSubGraph(ctx, subGraphName)

	if err != nil {
		handleErr(err)
	}

	fmt.Println("schema deleted")
}
