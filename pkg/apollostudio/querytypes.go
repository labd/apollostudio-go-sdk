package apollostudio

// types in this file all need to be Public, probably because the hasura
// GraphQL client uses reflection to get the type names and map them against
// query result data? As a result these types are public, but maybe we don't
// want to expose them from our SDK?

type HistoricQueryParametersInput struct {
	ExcludedClients               *string `json:"excludedClients"`
	ExcludedOperationNames        *string `json:"excludedOperationNames"`
	From                          *string `json:"from"`
	IgnoredOperations             *string `json:"ignoredOperations"`
	IncludedVariants              *string `json:"includedVariants"`
	QueryCountThreshold           *string `json:"queryCountThreshold"`
	QueryCountThresholdPercentage *string `json:"queryCountThresholdPercentage"`
	To                            *string `json:"to"`
}

type GitContextInput struct {
	Branch    *string `json:"branch"`
	Commit    *string `json:"commit"`
	Committer *string `json:"committer"`
	Message   *string `json:"message"`
	RemoteUrl *string `json:"remoteUrl"`
}

type SubgraphCheckAsyncInput struct {
	Config                HistoricQueryParametersInput `json:"config"`
	GitContext            GitContextInput              `json:"gitContext"`
	GraphRef              string                       `json:"graphRef"`
	IntrospectionEndpoint *string                      `json:"introspectionEndpoint"`
	IsSandbox             bool                         `json:"isSandbox"`
	ProposedSchema        string                       `json:"proposedSchema"`
	SubgraphName          string                       `json:"subgraphName"`
}

type PartialSchemaInput struct {
	Sdl string `json:"sdl"`
}
