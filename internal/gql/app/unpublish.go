package app

import (
	"context"

	"git.sr.ht/~emersion/gqlclient"
)

type appUnpublishResponse struct {
	ToolUnpublish App
}

func Unpublish(id string, client *gqlclient.Client) (App, error) {
	resp := appUnpublishResponse{}
	op := createAppUnpublishOperation(id)

	if err := client.Execute(context.TODO(), op, &resp); err != nil {
		return resp.ToolUnpublish, err
	}

	return resp.ToolUnpublish, nil
}

func createAppUnpublishOperation(id string) *gqlclient.Operation {
	op := gqlclient.NewOperation(`
	mutation ToolUnpublish($id: ID!) {
		toolUnpublish(id: $id) {
			id
			name
			description
			createdAt
			sharedUrl
			publicUrl
		}
	}
	`)
	op.Var("id", id)

	return op
}
