package logs

import (
	"context"
	"errors"
	"testing"
	"time"

	"numerous.com/cli/internal/app"

	"github.com/stretchr/testify/assert"
)

func dummyTaskLogPrinter(entry app.WorkloadLogEntry) {}

func TestTaskLogs(t *testing.T) {
	const instanceID = "test-task-instance-id"
	testError := errors.New("test error")

	t.Run("calls service expected parameters", func(t *testing.T) {
		closedCh := make(chan app.WorkloadLogEntry)
		close(closedCh)
		service := &TaskLogsServiceMock{}

		tail := 100
		expectedInput := app.TaskInstanceLogsInput{
			InstanceID: instanceID,
			Tail:       &tail,
			Follow:     false,
		}
		service.On("TaskInstanceLogs", expectedInput).Return(closedCh, nil)

		input := taskLogsInput{
			instanceID: instanceID,
			tail:       tail,
			follow:     false,
			printer:    dummyTaskLogPrinter,
		}
		err := taskLogs(context.TODO(), service, input)

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})

	t.Run("it stops when context is cancelled", func(t *testing.T) {
		ch := make(chan app.WorkloadLogEntry)
		service := &TaskLogsServiceMock{}

		expectedInput := app.TaskInstanceLogsInput{
			InstanceID: instanceID,
			Tail:       nil,
			Follow:     true,
		}
		service.On("TaskInstanceLogs", expectedInput).Return(ch, nil)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(time.Millisecond * 10)
			cancel()
		}()

		input := taskLogsInput{
			instanceID: instanceID,
			tail:       0,
			follow:     true,
			printer:    dummyTaskLogPrinter,
		}
		err := taskLogs(ctx, service, input)

		assert.NoError(t, err)
	})

	t.Run("given service returns error, it returns the error", func(t *testing.T) {
		var nilChan chan app.WorkloadLogEntry = nil
		service := &TaskLogsServiceMock{}

		expectedInput := app.TaskInstanceLogsInput{
			InstanceID: instanceID,
			Tail:       nil,
			Follow:     true,
		}
		service.On("TaskInstanceLogs", expectedInput).Return(nilChan, testError)

		input := taskLogsInput{
			instanceID: instanceID,
			tail:       0,
			follow:     true,
			printer:    dummyTaskLogPrinter,
		}
		err := taskLogs(context.TODO(), service, input)

		assert.ErrorIs(t, err, testError)
	})

	t.Run("prints expected entries", func(t *testing.T) {
		ch := make(chan app.WorkloadLogEntry)
		service := &TaskLogsServiceMock{}

		expectedInput := app.TaskInstanceLogsInput{
			InstanceID: instanceID,
			Tail:       nil,
			Follow:     true,
		}
		service.On("TaskInstanceLogs", expectedInput).Return(ch, nil)

		entry1 := app.WorkloadLogEntry{Timestamp: time.Date(2024, time.March, 1, 1, 1, 1, 1, time.UTC), Text: "Task started"}
		entry2 := app.WorkloadLogEntry{Timestamp: time.Date(2024, time.March, 1, 2, 2, 2, 2, time.UTC), Text: "Task processing"}
		expected := []app.WorkloadLogEntry{entry1, entry2}
		actual := []app.WorkloadLogEntry{}
		printer := func(e app.WorkloadLogEntry) {
			actual = append(actual, e)
		}

		go func() {
			defer close(ch)
			ch <- entry1
			ch <- entry2
		}()

		input := taskLogsInput{
			instanceID: instanceID,
			tail:       0,
			follow:     true,
			printer:    printer,
		}
		err := taskLogs(context.TODO(), service, input)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("handles channel close gracefully", func(t *testing.T) {
		ch := make(chan app.WorkloadLogEntry)
		service := &TaskLogsServiceMock{}

		expectedInput := app.TaskInstanceLogsInput{
			InstanceID: instanceID,
			Tail:       nil,
			Follow:     false,
		}
		service.On("TaskInstanceLogs", expectedInput).Return(ch, nil)

		close(ch)

		input := taskLogsInput{
			instanceID: instanceID,
			tail:       0,
			follow:     false,
			printer:    dummyTaskLogPrinter,
		}
		err := taskLogs(context.TODO(), service, input)

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})
}
