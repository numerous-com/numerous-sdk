package app

import (
	"context"
	"errors"

	"git.sr.ht/~emersion/gqlclient"
)

type appDeleteResponse struct {
	ToolDelete struct {
		Typename string `json:"__typename"`
		Result   string `json:"result"`
	}
}

func Delete(id string, client *gqlclient.Client) (*appDeleteResponse, error) {
	resp := appDeleteResponse{}
	op := createAppDeleteOperation(id)

	if err := client.Execute(context.TODO(), op, &resp); err != nil {
		return nil, err
	}

	if resp.ToolDelete.Typename == "ToolDeleteSuccess" || resp.ToolDelete.Typename == "ToolDeleteFailure" {
		return &resp, nil
	}

	return nil, errors.New("unexpected response from toolDelete mutation")
}

func createAppDeleteOperation(id string) *gqlclient.Operation {
	op := gqlclient.NewOperation(`
		mutation ToolDelete($id: ID!) {
			toolDelete(id: $id){
				__typename
				... on ToolDeleteSuccess {
					result
				}
				... on ToolDeleteFailure {
					result
				}
			}
		}
		`)
	op.Var("id", id)

	return op
}
