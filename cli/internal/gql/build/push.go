package build

import (
	"context"
	"os"

	"git.sr.ht/~emersion/gqlclient"
)

type pushResponse struct {
	BuildPush BuildConfiguration
}

func Push(file *os.File, appID string, client *gqlclient.Client) (BuildConfiguration, error) {
	resp := pushResponse{}
	op := createBuildOperation(file, appID)

	if err := client.Execute(context.TODO(), op, &resp); err != nil {
		return BuildConfiguration{}, err
	}

	return resp.BuildPush, nil
}

func createBuildOperation(file *os.File, appID string) *gqlclient.Operation {
	op := gqlclient.NewOperation(`
		mutation BuildPush($file: Upload!, $appID: ID!) {
			buildPush(file: $file, id: $appID) {
				buildId
			}
		}
	`)
	op.Var("appID", appID)
	op.Var("file", gqlclient.Upload{
		Filename: file.Name(),
		MIMEType: "application/zip",
		Body:     file,
	})

	return op
}
