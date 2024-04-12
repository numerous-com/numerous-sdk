package app

import (
	"context"

	"git.sr.ht/~emersion/gqlclient"
)

type appResponse struct {
	Tool App
}

func Query(appID string, client *gqlclient.Client) (App, error) {
	resp := appResponse{}

	op := queryAppOperation(appID)
	if err := client.Execute(context.Background(), op, &resp); err != nil {
		return resp.Tool, err
	}

	return resp.Tool, nil
}

func queryAppOperation(appID string) *gqlclient.Operation {
	op := gqlclient.NewOperation(`
	query ReadTool($id: ID!) {
		tool(id: $id) {
			id
			name
			description
			createdAt
			sharedUrl
			publicUrl
		}
	}
	`)
	op.Var("id", appID)

	return op
}
