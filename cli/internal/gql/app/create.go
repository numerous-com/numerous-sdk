package app

import (
	"context"

	"numerous/cli/manifest"

	"git.sr.ht/~emersion/gqlclient"
)

type appCreateResponse struct {
	ToolCreate App
}

func Create(m *manifest.Manifest, client *gqlclient.Client) (App, error) {
	resp := appCreateResponse{}
	jsonManifest, err := m.ToJSON()
	if err != nil {
		return resp.ToolCreate, err
	}

	op := createAppCreateOperation(jsonManifest)
	if err := client.Execute(context.TODO(), op, &resp); err != nil {
		return resp.ToolCreate, err
	}

	return resp.ToolCreate, nil
}

func createAppCreateOperation(m string) *gqlclient.Operation {
	op := gqlclient.NewOperation(`
	mutation ToolCreate($userID: ID!, $manifest:String!) {
		toolCreate(input:{userId: $userID, manifest: $manifest}) {
			id
			name
			description
			createdAt
			sharedUrl
			publicUrl
		}
	}
	`)
	// userID variable hardcoded as we do not have users implemented yet
	op.Var("userID", "1")
	op.Var("manifest", m)

	return op
}
