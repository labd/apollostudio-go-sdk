# Apollo Studio Go SDK

Go SDK for [Apollo Studio](https://studio.apollographql.com/).

## Installation

```bash 
go get github.com/labd/apollostudio-go-sdk
```

## Usage

```go
package main

import (
	"context"
	"github.com/labd/apollostudio-go-sdk/apollostudio"
)

func main() {
	key := "your-api-key"
	ref := "your-schema-reference"

	client, err := apollostudio.NewClient(key, ref)
	if err != nil {
		panic(err)
	}

	_, _ := client.SubmitSubGraph(
		context.Background(),
		&apollostudio.SubmitOptions{
			SubGraphSchema: []byte("schema { query: Query } type Query { hello: String }"),
			SubGraphName:   "my-subgraph",
			SubGraphURL:    "https://my-subgraph.com/graphql",
		},
	)
}
```

The client allows for several additional options to be set, which can extend its functionality.

```go
var clientOpts = []apollostudio.ClientOpt{
    apollostudio.WithHttpClient(http.DefaultClient),
    apollostudio.WithDebug(true),
    apollostudio.WithUrl("https://studio.apollographql.com/api/graphql"),                       
}

client, err := apollostudio.NewClient(key, ref, clientOpts...)

```

## Contributing

Apollo Studio GraphQL explorer can be found [here](https://studio.apollographql.com/public/apollo-platform/variant/main/explorer).
