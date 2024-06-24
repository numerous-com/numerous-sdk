package app

import (
	"context"

	"git.sr.ht/~emersion/gqlclient"
)

type appListResponse struct {
	Tools []App
}

func QueryList(client *gqlclient.Client) ([]App, error) {
	resp := appListResponse{[]App{}}

	op := queryAppListOperation()
	if err := client.Execute(context.Background(), op, &resp); err != nil {
		return resp.Tools, err
	}

	return resp.Tools, nil
}

func queryAppListOperation() *gqlclient.Operation {
	op := gqlclient.NewOperation(`
	query ReadTools {
		tools {
			id
			name
			description
			createdAt
			sharedUrl
			publicUrl
		}
	}
	`)

	return op
}
