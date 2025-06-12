# Tasks

With Numerous Tasks, you can define, execute, and manage long-running Python functions as distributed tasks.
Tasks provide a powerful way to handle asynchronous operations, background processing, and scalable workloads.

1. Define tasks using the simple `@task` decorator on Python functions.
2. Execute tasks locally for development or remotely on the Numerous platform.
3. Monitor task progress, status, and logs in real-time.
4. Manage task sessions and coordinate multiple related tasks.
5. Handle task cancellation, error recovery, and resource management.

!!! tip
    Remember to add `numerous` as a dependency in your project; most likely to your `requirements.txt` file.

Import the [Numerous SDK](http://www.pypi.org/project/numerous) in your Python code.

Now, you can add code to your app that is similar to the following:

```py
import time
from numerous.tasks import task, Session, TaskControl

# Define a simple task
@task
def process_data(data: list) -> dict:
    # Your task logic here
    result = {"processed": len(data), "items": data}
    return result

# Define a task with progress tracking
@task
def long_running_task(tc: TaskControl, iterations: int) -> str:
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

## Using Tasks

Numerous Tasks provides a flexible framework for defining and executing asynchronous operations with built-in progress tracking, logging, and session management.

### Defining Tasks

Use the `@task` decorator to convert any Python function into a Numerous Task:

```py
from numerous.tasks import task

@task
def simple_task(value: int) -> int:
    return value * 2

# Task with configuration
@task(max_parallel=5, size="medium")
def configured_task(data: list) -> dict:
    return {"count": len(data)}
```

#### Task Configuration Options

- `max_parallel`: Maximum number of concurrent instances (default: 1)
- `size`: Resource size hint - "small", "medium", "large" (default: "small")
- `name`: Custom name for the task (default: function name)

### Task Control and Progress Tracking

Tasks can receive a `TaskControl` object for advanced features like progress tracking, logging, and cancellation:

```py
from numerous.tasks import task, TaskControl

@task
def monitored_task(tc: TaskControl, items: list) -> dict:
    tc.log("Starting task execution", "info")
    tc.update_status("Initializing")
    
    results = []
    total = len(items)
    
    for i, item in enumerate(items):
        # Check for cancellation
        if tc.should_stop:
            tc.log("Task cancelled by user", "warning")
            break
            
        # Process item
        processed = process_item(item)  # Your processing logic
        results.append(processed)
        
        # Update progress
        progress = (i + 1) / total * 100
        tc.update_progress(progress, f"Processed {i+1}/{total} items")
        
        # Log important events
        if i % 10 == 0:
            tc.log(f"Checkpoint: processed {i+1} items", "info")
    
    tc.update_status("Completed")
    tc.log("Task finished successfully", "info")
    return {"processed": len(results), "results": results}
```

#### TaskControl Methods

- `tc.log(message, level)`: Log messages with levels: "debug", "info", "warning", "error"
- `tc.update_progress(percentage, status)`: Update progress (0-100) and status message
- `tc.update_status(status)`: Update just the status message
- `tc.should_stop`: Check if task should stop (boolean property)

### Session Management

Sessions provide a context for managing related tasks and enforcing concurrency limits:

```py
from numerous.tasks import Session

# Create a named session
with Session(name="data-processing-batch") as session:
    # All tasks created in this context belong to this session
    instance1 = task_a.instance()
    instance2 = task_b.instance()
    
    # Start tasks
    future1 = instance1.start(data_chunk_1)
    future2 = instance2.start(data_chunk_2)
    
    # Wait for results
    result1 = future1.result()
    result2 = future2.result()
```

#### Session Features

- **Task Grouping**: Organize related tasks together
- **Concurrency Control**: Enforce `max_parallel` limits per task type within the session
- **Resource Management**: Track and manage task instances
- **Context Management**: Automatic cleanup when session ends

### Task Execution Modes

Tasks support multiple execution modes for different development stages:

#### 1. Direct Execution (Development)

Execute tasks directly as function calls for quick testing:

```py
@task
def development_task(value: int) -> int:
    return value * 2

# Direct execution - no session required
result = development_task(5)  # Returns 10 immediately
```

#### 2. Local Task Instances (Testing)

Execute tasks as instances with full TaskControl features:

```py
with Session() as session:
    instance = development_task.instance()
    future = instance.start(5)
    result = future.result()  # Returns 10
```

#### 3. Remote Execution (Production)

Tasks execute on the Numerous platform with full scalability:

```py
# Same code works for remote execution
# Backend is configured via environment variables
with Session() as session:
    instance = production_task.instance()
    future = instance.start(large_dataset)
    result = future.result()  # Executes on platform
```

### Future Objects and Asynchronous Execution

Tasks return `Future` objects for asynchronous result handling:

```py
import time
from numerous.tasks import task, Session, TaskControl

@task(max_parallel=3)
def async_task(tc: TaskControl, value: int) -> int:
    tc.update_progress(50.0, "Processing")
    time.sleep(1)  # Simulate work
    tc.update_progress(100.0, "Complete")
    return value * 2

with Session() as session:
    # Start multiple tasks concurrently
    futures = []
    for i in range(5):
        instance = async_task.instance()
        future = instance.start(i)
        futures.append(future)
    
    # Collect results as they complete
    results = []
    for future in futures:
        result = future.result()  # Blocks until this task completes
        results.append(result)
    
    print(results)  # [0, 2, 4, 6, 8]
```

#### Future Methods and Properties

- `future.result(timeout=None)`: Get result (blocks until complete)
- `future.status`: Current status ("pending", "running", "completed", "failed", "cancelled")
- `future.done`: Boolean indicating if task is complete
- `future.error`: Exception if task failed, None otherwise
- `future.cancel()`: Attempt to cancel the task

### Task Cancellation

Tasks can be cancelled gracefully using TaskControl:

```py
import time
from numerous.tasks import task, Session, TaskControl

@task
def cancellable_task(tc: TaskControl, duration: int) -> str:
    for i in range(duration):
        if tc.should_stop:
            tc.log("Task cancelled gracefully", "info")
            return "Cancelled"
        
        tc.update_progress(i / duration * 100, f"Step {i+1}/{duration}")
        time.sleep(1)
    
    return "Completed"

with Session() as session:
    instance = cancellable_task.instance()
    future = instance.start(10)
    
    # Cancel after 3 seconds
    time.sleep(3)
    instance.stop()  # Requests graceful cancellation
    
    result = future.result()  # Will be "Cancelled"
```

### Error Handling

Tasks provide comprehensive error handling and recovery:

```py
from numerous.tasks import task, Session, TaskControl

@task
def error_prone_task(tc: TaskControl, should_fail: bool) -> str:
    try:
        tc.update_status("Processing")
        
        if should_fail:
            raise ValueError("Simulated error")
        
        tc.update_progress(100.0, "Success")
        return "Task completed successfully"
        
    except Exception as e:
        tc.log(f"Task failed: {str(e)}", "error")
        raise  # Re-raise to mark task as failed

with Session() as session:
    # Successful execution
    instance1 = error_prone_task.instance()
    future1 = instance1.start(False)
    result1 = future1.result()  # "Task completed successfully"
    
    # Failed execution
    instance2 = error_prone_task.instance()
    future2 = instance2.start(True)
    
    try:
        result2 = future2.result()
    except ValueError as e:
        print(f"Task failed: {e}")  # "Task failed: Simulated error"
        print(f"Future status: {future2.status}")  # "failed"
        print(f"Future error: {future2.error}")  # ValueError instance
```

### Advanced Configuration

#### Backend Configuration

Control where tasks execute using environment variables:

```bash
# Local execution (default)
export NUMEROUS_TASK_BACKEND=local

# Remote execution on Numerous platform
export NUMEROUS_TASK_BACKEND=remote
```

#### Custom Task Control Handlers

For advanced use cases, you can customize how TaskControl operations are handled:

```py
from numerous.tasks.control import TaskControlHandler, set_task_control_handler

class CustomTaskControlHandler(TaskControlHandler):
    def log(self, task_control, message, level, **extra_data):
        # Custom logging implementation
        print(f"[{level.upper()}] {task_control.task_definition_name}: {message}")
    
    def update_progress(self, task_control, progress, status):
        # Custom progress tracking
        print(f"Progress: {progress}% - {status}")
    
    def update_status(self, task_control, status):
        # Custom status updates
        print(f"Status: {status}")

# Set custom handler globally
set_task_control_handler(CustomTaskControlHandler())
```

### Best Practices

1. **Use TaskControl for Long-Running Tasks**: Always include `TaskControl` parameter for tasks that take more than a few seconds.

2. **Check for Cancellation**: Regularly check `tc.should_stop` in loops and long operations.

3. **Provide Progress Updates**: Update progress and status to help users track task execution.

4. **Handle Errors Gracefully**: Use try-catch blocks and log errors appropriately.

5. **Configure Concurrency**: Set appropriate `max_parallel` limits based on your resources.

6. **Use Sessions for Related Tasks**: Group related tasks in sessions for better organization.

7. **Test Locally First**: Use direct execution for development, then task instances for testing.

## Common Patterns

### Batch Processing

```py
@task(max_parallel=3)
def process_batch(tc: TaskControl, batch_id: str, items: list) -> dict:
    tc.log(f"Starting batch {batch_id} with {len(items)} items", "info")
    
    results = []
    for i, item in enumerate(items):
        if tc.should_stop:
            break
            
        result = process_single_item(item)
        results.append(result)
        
        progress = (i + 1) / len(items) * 100
        tc.update_progress(progress, f"Batch {batch_id}: {i+1}/{len(items)}")
    
    tc.log(f"Batch {batch_id} completed", "info")
    return {"batch_id": batch_id, "processed": len(results)}

# Process multiple batches concurrently
with Session(name="batch-processing") as session:
    futures = []
    for i, batch in enumerate(data_batches):
        instance = process_batch.instance()
        future = instance.start(f"batch_{i}", batch)
        futures.append(future)
    
    results = [f.result() for f in futures]
```

### Pipeline Processing

```py
@task
def stage_1(tc: TaskControl, data: dict) -> dict:
    tc.update_status("Stage 1: Data validation")
    # Validation logic
    return {"validated": data, "stage": 1}

@task
def stage_2(tc: TaskControl, data: dict) -> dict:
    tc.update_status("Stage 2: Data transformation")
    # Transformation logic
    return {"transformed": data, "stage": 2}

@task
def stage_3(tc: TaskControl, data: dict) -> dict:
    tc.update_status("Stage 3: Data storage")
    # Storage logic
    return {"stored": data, "stage": 3}

# Execute pipeline
with Session(name="data-pipeline") as session:
    # Stage 1
    instance1 = stage_1.instance()
    future1 = instance1.start(raw_data)
    result1 = future1.result()
    
    # Stage 2
    instance2 = stage_2.instance()
    future2 = instance2.start(result1)
    result2 = future2.result()
    
    # Stage 3
    instance3 = stage_3.instance()
    future3 = instance3.start(result2)
    final_result = future3.result()
```

## API Reference

See the [API reference](reference/numerous/tasks/index.md) for complete details on all classes and methods.

### Core Classes

- [`@task`](reference/numerous/tasks/task.md#numerous.tasks.task.task) - Decorator for defining tasks
- [`Task`](reference/numerous/tasks/task.md#numerous.tasks.task.Task) - Task definition class
- [`TaskInstance`](reference/numerous/tasks/task.md#numerous.tasks.task.TaskInstance) - Individual task execution instance
- [`TaskControl`](reference/numerous/tasks/control.md#numerous.tasks.control.TaskControl) - Task control and monitoring
- [`Session`](reference/numerous/tasks/session.md#numerous.tasks.session.Session) - Task session management
- [`Future`](reference/numerous/tasks/future.md#numerous.tasks.future.Future) - Asynchronous result handling

### Exceptions

- [`TaskError`](reference/numerous/tasks/exceptions.md#numerous.tasks.exceptions.TaskError) - Base task exception
- [`MaxInstancesReachedError`](reference/numerous/tasks/exceptions.md#numerous.tasks.exceptions.MaxInstancesReachedError) - Concurrency limit exceeded
- [`SessionNotFoundError`](reference/numerous/tasks/exceptions.md#numerous.tasks.exceptions.SessionNotFoundError) - No active session
- [`TaskCancelledError`](reference/numerous/tasks/exceptions.md#numerous.tasks.exceptions.TaskCancelledError) - Task was cancelled 