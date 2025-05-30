package deploy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/hasura/go-graphql-client/pkg/jsonutil"
	"numerous.com/cli/internal/archive"
	"numerous.com/cli/internal/output"
)

// GraphQLID type for GraphQL ID fields
type GraphQLID string

func (GraphQLID) GetGraphQLType() string {
	return "ID"
}

// GraphQLClient for making GraphQL requests
type GraphQLClient struct {
	endpoint     string
	token        string
	client       *http.Client
	subscription *graphql.SubscriptionClient
}

// NewGraphQLClient creates a new GraphQL client
func NewGraphQLClient(endpoint, token string) *GraphQLClient {
	// Create WebSocket URL for subscriptions
	wsURL := strings.Replace(endpoint, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)

	headers := map[string]interface{}{}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}

	subscriptionClient := graphql.NewSubscriptionClient(wsURL).
		WithConnectionParams(headers).
		WithSyncMode(true)

	return &GraphQLClient{
		endpoint:     endpoint,
		token:        token,
		client:       &http.Client{},
		subscription: subscriptionClient,
	}
}

// Input types
type CreateTaskCollectionInput struct {
	OrganizationSlug string                 `json:"organizationSlug"`
	Name             string                 `json:"name"`
	Version          string                 `json:"version"`
	Description      *string                `json:"description,omitempty"`
	Environment      TaskEnvironmentInput   `json:"environment"`
	Tasks            []*TaskDefinitionInput `json:"tasks"`
}

type DeployTaskCollectionInput struct {
	TaskCollectionID string `json:"taskCollectionId"`
}

type TaskEnvironmentInput struct {
	Type   string             `json:"type"`
	Python *PythonEnvironment `json:"python,omitempty"`
	Docker *DockerEnvironment `json:"docker,omitempty"`
}

type PythonEnvironment struct {
	Version          string  `json:"version"`
	RequirementsFile *string `json:"requirementsFile,omitempty"`
}

type DockerEnvironment struct {
	Dockerfile string  `json:"dockerfile"`
	Context    *string `json:"context,omitempty"`
}

type TaskDefinitionInput struct {
	FunctionName      string   `json:"functionName"`
	SourceFile        *string  `json:"sourceFile,omitempty"`
	Entrypoint        []string `json:"entrypoint"`
	APIEndpoint       *string  `json:"apiEndpoint,omitempty"`
	PythonStub        *string  `json:"pythonStub,omitempty"`
	Description       *string  `json:"description,omitempty"`
	DecoratedFunction *string  `json:"decoratedFunction,omitempty"`
}

// Response types
type CreateTaskCollectionResponse struct {
	CreateTaskCollection struct {
		Success        bool    `json:"success"`
		Error          *string `json:"error"`
		TaskCollection *struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"taskCollection"`
	} `json:"createTaskCollection"`
}

type TaskCollectionUploadURLResponse struct {
	TaskCollectionUploadURL struct {
		URL string `json:"url"`
	} `json:"taskCollectionUploadURL"`
}

type DeployTaskCollectionResponse struct {
	DeployTaskCollection struct {
		Success        bool    `json:"success"`
		Error          *string `json:"error"`
		TaskCollection *struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"taskCollection"`
	} `json:"deployTaskCollection"`
}

// Task deployment event types
type TaskBuildMessageEvent struct {
	Message string
}

type TaskBuildErrorEvent struct {
	Message string
}

type TaskBuildSuccessEvent struct {
	Message string
}

type TaskDeployEventsInput struct {
	TaskCollectionID string
	Handler          func(TaskDeployEvent) error
}

var ErrNoTaskDeployEventsHandler = errors.New("no task deploy events handler defined")

type TaskDeployEvent struct {
	Typename     string                `graphql:"__typename"`
	BuildMessage TaskBuildMessageEvent `graphql:"... on TaskBuildMessageEvent"`
	BuildError   TaskBuildErrorEvent   `graphql:"... on TaskBuildErrorEvent"`
	BuildSuccess TaskBuildSuccessEvent `graphql:"... on TaskBuildSuccessEvent"`
}

