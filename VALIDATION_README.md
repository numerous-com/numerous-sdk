# Task Workflow Validation Script

This directory contains a comprehensive validation script that tests the complete Numerous task workflow from authentication to task execution.

## 🚀 Quick Start

For the fastest way to validate your setup, use the interactive quick validation script:

```bash
./quick-validate.sh
```

This will guide you through common validation scenarios:
- Full validation with interactive organization selection
- Quick testing with existing tokens
- Development mode with verbose output
- CI/CD integration modes

For direct access to the full validation script:

```bash
./validate-task-workflow.sh --org my-org
```

## Overview

The `validate-task-workflow.sh` script validates the entire task workflow by:

1. **Authentication** - Login to the Numerous platform
2. **Organization Listing** - List available organizations
3. **Task Deployment** - Deploy a task collection to the platform
4. **Task Execution** - Run tasks remotely from the deployed collection
5. **Local Testing** - Test local task execution
6. **Cleanup** - Clean up deployed resources (optional)

## Prerequisites

1. **Built CLI Binary**: The script requires the CLI to be built:
   ```bash
   go build -o bin/numerous .
   ```

2. **Example Task Collection**: The script uses the Python validator example:
   ```
   examples/validator/python-tasks/
   ```

3. **API Server**: A running Numerous API server (local or remote)

## Usage

### Basic Usage

```bash
# Interactive mode - will prompt for organization
./validate-task-workflow.sh

# Specify organization
./validate-task-workflow.sh --org my-org

# With verbose output
./validate-task-workflow.sh --org my-org --verbose
```

### Advanced Usage

```bash
# Skip login (use existing token)
./validate-task-workflow.sh --org my-org --skip-login

# Don't cleanup deployed resources
./validate-task-workflow.sh --org my-org --no-cleanup

# Use custom API URL
./validate-task-workflow.sh --org my-org --api-url https://api.numerous.com/graphql

# Full verbose run with no cleanup
./validate-task-workflow.sh --org my-org --verbose --no-cleanup
```

## Command Line Options

| Option | Description |
|--------|-------------|
| `--org SLUG` | Organization slug to use (required if not interactive) |
| `--api-url URL` | API endpoint URL (default: `$NUMEROUS_API_URL` or `http://localhost:8080/graphql`) |
| `--skip-login` | Skip login step (assume already authenticated) |
| `--no-cleanup` | Don't cleanup deployed resources |
| `--verbose` | Enable verbose output |
| `--help` | Show help message |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `NUMEROUS_API_URL` | API endpoint URL |
| `NUMEROUS_ACCESS_TOKEN` | Access token (when skipping login) |

## Workflow Steps

### 1. Pre-flight Checks
- Verifies CLI binary exists at `bin/numerous`
- Checks example task collection exists
- Validates task manifest file

### 2. Authentication
- Performs login via CLI (unless `--skip-login`)
- Checks existing token status
- Uses environment token if available

### 3. Organization Selection
- Lists available organizations
- Uses specified org or prompts user to select
- Validates organization access

### 4. Task Collection Deployment
- Deploys Python validator task collection
- Uses timestamp-based collection name for uniqueness
- Waits for deployment completion

### 5. Remote Task Testing
- Lists tasks in deployed collection
- Runs each task from the Python validator:
  - `validate_environment`
  - `process_data`
  - `file_operations`
  - `network_check`
- Reports success/failure for each task

### 6. Local Task Testing
- Tests local execution with same task collection
- Verifies local vs remote execution consistency

### 7. Cleanup (Optional)
- Removes deployed task collection
- Note: Cleanup functionality requires implementation

## Output

The script provides colored output with clear status indicators:

- 🔵 **Step indicators** - Current workflow step
- ✅ **Success messages** - Successful operations
- ⚠️ **Warnings** - Non-critical issues
- ❌ **Error messages** - Critical failures
- ℹ️ **Information** - Additional details

## Example Output

```
🚀 Starting Numerous Task Workflow Validation
=============================================

Configuration:
  CLI Binary: /path/to/bin/numerous
  Example Tasks: /path/to/examples/validator/python-tasks
  API URL: http://localhost:8080/graphql
  Collection Name: test-validation-1748531358
  Organization: my-org
  Verbose: true
  Skip Login: false
  Cleanup: true

🔵 Checking CLI binary
✅ CLI binary found
🔵 Checking example task collection
✅ Example task collection found
🔵 Performing login
✅ Login completed
🔵 Listing organizations
✅ Organizations retrieved
🔵 Deploying task collection
✅ Task collection deployed successfully
🔵 Listing tasks in deployed collection
✅ Tasks listed successfully
🔵 Running tasks from deployed collection
ℹ️ Running task: validate_environment
✅ Task 'validate_environment' completed successfully
...
🎉 Validation completed successfully!
```

## Exit Codes

- `0` - All validation steps completed successfully
- `1` - Validation failed (check output for details)

## Troubleshooting

### Common Issues

1. **CLI Binary Not Found**
   ```bash
   go build -o bin/numerous .
   ```

2. **Example Tasks Missing**
   - Ensure you're in the `numerous-sdk` directory
   - Check that `examples/validator/python-tasks/` exists

3. **Authentication Failed**
   - Check API URL is correct
   - Verify network connectivity
   - Use `--skip-login` with `NUMEROUS_ACCESS_TOKEN` if needed

4. **Organization Not Found**
   - Run `./bin/numerous organization list` to see available orgs
   - Check organization permissions

5. **Deployment Failed**
   - Verify task manifest syntax
   - Check organization permissions
   - Review API server logs

### Debug Mode

For detailed debugging, run with verbose output:

```bash
./validate-task-workflow.sh --org my-org --verbose
```

### Manual Cleanup

If automatic cleanup fails, manually remove deployed collections:

```bash
# List deployed collections
./bin/numerous task list --org my-org

# Note: Manual cleanup commands to be implemented
```

## Development

### Adding New Test Cases

To add new validation steps:

1. Create a new function following the pattern:
   ```bash
   my_new_test() {
       print_step "Testing new functionality"
       
       if $CLI_BINARY my_command $VERBOSE_FLAG; then
           print_success "Test passed"
       else
           print_error "Test failed"
           exit 1
       fi
   }
   ```

2. Add the function call to the `main()` function

3. Update this README with the new test description

### Testing the Script

Test the script itself without running full validation:

```bash
# Test help
./validate-task-workflow.sh --help

# Test pre-flight checks only
./validate-task-workflow.sh --org test --skip-login 2>&1 | head -20
```

## Integration with CI/CD

This script can be used in CI/CD pipelines:

```yaml
# Example GitHub Actions usage
- name: Run Task Workflow Validation
  run: |
    cd numerous-sdk
    go build -o bin/numerous .
    ./validate-task-workflow.sh --org ${{ secrets.TEST_ORG }} --skip-login
  env:
    NUMEROUS_ACCESS_TOKEN: ${{ secrets.API_TOKEN }}
    NUMEROUS_API_URL: https://api.numerous.com/graphql
```

This provides a comprehensive test of the entire Numerous task platform functionality. 