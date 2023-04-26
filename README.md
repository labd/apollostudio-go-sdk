# Amplience GO SDK

Go SDK for [Apollo Studio](https://studio.apollographql.com/).

## Development

To test the API, run `cmd/debug/main.go` file with your own Apollo studio graph credentials.

```
export APOLLO_API_KEY=your_api_key
export APOLLO_GRAPH_REF=your_graph_ref
export APOLLO_SUB_GRAPH_SCHEMA=your_sub_graph_schema
export APOLLO_SUB_GRAPH_NAME=your_sub_graph_name
export APOLLO_SUB_GRAPH_URL=your_sub_graph_url

go build ./cmd/debug
./debug -h
```

## Contributing

Apollo Studio GraphQL explorer can be found [here](https://studio.apollographql.com/public/apollo-platform/variant/main/explorer).