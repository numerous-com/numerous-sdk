# Tasks

With Numerous Tasks, you can define, execute, and manage long-running Python functions as distributed tasks.
Tasks provide a powerful way to handle asynchronous operations, background processing, and scalable workloads with seamless integration across development, testing, and production environments.

## Key Features

1. **Define tasks** using the simple `@task` decorator on Python functions
2. **Execute tasks** locally for development or remotely on the Numerous platform
3. **Monitor progress** with real-time status updates, progress tracking, and structured logging
4. **Manage sessions** to coordinate multiple related tasks with concurrency control
5. **Handle cancellation** and error recovery with graceful shutdown patterns
6. **Integrate seamlessly** with FastAPI, Streamlit, and other Python frameworks
7. **Deploy easily** using the Numerous CLI for production workloads

!!! tip
    Remember to add `numerous` as a dependency in your project; most likely to your `requirements.txt` file.

## Quick Start

Import the [Numerous SDK](http://www.pypi.org/project/numerous) in your Python code:

```python
import time
from numerous.tasks import task, Session, TaskControl

# Define a simple task
@task
def process_data(data: list) -> dict:
    """Process a list of data items."""
    result = {"processed": len(data), "items": data}
    return result

# Define a task with progress tracking
@task
def long_running_task(tc: TaskControl, iterations: int) -> str:
    """Execute a long-running task with progress updates."""
    for i in range(iterations):
        # Check if task should stop
        if tc.should_stop:
            tc.log("Task was cancelled", "warning")
            break
            
        # Update progress
        progress = (i + 1) / iterations * 100
        tc.update_progress(progress, f"Processing item {i+1}/{iterations}")
        
        # Simulate work
        time.sleep(0.1)
    
    tc.log("Task completed successfully", "info")
    return f"Processed {iterations} items"

# Execute tasks within a session
with Session() as session:
    # Direct execution (for simple cases)
    result = process_data([1, 2, 3, 4, 5])
    print(result)  # {"processed": 5, "items": [1, 2, 3, 4, 5]}
    
    # Task instance execution (for advanced features)
    instance = long_running_task.instance()
    future = instance.start(10)
    
    # Get result (blocks until complete)
    result = future.result()
    print(result)  # "Processed 10 items"
```

## Development Workflows

Numerous Tasks support multiple execution modes to accommodate different stages of your development workflow:

### 1. Direct Execution (Development & Testing)

Perfect for rapid development and unit testing. Tasks execute synchronously as regular function calls:

```python
from numerous.tasks import task

@task
def validate_data(data: dict) -> bool:
    """Validate data structure."""
    return all(key in data for key in ['id', 'name', 'value'])

# Direct execution - no session required
# Great for development and testing
is_valid = validate_data({'id': 1, 'name': 'test', 'value': 42})
print(is_valid)  # True
```

**Use cases:**
- Unit testing task logic
- Interactive development in Jupyter notebooks
- Quick validation of task behavior

### 2. Local Task Instances (Integration Testing)

Execute tasks as instances with full TaskControl features while running locally:

```python
from numerous.tasks import task, Session, TaskControl

@task
def integration_test_task(tc: TaskControl, config: dict) -> dict:
    """Test task with full monitoring capabilities."""
    tc.log("Starting integration test", "info")
    tc.update_progress(50, "Validating configuration")
    
    # Your test logic here
    result = {"status": "success", "config": config}
    
    tc.update_progress(100, "Test completed")
    tc.log("Integration test finished", "info")
    return result

# Local execution with task control
with Session() as session:
    instance = integration_test_task.instance()
    future = instance.start({"env": "test"})
    result = future.result()
    print(result)
```

**Use cases:**
- Integration testing with monitoring
- Local development with progress tracking
- Testing cancellation and error handling

### 3. Remote Execution (Production)

Execute tasks on the Numerous platform with automatic scaling and resource management:

```python
import os

# Configure for remote execution
os.environ['NUMEROUS_TASK_BACKEND'] = 'remote'

# Same code works for remote execution
with Session() as session:
    instance = process_large_dataset.instance()
    future = instance.start(massive_dataset)
    result = future.result()  # Executes on Numerous platform
```

**Use cases:**
- Production workloads
- Resource-intensive tasks
- Scaling beyond local machine capabilities

## Framework Integration

### FastAPI Integration

Integrate tasks seamlessly with FastAPI applications for robust web APIs:

```python
from fastapi import FastAPI, Request, BackgroundTasks
from numerous.tasks import task, Session, TaskControl
from numerous.frameworks.fastapi import get_session

app = FastAPI()

@task
def process_upload(tc: TaskControl, file_data: bytes, filename: str) -> dict:
    """Process uploaded file in background."""
    tc.log(f"Processing file: {filename}", "info")
    tc.update_progress(50, "Analyzing file")
    
    # Your processing logic
    result = {"filename": filename, "size": len(file_data), "status": "processed"}
    
    tc.update_progress(100, "Processing complete")
    return result

@app.post("/upload")
async def upload_file(request: Request, file_data: bytes, filename: str):
    """Upload endpoint that triggers background processing."""
    # Get user session from request
    session = get_session(request)
    
    # Start background task
    with session:
        instance = process_upload.instance()
        future = instance.start(file_data, filename)
        
        return {
            "message": "File uploaded successfully",
            "task_id": instance.id,
            "status": "processing"
        }

@app.get("/status/{task_id}")
async def get_task_status(task_id: str):
    """Check task status."""
    # Implementation depends on your task tracking needs
    return {"task_id": task_id, "status": "completed"}
```

### Streamlit Integration

Create interactive data processing applications with Streamlit:

```python
import streamlit as st
from numerous.tasks import task, Session, TaskControl
from numerous.frameworks.streamlit import get_session

@task
def analyze_data(tc: TaskControl, data: list, analysis_type: str) -> dict:
    """Analyze data with progress updates."""
    tc.log(f"Starting {analysis_type} analysis", "info")
    
    results = {}
    total_steps = len(data)
    
    for i, item in enumerate(data):
        if tc.should_stop:
            break
            
        # Simulate analysis
        progress = (i + 1) / total_steps * 100
        tc.update_progress(progress, f"Analyzing item {i+1}/{total_steps}")
        
        # Add to results
        results[f"item_{i}"] = {"value": item, "analysis": analysis_type}
    
    tc.log("Analysis completed", "info")
    return results

# Streamlit app
st.title("Data Analysis with Tasks")

# Get user session
session = get_session()

# UI elements
data_input = st.text_area("Enter data (one item per line)")
analysis_type = st.selectbox("Analysis Type", ["statistical", "qualitative"])

if st.button("Start Analysis"):
    if data_input:
        data = [line.strip() for line in data_input.split('\n') if line.strip()]
        
        with session:
            instance = analyze_data.instance()
            future = instance.start(data, analysis_type)
            
            # Show progress
            progress_bar = st.progress(0)
            status_text = st.empty()
            
            # In a real app, you'd want to poll for progress
            # This is a simplified example
            with st.spinner("Processing..."):
                result = future.result()
                
            st.success("Analysis completed!")
            st.json(result)
```

## CLI Integration and Deployment

### Local Development with CLI

Use the Numerous CLI to manage your task-based applications:

```bash
# Initialize a new project
numerous init

# Configure your numerous.toml
```

```toml
# numerous.toml
name = "Task Processing App"
description = "An application that processes data using Numerous Tasks"
port = 8000

[python]
version = "3.11"
library = "fastapi"
app_file = "main.py"
requirements_file = "requirements.txt"

[deploy]
app = "my-task-app"
organization = "my-org-slug"
```

### Deployment Workflow

1. **Develop locally** using direct execution and local task instances
2. **Test integration** with your chosen framework (FastAPI, Streamlit)
3. **Deploy to production** using the Numerous CLI

```bash
# Deploy your application
numerous deploy

# Monitor logs
numerous logs

# Check deployment status
numerous list
```

### Environment Configuration

Configure task execution based on your environment:

```python
import os
from numerous.tasks import task, Session, TaskControl

# Configure backend based on environment
if os.getenv('ENVIRONMENT') == 'production':
    os.environ['NUMEROUS_TASK_BACKEND'] = 'remote'
else:
    os.environ['NUMEROUS_TASK_BACKEND'] = 'local'

@task
def environment_aware_task(tc: TaskControl, data: dict) -> dict:
    """Task that adapts to different environments."""
    env = os.getenv('ENVIRONMENT', 'development')
    tc.log(f"Running in {env} environment", "info")
    
    if env == 'production':
        # Use production-specific logic
        tc.log("Using production optimizations", "info")
    else:
        # Use development-specific logic
        tc.log("Using development settings", "info")
    
    return {"environment": env, "data": data}
```

## Session Management

Sessions provide context for managing related tasks with built-in concurrency control:

```python
from numerous.tasks import Session, task, TaskControl

@task(max_parallel=3)
def batch_processor(tc: TaskControl, batch_id: str, items: list) -> dict:
    """Process a batch of items."""
    tc.log(f"Processing batch {batch_id}", "info")
    
    results = []
    for i, item in enumerate(items):
        if tc.should_stop:
            break
            
        # Process item
        result = {"item": item, "processed": True}
        results.append(result)
        
        # Update progress
        progress = (i + 1) / len(items) * 100
        tc.update_progress(progress, f"Batch {batch_id}: {i+1}/{len(items)}")
    
    return {"batch_id": batch_id, "results": results}

# Process multiple batches in a coordinated session
with Session(name="batch-processing-session") as session:
    futures = []
    
    # Start multiple batches (respecting max_parallel=3)
    for i, batch_data in enumerate(dataset_batches):
        instance = batch_processor.instance()
        future = instance.start(f"batch_{i}", batch_data)
        futures.append(future)
    
    # Collect results
    results = []
    for future in futures:
        result = future.result()
        results.append(result)
    
    print(f"Processed {len(results)} batches")
```

## Task Definition and Configuration

### Basic Task Definition

```python
from numerous.tasks import task, TaskControl

@task
def basic_task(value: int) -> int:
    """A simple task that doubles a value."""
    return value * 2

# With configuration
@task(name="configured_task", max_parallel=5, size="medium")
def configured_task(tc: TaskControl, data: list) -> dict:
    """A task with custom configuration."""
    tc.log("Processing data", "info")
    return {"processed": len(data), "items": data}
```

### Configuration Options

- **`name`**: Custom name for the task (default: function name)
- **`max_parallel`**: Maximum concurrent instances (default: 1)
- **`size`**: Resource size hint - "small", "medium", "large" (default: "small")

### Advanced Task Control

```python
@task
def advanced_task(tc: TaskControl, items: list) -> dict:
    """Task demonstrating all TaskControl features."""
    tc.log("Starting advanced task", "info")
    tc.update_status("Initializing")
    
    results = []
    total = len(items)
    
    for i, item in enumerate(items):
        # Cancellation check
        if tc.should_stop:
            tc.log("Task cancelled by user", "warning")
            break
            
        # Process item
        processed = process_item(item)
        results.append(processed)
        
        # Progress updates
        progress = (i + 1) / total * 100
        tc.update_progress(progress, f"Processed {i+1}/{total} items")
        
        # Structured logging
        if i % 10 == 0:
            tc.log(f"Checkpoint: processed {i+1} items", "info")
    
    tc.update_status("Completed")
    tc.log("Task finished successfully", "info")
    return {"processed": len(results), "results": results}
```

## Error Handling and Cancellation

### Graceful Cancellation

```python
from numerous.tasks import task, TaskControl
from numerous.tasks.exceptions import TaskCancelledError

@task
def cancellable_task(tc: TaskControl, duration: int) -> str:
    """Task that handles cancellation gracefully."""
    for i in range(duration):
        if tc.should_stop:
            tc.log("Graceful shutdown initiated", "warning")
            # Cleanup logic here
            raise TaskCancelledError("Task cancelled by user request")
        
        tc.update_progress(i / duration * 100, f"Step {i+1}/{duration}")
        time.sleep(1)
    
    return "Task completed successfully"

# Usage with cancellation
with Session() as session:
    instance = cancellable_task.instance()
    future = instance.start(30)
    
    # Cancel after 5 seconds
    import threading
    def cancel_later():
        time.sleep(5)
        instance.stop()
    
    threading.Thread(target=cancel_later).start()
    
    try:
        result = future.result()
        print(result)
    except TaskCancelledError:
        print("Task was cancelled")
```

### Error Recovery

```python
from numerous.tasks import task, TaskControl

@task
def resilient_task(tc: TaskControl, items: list, max_retries: int = 3) -> dict:
    """Task with built-in error recovery."""
    tc.log(f"Processing {len(items)} items with {max_retries} max retries", "info")
    
    results = []
    failures = []
    
    for i, item in enumerate(items):
        if tc.should_stop:
            break
            
        retries = 0
        while retries < max_retries:
            try:
                # Process item (might fail)
                result = risky_operation(item)
                results.append(result)
                break
            except Exception as e:
                retries += 1
                tc.log(f"Attempt {retries} failed for item {i}: {e}", "warning")
                
                if retries >= max_retries:
                    failures.append({"item": item, "error": str(e)})
                    tc.log(f"Max retries exceeded for item {i}", "error")
                else:
                    time.sleep(1)  # Wait before retry
        
        # Update progress
        progress = (i + 1) / len(items) * 100
        tc.update_progress(progress, f"Processed {i+1}/{len(items)} items")
    
    return {
        "successful": len(results),
        "failed": len(failures),
        "results": results,
        "failures": failures
    }
```

## Code Examples

Explore comprehensive examples in the repository:

- **[Basic Local Task Example](https://github.com/numerous-com/numerous-sdk/blob/main/python/examples/basic_local_task.py)**: Demonstrates task definition, execution, and session management
- **[Cancellation and Logging Example](https://github.com/numerous-com/numerous-sdk/blob/main/python/examples/cancellation_and_logging_task.py)**: Shows cancellation handling and structured logging
- **[Failing Task Example](https://github.com/numerous-com/numerous-sdk/blob/main/python/examples/failing_task_example.py)**: Demonstrates error handling and recovery patterns

## Common Patterns

### Data Pipeline Processing

```python
@task
def extract_data(tc: TaskControl, source: str) -> dict:
    """Extract data from source."""
    tc.log(f"Extracting data from {source}", "info")
    # Extraction logic
    return {"source": source, "data": extracted_data}

@task
def transform_data(tc: TaskControl, raw_data: dict) -> dict:
    """Transform extracted data."""
    tc.log("Transforming data", "info")
    # Transformation logic
    return {"transformed": raw_data}

@task
def load_data(tc: TaskControl, processed_data: dict) -> dict:
    """Load processed data to destination."""
    tc.log("Loading data", "info")
    # Loading logic
    return {"loaded": True, "records": len(processed_data)}

# Execute pipeline
with Session(name="etl-pipeline") as session:
    # Extract
    extract_instance = extract_data.instance()
    extract_future = extract_instance.start("database")
    raw_data = extract_future.result()
    
    # Transform
    transform_instance = transform_data.instance()
    transform_future = transform_instance.start(raw_data)
    processed_data = transform_future.result()
    
    # Load
    load_instance = load_data.instance()
    load_future = load_instance.start(processed_data)
    final_result = load_future.result()
```

### Parallel Batch Processing

```python
@task(max_parallel=4)
def process_batch(tc: TaskControl, batch_id: str, items: list) -> dict:
    """Process a batch of items in parallel."""
    tc.log(f"Processing batch {batch_id} with {len(items)} items", "info")
    
    results = []
    for i, item in enumerate(items):
        if tc.should_stop:
            break
            
        # Process individual item
        result = process_single_item(item)
        results.append(result)
        
        # Update progress
        progress = (i + 1) / len(items) * 100
        tc.update_progress(progress, f"Batch {batch_id}: {i+1}/{len(items)}")
    
    tc.log(f"Completed batch {batch_id}", "info")
    return {"batch_id": batch_id, "processed": len(results), "results": results}

# Process multiple batches concurrently
with Session(name="parallel-processing") as session:
    futures = []
    
    # Start up to 4 batches concurrently (max_parallel=4)
    for i, batch_data in enumerate(data_batches):
        instance = process_batch.instance()
        future = instance.start(f"batch_{i}", batch_data)
        futures.append(future)
    
    # Wait for all batches to complete
    results = []
    for future in futures:
        result = future.result()
        results.append(result)
    
    print(f"Processed {len(results)} batches")
```

## Best Practices

### 1. Task Design

- **Single Responsibility**: Each task should have a clear, single purpose
- **Idempotency**: Tasks should be safe to retry and produce consistent results
- **Statelessness**: Avoid shared state between task instances

### 2. Progress and Logging

- **Regular Progress Updates**: Use `tc.update_progress()` for long-running tasks
- **Structured Logging**: Use appropriate log levels and meaningful messages
- **Status Updates**: Keep users informed with `tc.update_status()`

### 3. Error Handling

- **Graceful Degradation**: Handle partial failures appropriately
- **Retry Logic**: Implement exponential backoff for transient failures
- **Cancellation Checks**: Regularly check `tc.should_stop` in loops

### 4. Resource Management

- **Appropriate Sizing**: Set correct `size` parameter for resource requirements
- **Concurrency Limits**: Configure `max_parallel` based on available resources
- **Session Management**: Use sessions to organize related tasks

### 5. Development Workflow

- **Start Simple**: Begin with direct execution for rapid development
- **Test Locally**: Use local task instances for integration testing
- **Deploy Incrementally**: Test thoroughly before switching to remote execution

## API Reference

### Core Classes

- **@task** - Decorator for defining tasks ([reference](reference/numerous/tasks/task.md))
- **Task** - Task definition class ([reference](reference/numerous/tasks/task.md))
- **TaskInstance** - Individual task execution instance ([reference](reference/numerous/tasks/task.md))
- **TaskControl** - Task control and monitoring ([reference](reference/numerous/tasks/control.md))
- **Session** - Task session management ([reference](reference/numerous/tasks/session.md))
- **Future** - Asynchronous result handling ([reference](reference/numerous/tasks/future.md))

### Task Control Methods

- **tc.log(message, level)** - Log messages with levels: "debug", "info", "warning", "error"
- **tc.update_progress(percentage, status)** - Update progress (0-100) and status message
- **tc.update_status(status)** - Update just the status message
- **tc.should_stop** - Check if task should stop (boolean property)

### Future Methods

- **future.result(timeout=None)** - Get result (blocks until complete)
- **future.status** - Current status ("pending", "running", "completed", "failed", "cancelled")
- **future.done** - Boolean indicating if task is complete
- **future.error** - Exception if task failed, None otherwise
- **future.cancel()** - Attempt to cancel the task

### Framework Integration

- **numerous.frameworks.fastapi.get_session(request)** - Get session for FastAPI applications ([reference](reference/numerous/frameworks/fastapi.md))
- **numerous.frameworks.streamlit.get_session()** - Get session for Streamlit applications ([reference](reference/numerous/frameworks/streamlit.md))

### Exceptions

- **TaskError** - Base task exception ([reference](reference/numerous/tasks/exceptions.md))
- **MaxInstancesReachedError** - Concurrency limit exceeded ([reference](reference/numerous/tasks/exceptions.md))
- **SessionNotFoundError** - No active session ([reference](reference/numerous/tasks/exceptions.md))
- **TaskCancelledError** - Task was cancelled ([reference](reference/numerous/tasks/exceptions.md))

## Advanced Topics

### Custom Task Control Handlers

For advanced use cases, customize how TaskControl operations are handled:

```python
from numerous.tasks.control import TaskControlHandler, set_task_control_handler

class CustomTaskControlHandler(TaskControlHandler):
    """Custom handler for task control operations."""
    
    def log(self, task_control, message, level, **extra_data):
        """Custom logging implementation."""
        print(f"[{level.upper()}] {task_control.task_definition_name}: {message}")
    
    def update_progress(self, task_control, progress, status):
        """Custom progress tracking."""
        print(f"Progress: {progress:.1f}% - {status}")
    
    def update_status(self, task_control, status):
        """Custom status updates."""
        print(f"Status: {status}")

# Set custom handler globally
set_task_control_handler(CustomTaskControlHandler())
```

### Backend Configuration

Control task execution environment:

```python
import os

# Local execution (default)
os.environ['NUMEROUS_TASK_BACKEND'] = 'local'

# Remote execution on Numerous platform
os.environ['NUMEROUS_TASK_BACKEND'] = 'remote'
```

### Task Metadata

Tasks can include metadata and documentation for better organization:

```python
from numerous.tasks import task

@task(name="data_processor", max_parallel=3, size="medium")
def data_processing_task(tc: TaskControl, data: dict) -> dict:
    """Process data with comprehensive logging and monitoring."""
    tc.log(f"Processing data batch with {len(data)} items", "info")
    tc.update_status("Initializing data processing")
    
    # Your processing logic here
    processed_data = {"processed": data, "timestamp": "2024-01-01"}
    
    tc.update_progress(100, "Processing complete")
    return processed_data
```

For complete API documentation, see the [API reference](reference/numerous/tasks/index.md). 