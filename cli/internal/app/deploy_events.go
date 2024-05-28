package app

import (
	"context"
	"errors"

	"github.com/hasura/go-graphql-client"
	"github.com/hasura/go-graphql-client/pkg/jsonutil"
)

type DeployEvent struct {
	Message string
}

type DeployEventsInput struct {
	DeploymentVersionID string
	Handler             func(DeployEvent) bool
}

var ErrNoDeployEventsHandler = errors.New("no deploy events handler defined")

type DeployEventMessage struct {
	Typename string `json:"__typename"`
	Message  string
}

type DeployEventsSubscription struct {
	AppDeployEvents DeployEvent `graphql:"appDeployEvents(appDeploymentVersionID: $deployVersionID)"`
}

type GraphQLID string

func (GraphQLID) GetGraphQLType() string {
	return "ID"
}

func (s *Service) DeployEvents(ctx context.Context, input DeployEventsInput) error {
	if input.Handler == nil {
		return ErrNoDeployEventsHandler
	}

	variables := map[string]any{"deployVersionID": GraphQLID(input.DeploymentVersionID)}
	_, err := s.subscription.Subscribe(&DeployEventsSubscription{}, variables, func(message []byte, err error) error {
		if err != nil {
			return err
		}

		var value DeployEventsSubscription

		err = jsonutil.UnmarshalGraphQL(message, &value)
		if err != nil {
			return err
		}

		ok := input.Handler(DeployEvent{Message: value.AppDeployEvents.Message})
		if !ok {
			return graphql.ErrSubscriptionStopped
		}

		return nil
	})
	if err != nil {
		return nil
	}

	err = s.subscription.Run()
	if err != nil {
		return err
	}

	return nil
}
