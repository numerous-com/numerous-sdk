# Numerous Tasks Validator Examples

This directory contains example task collections that demonstrate and validate the functionality of Numerous Tasks. These examples serve as both documentation and testing tools for the task execution system.

## Overview

The validator examples include:

- **Python Tasks** (`python-tasks/`): Demonstrates Python-based task collections
- **Docker Tasks** (`docker-tasks/`): Demonstrates Docker-based task collections with multiple execution modes

## Python Task Collection

The Python validator demonstrates:

- ✅ Environment validation (Python version, dependencies)
- 📊 Data processing with pandas/numpy
- 📁 File I/O operations
- 🌐 Network connectivity testing

### Usage

```bash
# List available tasks
numerous task list examples/validator/python-tasks

# Validate the task collection
numerous task validate examples/validator/python-tasks --verbose

# Run individual tasks
numerous task run validate_environment --task-dir examples/validator/python-tasks
numerous task run process_data --task-dir examples/validator/python-tasks
numerous task run file_operations --task-dir examples/validator/python-tasks
numerous task run network_check --task-dir examples/validator/python-tasks
```

## Docker Task Collection

The Docker validator demonstrates:

- 🐳 Container environment validation
- 📊 Data processing in containers
- 🖥️ System information gathering via shell scripts
- 🌐 HTTP API endpoints for task execution

### Task Execution Modes

1. **Entrypoint Tasks**: Execute Python scripts or shell commands
2. **API Endpoint Tasks**: HTTP endpoints for web-based task execution

### Usage

```bash
# List available tasks
numerous task list examples/validator/docker-tasks

# Validate the task collection
numerous task validate examples/validator/docker-tasks --verbose

# Run tasks in Docker containers (requires Docker)
numerous task run validate_container --task-dir examples/validator/docker-tasks
numerous task run system_info --task-dir examples/validator/docker-tasks

# Run tasks locally without Docker
numerous task run validate_container --task-dir examples/validator/docker-tasks --no-docker
numerous task run system_info --task-dir examples/validator/docker-tasks --no-docker
```

## Task Collection Structure

### Python Tasks (`python-tasks/`)

```
python-tasks/
├── numerous-task.toml    # Task collection manifest
├── requirements.txt      # Python dependencies
├── tasks/
│   ├── __init__.py      # Python package marker
│   └── validator.py     # Task implementations
└── stubs/
    └── validator.pyi    # Type hints for IDE support
```

### Docker Tasks (`docker-tasks/`)

```
docker-tasks/
├── numerous-task.toml    # Task collection manifest
├── Dockerfile           # Container definition
├── requirements.txt     # Python dependencies
├── app.py              # Flask web server for API endpoints
└── tasks/
    ├── validator.py    # Python task implementations
    └── entrypoint.sh   # Shell script tasks
```

## Validation Features

The validator examples test:

- ✅ **Manifest Parsing**: TOML syntax and required fields
- ✅ **Environment Configuration**: Python/Docker setup validation
- ✅ **File References**: Source files, requirements, Dockerfiles
- ✅ **Task Definitions**: Function names, entrypoints, API endpoints
- ✅ **Execution Modes**: Direct Python, Docker containers, local fallback
- ✅ **Error Handling**: Missing files, invalid configurations
- ✅ **Output Formats**: Text and JSON output options

## Development Workflow

Use these examples to:

1. **Test New Features**: Add new validation tasks as features are developed
2. **Validate Changes**: Run validators after making changes to the task system
3. **Debug Issues**: Use verbose output to troubleshoot problems
4. **Document Patterns**: Show best practices for task collection structure

## Example Commands

```bash
# Quick validation of both examples
numerous task validate examples/validator/python-tasks
numerous task validate examples/validator/docker-tasks

# Run all Python validation tasks
for task in validate_environment process_data file_operations network_check; do
  numerous task run $task --task-dir examples/validator/python-tasks
done

# Test Docker tasks locally (no Docker required)
numerous task run validate_container --task-dir examples/validator/docker-tasks --no-docker
numerous task run system_info --task-dir examples/validator/docker-tasks --no-docker

# JSON output for programmatic use
numerous task run validate_environment --task-dir examples/validator/python-tasks --output json
```

## Contributing

When adding new task collection features:

1. Add corresponding validation tasks to these examples
2. Update the manifest files with new configuration options
3. Test both Python and Docker execution modes
4. Ensure validation passes with `numerous task validate`
5. Document new patterns in this README 

## Task Deployment Improvements

### Build Event Streaming

Task deployments now support verbose build output similar to app deployments. Use the `--verbose` or `-v` flag to see detailed build progress:

```bash
numerous deploy --verbose --organization your-org python-tasks/
```

This will show:
- Source archive creation progress
- Upload status updates  
- Build step execution
- Docker image build and push status
- Deployment completion confirmation

### Code Update Handling

The deployment system now includes improved cache invalidation to ensure your code changes are always reflected in new deployments:

- **Unique Archive Names**: Each deployment creates a uniquely timestamped source archive
- **Build Args**: Build timestamp is passed as a Docker build argument to invalidate Docker layer cache
- **NoCache Option**: Docker builds use `NoCache: true` to force rebuilds
- **Exclusion Patterns**: Common development files are automatically excluded from the build context

### Usage Examples

**Basic deployment:**
```bash
numerous deploy --organization your-org python-tasks/
```

**Verbose deployment with detailed output:**
```bash
numerous deploy --verbose --organization your-org python-tasks/
```

**Dry run to see what would be deployed:**
```bash
numerous deploy --dry-run --organization your-org python-tasks/
```

### Troubleshooting

If you're still not seeing code updates after deployment:

1. **Check the deployment logs**: Use `--verbose` to see detailed build output
2. **Verify file exclusions**: Make sure your source files aren't being excluded
3. **Check Docker cache**: The system should automatically invalidate cache, but you can verify in the build logs
4. **Validate manifest**: Use `--dry-run` to check your task configuration

### Build Process Details

The task deployment process now follows these steps:

1. **Create Task Collection**: Register the collection metadata
2. **Create Source Archive**: Package source files with timestamp-based naming
3. **Get Upload URL**: Obtain a secure upload endpoint
4. **Upload Source Archive**: Transfer the packaged source to the platform  
5. **Deploy with Streaming**: Start the build process with real-time event streaming (when verbose mode is enabled)

Each step provides feedback and the verbose mode shows detailed progress for the build and deployment phases. 