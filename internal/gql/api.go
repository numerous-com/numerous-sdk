package gql

import "os"

// Variables overridden during build for production
var (
	httpURL string = "http://localhost:8080/query"
	wsURL   string = "ws://localhost:8080/query"
)

func GetHTTPURL() string {
	if url := os.Getenv("NUMEROUS_GRAPHQL_HTTP_URL"); url != "" {
		return url
	} else {
		return httpURL
	}
}

func GetWSURL() string {
	if url := os.Getenv("NUMEROUS_GRAPHQL_WS_URL"); url != "" {
		return url
	} else {
		return wsURL
	}
}
