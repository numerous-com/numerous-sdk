package jobs

import (
	"context"

	"git.sr.ht/~emersion/gqlclient"
)

func JobsByTool(id string, client *gqlclient.Client) ([]Job, error) {
	resp := jobsByToolResponse{}

	op := createJobsByToolOperation(id)
	if err := client.Execute(context.TODO(), op, &resp); err != nil {
		return resp.JobsByTool, err
	}

	return resp.JobsByTool, nil
}

func createJobsByToolOperation(id string) *gqlclient.Operation {
	op := gqlclient.NewOperation(`
		query JobsByTool($id: ID!) {
			jobsByTool(id: $id) {
				id
			}
		}
	`)

	op.Var("id", id)

	return op
}
