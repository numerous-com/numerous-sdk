package jobs

import (
	"context"

	"git.sr.ht/~emersion/gqlclient"
)

func JobStop(id string, client *gqlclient.Client) (string, error) {
	resp := jobStopResponse{}
	op := createJobStopOperation(id)
	if err := client.Execute(context.TODO(), op, &resp); err != nil {
		return resp.Message, err
	}

	return resp.Message, nil
}

func createJobStopOperation(id string) *gqlclient.Operation {
	op := gqlclient.NewOperation(`
		mutation JobStop($id: ID!) {
			jobStop(id: $id) {
				message
			}
		}
	`)

	op.Var("id", id)

	return op
}
