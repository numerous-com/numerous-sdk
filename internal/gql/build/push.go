package build

import (
	"context"
	"os"

	"numerous.com/cli/internal/gql/secret"

	"git.sr.ht/~emersion/gqlclient"
)

type pushResponse struct {
	BuildPush BuildConfiguration
}

func Push(file *os.File, appID string, client *gqlclient.Client, secrets map[string]string) (BuildConfiguration, error) {
	resp := pushResponse{}
	convertedSecrets := secret.AppSecretsFromMap(secrets)

	op := createBuildOperation(file, appID, convertedSecrets)

	if err := client.Execute(context.TODO(), op, &resp); err != nil {
		return BuildConfiguration{}, err
	}

	return resp.BuildPush, nil
}

type buildPushInput struct {
	Secrets []*secret.AppSecret `json:"secrets"`
}

func createBuildOperation(file *os.File, appID string, secrets []*secret.AppSecret) *gqlclient.Operation {
	op := gqlclient.NewOperation(`
		mutation BuildPush($file: Upload!, $appID: ID!, $input: BuildPushInput!) {
			buildPush(file: $file, id: $appID, input: $input) {
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
	op.Var("input", &buildPushInput{Secrets: secrets})

	return op
}
