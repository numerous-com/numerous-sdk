# Integration Tests

This directory contains integration tests for the Numerous SDK that test actual API communication and end-to-end task execution flows.

## Requirements

These tests require:

1. **API Availability**: The Numerous platform API must be available (by default at localhost:8080)
2. **Environment Variables**: Proper configuration via environment variables
3. **Integration Flag**: Tests must be run with the `--integration` flag

## Environment Variables

The integration tests use the same environment variables as the Numerous SDK client:

- `NUMEROUS_API_URL`: API endpoint URL (defaults to localhost:8080 if not set)
- `NUMEROUS_TOKEN`: Authentication token (if required)
- Other client configuration variables as needed

## Running Integration Tests

### Run All Integration Tests
```bash
pytest tests/integration/ --integration
```

### Run Specific Integration Test Files
```bash
pytest tests/integration/test_remote_handler_integration.py --integration
pytest tests/integration/test_task_execution_integration.py --integration
```

### Run Integration Tests with Verbose Output
```bash
pytest tests/integration/ --integration -v
```

### Run Integration Tests with Live Log Output
```bash
pytest tests/integration/ --integration -s --log-cli-level=INFO
```

## Test Structure

### test_remote_handler_integration.py
Tests the `RemoteTaskControlHandler` functionality:
- Handler initialization and connection
- Logging with API communication
- Progress and status updates
- Error handling and fallback modes
- Concurrent operations
- Stop request handling

### test_task_execution_integration.py
Tests end-to-end task execution with API persistence:
- Simple task execution with remote logging
- Data processing tasks with detailed progress tracking
- Task execution in fallback mode when API is unavailable

## Test Behavior

### With API Available
- Tests will communicate with the actual API
- Logs, progress updates, and status changes will be sent to the API
- Real session IDs and task instance IDs will be used
- All integration test functionality will be validated

### Without API Available
- **Integration tests will FAIL immediately** when RemoteTaskControlHandler initialization fails
- RemoteTaskControlHandler requires API connection and will raise RuntimeError if connection fails
- No fallback behavior exists - users must explicitly choose local backends for local execution
- Clear error messages guide users to ensure API availability or use different task control handlers

## Skipping Integration Tests

If the `--integration` flag is not provided, all integration tests will be automatically skipped with a clear message explaining the requirements.

## CI/CD Considerations

These integration tests can be run in CI/CD pipelines by:

1. Setting up the Numerous platform API in the CI environment
2. Configuring the required environment variables
3. Running tests with the `--integration` flag

Example CI configuration:
```yaml
test-integration:
  steps:
    - name: Setup API
      # Steps to start API service
    - name: Run Integration Tests
      run: pytest tests/integration/ --integration
  environment:
    NUMEROUS_API_URL: http://localhost:8080
```

## Debugging Integration Tests

To debug integration test failures:

1. **Check API Availability**: Ensure the API is running and accessible
2. **Verify Environment Variables**: Check that all required variables are set
3. **Run with Verbose Output**: Use `-v` and `-s` flags for detailed output
4. **Check Logs**: Integration tests include detailed logging for troubleshooting

Example debug run:
```bash
pytest tests/integration/test_remote_handler_integration.py::TestRemoteHandlerIntegration::test_remote_handler_log_with_api_connection --integration -v -s --log-cli-level=DEBUG
``` 