type TaskDeployEventsSubscription struct {
	TaskDeployEvents TaskDeployEvent `graphql:"taskDeployEvents(taskCollectionId: $taskCollectionID)"`
}

// DeployTaskCollectionGraphQL deploys a task collection via GraphQL using the new multi-step process
func (c *GraphQLClient) DeployTaskCollectionGraphQL(ctx context.Context, input CreateTaskCollectionInput, sourceDir string, verbose bool) (*DeployTaskCollectionResponse, error) {
	// Step 1: Create task collection
	createTask := output.StartTask("Creating task collection")
	createResponse, err := c.createTaskCollection(ctx, input)
	if err != nil {
		createTask.Error()
		return nil, fmt.Errorf("failed to create task collection: %w", err)
	}

	if !createResponse.CreateTaskCollection.Success {
		createTask.Error()
		return nil, fmt.Errorf("task collection creation failed: %s", *createResponse.CreateTaskCollection.Error)
	}

	taskCollectionID := createResponse.CreateTaskCollection.TaskCollection.ID
	createTask.Done()

	// Step 2: Create source archive
	archiveTask := output.StartTask("Creating source archive")
	archivePath, err := c.createSourceArchive(sourceDir)
	if err != nil {
		archiveTask.Error()
		return nil, fmt.Errorf("failed to create source archive: %w", err)
	}
	defer os.Remove(archivePath) // Clean up
	archiveTask.Done()

	// Step 3: Get upload URL
	uploadURLTask := output.StartTask("Getting upload URL")
	uploadURLResponse, err := c.getTaskCollectionUploadURL(ctx, taskCollectionID)
	if err != nil {
		uploadURLTask.Error()
		return nil, fmt.Errorf("failed to get upload URL: %w", err)
	}
	uploadURLTask.Done()

	// Step 4: Upload source archive
	uploadTask := output.StartTask("Uploading source archive")
	if err := c.uploadSourceArchive(uploadURLResponse.TaskCollectionUploadURL.URL, archivePath); err != nil {
		uploadTask.Error()
		return nil, fmt.Errorf("failed to upload source archive: %w", err)
	}
	uploadTask.Done()

	// Step 5: Deploy task collection
	deployTask := output.StartTask("Deploying task collection")
	deployInput := DeployTaskCollectionInput{
		TaskCollectionID: taskCollectionID,
	}

	// Start streaming deployment events BEFORE deployment begins if verbose mode is enabled
	var streamingDone chan error
	if verbose {
		streamingDone = make(chan error, 1)
		go func() {
			streamingDone <- c.streamTaskDeployEvents(ctx, taskCollectionID, deployTask)
		}()
		// Give the subscription a moment to be established
		time.Sleep(100 * time.Millisecond)
	}

	deployResponse, err := c.deployTaskCollection(ctx, deployInput)
	if err != nil {
		deployTask.Error()
		return nil, fmt.Errorf("failed to deploy task collection: %w", err)
	}

	if !deployResponse.DeployTaskCollection.Success {
		deployTask.Error()
		return nil, fmt.Errorf("deployment failed: %s", *deployResponse.DeployTaskCollection.Error)
	}

	// Wait for event streaming to complete if it was started
	if verbose && streamingDone != nil {
		select {
		case streamErr := <-streamingDone:
			if streamErr != nil {
				// Don't fail the deployment if event streaming fails, just log it
				deployTask.AddLine("Warning", fmt.Sprintf("Event streaming ended with error: %v", streamErr))
			}
		case <-time.After(10 * time.Second):
			// Don't wait forever for streaming to finish
			deployTask.AddLine("Warning", "Event streaming timed out")
		}
	}

	deployTask.Done()

	return deployResponse, nil
}

