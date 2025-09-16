package app

import (
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/hasura/go-graphql-client/pkg/jsonutil"
)

type WorkloadLogEntry struct {
	Timestamp time.Time
	Text      string
}

type TaskInstanceLogsInput struct {
	InstanceID string
	Tail       *int
	Follow     bool
}

type TaskInstanceLogsSubscription struct {
	TaskInstanceLogs WorkloadLogEntry `graphql:"taskInstanceLogs(input: {taskInstanceID: $taskInstanceID, tail: $tail, follow: $follow})"`
}

func (s *Service) TaskInstanceLogs(input TaskInstanceLogsInput) (chan WorkloadLogEntry, error) {
	ch := make(chan WorkloadLogEntry)

	handler := func(message []byte, err error) error {
		if err != nil {
			return err
		}

		var value TaskInstanceLogsSubscription

		err = jsonutil.UnmarshalGraphQL(message, &value)
		if err != nil {
			return err
		}

		ch <- value.TaskInstanceLogs

		return nil
	}

	vars := make(map[string]any)
	vars["taskInstanceID"] = graphql.ID(input.InstanceID)
	vars["tail"] = input.Tail
	vars["follow"] = input.Follow

	_, err := s.subscription.Subscribe(&TaskInstanceLogsSubscription{}, vars, handler, graphql.OperationName("CLITaskInstanceLogs"))
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)
		s.subscription.Run() // nolint:errcheck
	}()

	return ch, nil
}
