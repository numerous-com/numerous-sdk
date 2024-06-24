package app

import (
	"context"

	"git.sr.ht/~emersion/gqlclient"
)

type appPublishResponse struct {
	ToolPublish App
}

func Publish(id string, client *gqlclient.Client) (App, error) {
	resp := appPublishResponse{}
	op := createAppPublishOperation(id)

	if err := client.Execute(context.TODO(), op, &resp); err != nil {
		return resp.ToolPublish, err
	}

	return resp.ToolPublish, nil
}

func createAppPublishOperation(id string) *gqlclient.Operation {
	op := gqlclient.NewOperation(`
	mutation ToolPublish($id: ID!) {
		toolPublish(id: $id) {
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