// createTaskCollection creates a new task collection
func (c *GraphQLClient) createTaskCollection(ctx context.Context, input CreateTaskCollectionInput) (*CreateTaskCollectionResponse, error) {
	query := `
		mutation CreateTaskCollection($input: CreateTaskCollectionInput!) {
			createTaskCollection(input: $input) {
				success
				error
				taskCollection {
					id
					name
					status
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": input,
	}

	var response CreateTaskCollectionResponse
	if err := c.executeQuery(ctx, query, variables, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// getTaskCollectionUploadURL gets the upload URL for a task collection
func (c *GraphQLClient) getTaskCollectionUploadURL(ctx context.Context, taskCollectionID string) (*TaskCollectionUploadURLResponse, error) {
	query := `
		mutation TaskCollectionUploadURL($taskCollectionId: ID!) {
			taskCollectionUploadURL(taskCollectionId: $taskCollectionId) {
				url
			}
		}
	`

	variables := map[string]interface{}{
		"taskCollectionId": taskCollectionID,
	}

	var response TaskCollectionUploadURLResponse
	if err := c.executeQuery(ctx, query, variables, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// deployTaskCollection deploys a created task collection
func (c *GraphQLClient) deployTaskCollection(ctx context.Context, input DeployTaskCollectionInput) (*DeployTaskCollectionResponse, error) {
	query := `
		mutation DeployTaskCollection($input: DeployTaskCollectionInput!) {
			deployTaskCollection(input: $input) {
				success
				error
				taskCollection {
					id
					name
					status
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": input,
	}

	var response DeployTaskCollectionResponse
	if err := c.executeQuery(ctx, query, variables, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// createSourceArchive creates a tar archive of the source directory
func (c *GraphQLClient) createSourceArchive(sourceDir string) (string, error) {
	// Create unique archive name to prevent caching issues
	timestamp := time.Now().Unix()
	archivePath := filepath.Join(os.TempDir(), fmt.Sprintf("task-collection-source-%d.tar", timestamp))

	// Use the same archive creation logic as apps
	if err := archive.TarCreate(sourceDir, archivePath, []string{}); err != nil {
		return "", fmt.Errorf("failed to create tar archive: %w", err)
	}

	return archivePath, nil
}

// uploadSourceArchive uploads the source archive to the provided URL
func (c *GraphQLClient) uploadSourceArchive(uploadURL, archivePath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive file: %w", err)
	}
	defer file.Close()

	req, err := http.NewRequest("PUT", uploadURL, file)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	// Set appropriate headers
	req.Header.Set("Content-Type", "application/tar")

	// Get file size for Content-Length
	if stat, err := file.Stat(); err == nil {
		req.ContentLength = stat.Size()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// executeQuery sends a GraphQL query and decodes the response
func (c *GraphQLClient) executeQuery(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GraphQL request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var graphqlResponse struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &graphqlResponse); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(graphqlResponse.Errors) > 0 {
		return fmt.Errorf("GraphQL errors: %s", graphqlResponse.Errors[0].Message)
	}

	if err := json.Unmarshal(graphqlResponse.Data, result); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// convertTaskManifestToGraphQLInput converts task manifest to GraphQL input
func convertTaskManifestToGraphQLInput(manifest *TaskManifestCollection, orgSlug string) CreateTaskCollectionInput {
	input := CreateTaskCollectionInput{
		OrganizationSlug: orgSlug,
		Name:             manifest.Name,
		Version:          manifest.Version,
		Description:      &manifest.Description,
		Tasks:            make([]*TaskDefinitionInput, len(manifest.Task)),
	}

	// Set environment
	if manifest.Python != nil {
		input.Environment = TaskEnvironmentInput{
			Type: "PYTHON",
			Python: &PythonEnvironment{
				Version:          manifest.Python.Version,
				RequirementsFile: &manifest.Python.RequirementsFile,
			},
		}
	} else if manifest.Docker != nil {
		context := "."
		if manifest.Docker.Context != "" {
			context = manifest.Docker.Context
		}
		input.Environment = TaskEnvironmentInput{
			Type: "DOCKER",
			Docker: &DockerEnvironment{
				Dockerfile: manifest.Docker.Dockerfile,
				Context:    &context,
			},
		}
	}

	// Convert tasks
	for i, task := range manifest.Task {
		graphqlEntrypoint := task.Entrypoint
		if graphqlEntrypoint == nil {
			graphqlEntrypoint = []string{}
		}
		input.Tasks[i] = &TaskDefinitionInput{
			FunctionName:      task.FunctionName,
			SourceFile:        &task.SourceFile,
			Entrypoint:        graphqlEntrypoint,
			APIEndpoint:       &task.APIEndpoint,
			PythonStub:        &task.PythonStub,
			Description:       &task.Description,
			DecoratedFunction: &task.DecoratedFunction,
		}
	}

	return input
}

// streamTaskDeployEvents streams deployment events for verbose logging
func (c *GraphQLClient) streamTaskDeployEvents(ctx context.Context, taskCollectionID string, task *output.Task) error {
	// Use the TaskDeployEvents method for streaming
	eventsInput := TaskDeployEventsInput{
		TaskCollectionID: taskCollectionID,
		Handler: func(de TaskDeployEvent) error {
			switch de.Typename {
			case "TaskBuildMessageEvent":
				for _, l := range strings.Split(de.BuildMessage.Message, "\n") {
					task.AddLine("Build", l)
				}
			case "TaskBuildErrorEvent":
				for _, l := range strings.Split(de.BuildError.Message, "\n") {
					task.AddLine("Error", l)
				}
				// Don't fail deployment on build messages, just log them
			case "TaskBuildSuccessEvent":
				for _, l := range strings.Split(de.BuildSuccess.Message, "\n") {
					task.AddLine("Success", l)
				}
				// stop subscription after success
				return graphql.ErrSubscriptionStopped
			}
			// continue streaming for other events
			return nil
		},
	}

	return c.TaskDeployEvents(ctx, eventsInput)
}

// TaskDeployEvents streams task deployment events
func (c *GraphQLClient) TaskDeployEvents(ctx context.Context, input TaskDeployEventsInput) error {
	if input.Handler == nil {
		return ErrNoTaskDeployEventsHandler
	}
	defer c.subscription.Close()

	var handlerError error
	variables := map[string]any{"taskCollectionID": GraphQLID(input.TaskCollectionID)}
	handler := func(message []byte, err error) error {
		if err != nil {
			return err
		}

		var value TaskDeployEventsSubscription

		err = jsonutil.UnmarshalGraphQL(message, &value)
		if err != nil {
			return err
		}

		// clean value based on typename
		switch value.TaskDeployEvents.Typename {
		case "TaskBuildMessageEvent":
			value.TaskDeployEvents.BuildError = TaskBuildErrorEvent{}
			value.TaskDeployEvents.BuildSuccess = TaskBuildSuccessEvent{}
		case "TaskBuildErrorEvent":
			value.TaskDeployEvents.BuildMessage = TaskBuildMessageEvent{}
			value.TaskDeployEvents.BuildSuccess = TaskBuildSuccessEvent{}
		case "TaskBuildSuccessEvent":
			value.TaskDeployEvents.BuildMessage = TaskBuildMessageEvent{}
			value.TaskDeployEvents.BuildError = TaskBuildErrorEvent{}
		}

		// run handler
		handlerError = input.Handler(value.TaskDeployEvents)
		if handlerError != nil {
			return graphql.ErrSubscriptionStopped
		}

		return nil
	}

	_, err := c.subscription.Subscribe(&TaskDeployEventsSubscription{}, variables, handler, graphql.OperationName("CLITaskDeployEvents"))
	if err != nil {
		return err
	}

	err = c.subscription.Run()

	// first we check if the handler found any errors
	if handlerError != nil {
		return handlerError
	}

	return err
}
