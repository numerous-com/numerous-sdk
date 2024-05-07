package build

import (
	"context"
	"encoding/base64"
	"os"

	"git.sr.ht/~emersion/gqlclient"
)

type pushResponse struct {
	BuildPush BuildConfiguration
}

func Push(file *os.File, appID string, client *gqlclient.Client, secrets map[string]string) (BuildConfiguration, error) {
	resp := pushResponse{}
	convertedSecrets := appSecretsFromMap(secrets)

	op := createBuildOperation(file, appID, convertedSecrets)

	if err := client.Execute(context.TODO(), op, &resp); err != nil {
		return BuildConfiguration{}, err
	}

	return resp.BuildPush, nil
}

func appSecretsFromMap(secrets map[string]string) []*appSecret {
	convertedSecrets := make([]*appSecret, 0)

	for name, value := range secrets {
		secret := &appSecret{
			Name:        name,
			Base64Value: base64.StdEncoding.EncodeToString([]byte(value)),
		}
		convertedSecrets = append(convertedSecrets, secret)
	}

	return convertedSecrets
}

type appSecret struct {
	Name        string `json:"name"`
	Base64Value string `json:"base64Value"`
}

type buildPushInput struct {
	Secrets []*appSecret `json:"secrets"`
}

func createBuildOperation(file *os.File, appID string, secrets []*appSecret) *gqlclient.Operation {
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
