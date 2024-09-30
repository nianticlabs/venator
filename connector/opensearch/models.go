package opensearch

// BulkQueryResponse defines a struct to match the structure of an OpenSearch bulk query (Publish) response.
type BulkQueryResponse struct {
	Took   int                                 `json:"took"`
	Errors bool                                `json:"errors"`
	Items  []map[string]map[string]interface{} `json:"items"`
}

// QueryResponse defines a struct to match the structure of an OpenSearch query response.
type QueryResponse struct {
	Schema   []map[string]string `json:"schema"`
	Datarows [][]interface{}     `json:"datarows"`
	Total    int                 `json:"total"`
	Size     int                 `json:"size"`
	Status   int                 `json:"status"`
}

type QueryErrorResponse struct {
	Error  QueryError `json:"error"`
	Status int        `json:"status"`
}

type QueryError struct {
	Reason  string `json:"reason"`
	Details string `json:"details"`
	Type    string `json:"type"`
}

type CreateReq struct {
	Index string `json:"_index"`
}

type BulkRequestOp struct {
	Create *CreateReq `json:"create,omitempty"`
}
