package app

import (
	"context"
	"errors"

	"github.com/hasura/go-graphql-client"
	"github.com/hasura/go-graphql-client/pkg/jsonutil"
)

type AppBuildMessageEvent struct {
	Message string
}

type AppBuildErrorEvent struct {
	Message string
}

type AppDeploymentStatusEvent struct {
	Status string
}

type DeployEventsInput struct {
	DeploymentVersionID string
	Handler             func(DeployEvent) error
}

var ErrNoDeployEventsHandler = errors.New("no deploy events handler defined")

type DeployEvent struct {
	Typename         string                   `graphql:"__typename"`
	DeploymentStatus AppDeploymentStatusEvent `graphql:"... on AppDeploymentStatusEvent"`
	BuildMessage     AppBuildMessageEvent     `graphql:"... on AppBuildMessageEvent"`
	BuildError       AppBuildErrorEvent       `graphql:"... on AppBuildErrorEvent"`
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
	defer s.subscription.Close()

	var handlerError error
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

		// clean value
		switch value.AppDeployEvents.Typename {
		case "AppBuildMessageEvent":
			value.AppDeployEvents.BuildError = AppBuildErrorEvent{}
		case "AppBuildErrorEvent":
			value.AppDeployEvents.BuildMessage = AppBuildMessageEvent{}
		}

		// run handler
		handlerError = input.Handler(value.AppDeployEvents)
		if handlerError != nil {
			return graphql.ErrSubscriptionStopped
		}

		return nil
	})
	if err != nil {
		return nil
	}

	err = s.subscription.Run()

	// first we check if the handler found any errors
	if handlerError != nil {
		return handlerError
	}

	return err
}
