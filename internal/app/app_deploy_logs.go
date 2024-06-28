package app

import (
	"time"

	"numerous.com/cli/internal/appident"

	"github.com/hasura/go-graphql-client/pkg/jsonutil"
)

type AppDeployLogEntry struct {
	Timestamp time.Time
	Text      string
}

type AppDeployLogsSubscription struct {
	AppDeployLogs AppDeployLogEntry `graphql:"appDeployLogs(input: {organizationSlug: $orgSlug, appSlug: $appSlug})"`
}

func (s *Service) AppDeployLogs(ai appident.AppIdentifier) (chan AppDeployLogEntry, error) {
	ch := make(chan AppDeployLogEntry)

	handler := func(message []byte, err error) error {
		if err != nil {
			return err
		}

		var value AppDeployLogsSubscription

		err = jsonutil.UnmarshalGraphQL(message, &value)
		if err != nil {
			return err
		}

		ch <- value.AppDeployLogs

		return nil
	}

	vars := make(map[string]any)
	vars["orgSlug"] = ai.OrganizationSlug
	vars["appSlug"] = ai.AppSlug
	_, err := s.subscription.Subscribe(&AppDeployLogsSubscription{}, vars, handler)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)
		s.subscription.Run() // nolint:errcheck
	}()

	return ch, nil
}
