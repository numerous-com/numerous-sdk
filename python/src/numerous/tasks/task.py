"""
Task management API - Object-Oriented Interface.

This module provides a task execution system for running long-running operations
asynchronously with progress tracking, cancellation, and result handling.
"""
from __future__ import annotations

from concurrent.futures import ThreadPoolExecutor, Future
from typing import Any, Callable
from datetime import datetime
from enum import Enum
import os
import uuid
import threading


# ============================================================================
# Global State
# ============================================================================

# Global registry for tasks
_task_registry: list[Task] = []


# ============================================================================
# Enumerations
# ============================================================================

class TaskStatus(Enum):
    """
    Task instance status.
    
    PENDING: Task is created but not yet running
    RUNNING: Task is currently executing
    COMPLETED: Task finished successfully
    FAILED: Task failed with an error
    CANCELLED: Task was cancelled before completion
    """
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


class TaskInstanceWorkload(Enum):
    """
    Task instance workload type.
    
    LOCAL: Task runs in local thread pool
    REMOTE: Task runs on remote platform
    """
    LOCAL = "local"
    REMOTE = "remote"


# ============================================================================
# Core Classes - Task Definition
# ============================================================================

class Task:
    """
    Task definition that wraps a Python function.
    
    A Task represents a reusable operation that can be executed multiple times
    with different inputs, creating TaskInstance objects for each execution.
    
    Attributes:
        id: Unique identifier for this task
        name: Human-readable name of the task
        command: Optional command associated with the task
        app_id: Optional application identifier
        app_version_id: Optional application version identifier
    """

    id: str
    command: list[str]
    app_id: str 
    app_version_id: str
    name: str
    _func: Callable[[dict[str, Any]], Any]
    _instances: list[TaskInstance]

    def __init__(self, func: Callable[[dict[str, Any]], Any], name: str = None):
        """
        Initialize a Task with a function.
        
        Args:
            func: The function to wrap as a task
            name: Optional name for the task (defaults to function name)
        """
        self._func = func
        self.name = name or func.__name__
        self.id = str(uuid.uuid4())
        self._instances = []
        # Register this task globally
        _task_registry.append(self)

    def create_instance(self, inputs: dict[str, Any]) -> TaskInstance:
        """
        Create a new task instance with the given inputs.

        Args:
            inputs: A dictionary of inputs to the task. Must be JSON serializable.

        Returns:
            A new task instance.
        """
        instance = TaskInstance(
            task=self,
            inputs=inputs,
            workload=TaskInstanceWorkload.LOCAL
        )
        # Track the instance
        self._instances.append(instance)
        # Execute the task instance using the global executor
        executor.execute(instance, block=False)
        return instance

    def list_instances(self) -> list[TaskInstance]:
        """
        List all task instances for the task.

        Returns:
            A list of task instances.
        """
        return self._instances

    def __call__(self, *args, **kwargs) -> TaskInstance:
        """
        Create a new task instance with the given inputs.
        
        This allows the task to be called like a function:
            task_instance = my_task(x=1, y=2)
        
        Supports three calling styles:
            1. Dictionary: my_task({'x': 1, 'y': 2})
            2. Positional: my_task(1, 2)
            3. Keyword: my_task(x=1, y=2)

        Args:
            *args: Positional arguments to pass to the task function.
            **kwargs: Keyword arguments to pass to the task function.

        Returns:
            A new task instance.
        """
        import inspect
        
        # If called with a single dict argument, use it directly
        if len(args) == 1 and isinstance(args[0], dict) and not kwargs:
            inputs = args[0]
        else:
            # Map positional and keyword arguments to parameter names
            sig = inspect.signature(self._func)
            params = list(sig.parameters.keys())
            
            # Remove task_controller from params as it's injected
            if 'task_controller' in params:
                params.remove('task_controller')
            
            # Build inputs dictionary
            inputs = {}
            
            # Add positional arguments
            for i, arg in enumerate(args):
                if i < len(params):
                    inputs[params[i]] = arg
            
            # Add keyword arguments
            inputs.update(kwargs)
        
        return self.create_instance(inputs)


# ============================================================================
# Core Classes - Task Instance Controller
# ============================================================================

class TaskInstanceController:
    """
    Controller for managing task instance execution.
    
    Provides an interface for task functions to:
    - Check if they should stop execution
    - Report progress updates
    - Set status and output
    
    This is injected into task functions that declare a 'task_controller' parameter.
    """

    def __init__(self, task_instance: TaskInstance = None):
        """
        Initialize a controller for a task instance.
        
        Args:
            task_instance: The task instance this controller manages
        """
        self._task_instance = task_instance
        self._should_stop = False

    def should_stop(self) -> bool:
        """
        Whether the task instance should be stopped or not.

        Returns:
            True if the task instance should be stopped, False otherwise.
        """
        return self._should_stop

    def request_stop(self):
        """
        Request the task instance to stop.
        """
        self._should_stop = True

    def set_progress(self, progress: float):
        """
        Set the progress of the task instance.

        Args:
            progress: The progress to set for the task instance. Must be a float between 0 and 1.

        Raises:
            ValueError: If the progress is not a float between 0 and 1.
        """
        if not 0 <= progress <= 1:
            raise ValueError("Progress must be between 0 and 1")
        if self._task_instance:
            with self._task_instance._lock:
                self._task_instance._progress = progress

    def set_status(self, status: TaskStatus | str):
        """
        Set the status of the task instance.

        Args:
            status: The status to set for the task instance.
                   Can be a TaskStatus enum or a string value.

        Raises:
            ValueError: If the status is not a valid status.
        """
        if self._task_instance:
            with self._task_instance._lock:
                if isinstance(status, str):
                    self._task_instance._status = TaskStatus(status)
                else:
                    self._task_instance._status = status

    def set_output(self, output: dict[str, Any]):
        """
        Set the output of the task instance. Must be JSON serializable.

        Args:
            output: The output to set for the task instance. Must be JSON serializable.

        Raises:
            TypeError: If the output is not JSON serializable.
        """
        if self._task_instance:
            with self._task_instance._lock:
                self._task_instance._output = output


# ============================================================================
# Core Classes - Task Instance
# ============================================================================

class TaskInstance:
    """
    A single execution of a Task with specific inputs.
    
    TaskInstance represents a running or completed task execution. It tracks:
    - Execution state (pending, running, completed, failed, cancelled)
    - Progress (0.0 to 1.0)
    - Result or error
    - Controller for task function interaction
    
    Attributes:
        id: Unique identifier for this instance
        task: The Task definition being executed
        created_at: Timestamp when instance was created
        workload: Where the task runs (LOCAL or REMOTE)
        inputs: Input parameters for this execution
    """

    id: str
    task: Task
    created_at: datetime
    workload: TaskInstanceWorkload
    inputs: dict[str, Any]
    _future: Future | None
    _progress: float
    _status: TaskStatus
    _output: dict[str, Any] | None
    _result: Any
    _controller: TaskInstanceController | None

    def __init__(self, task: Task, inputs: dict[str, Any], workload: TaskInstanceWorkload):
        """
        Initialize a task instance.
        
        Args:
            task: The Task definition to execute
            inputs: Input parameters for the task function
            workload: Where to run the task (LOCAL or REMOTE)
        """
        self.id = str(uuid.uuid4())
        self.task = task
        self.created_at = datetime.now()
        self.workload = workload
        self.inputs = inputs
        self._future = None
        self._progress = 0.0
        self._status = TaskStatus.PENDING
        self._output = None
        self._result = None
        self._controller = None
        self._lock = threading.Lock()

    def _execute(self):
        """
        Execute the task function with its inputs and controller.
        
        This method:
        1. Creates a controller for this instance
        2. Inspects the function signature
        3. Injects the controller if the function expects it
        4. Executes the function
        5. Updates status and progress
        
        Called by the executor in a separate thread.
        """
        try:
            with self._lock:
                self._status = TaskStatus.RUNNING
            # Create a controller for this instance
            self._controller = TaskInstanceController(self)
            
            # Get the function signature to check if it expects task_controller
            import inspect
            sig = inspect.signature(self.task._func)
            
            # Prepare the arguments
            kwargs = dict(self.inputs)
            if 'task_controller' in sig.parameters:
                kwargs['task_controller'] = self._controller
            
            # Call the task function with inputs
            self._result = self.task._func(**kwargs)
            with self._lock:
                self._status = TaskStatus.COMPLETED
                self._progress = 1.0
        except Exception as e:
            with self._lock:
                self._status = TaskStatus.FAILED
            raise

    def stop(self) -> None:
        """
        Stop the task instance.

        In local mode, this sets the stop flag on the controller.
        In remote mode, this will call the API to stop the task instance.
        """
        if self._controller:
            self._controller.request_stop()

    def logs(self) -> list[str]:
        """
        Get the logs for the task instance.

        In local mode, this will return an empty list.
        In remote mode, this will call the API to get the logs for the task instance.

        Returns:
            A list of logs.
        """
        ...

    @property
    def status(self) -> TaskStatus:
        """
        Get the current status of the task instance.

        Returns:
            The current TaskStatus.
        """
        return self._status

    @property
    def is_done(self) -> bool:
        """
        Whether the task instance is done or not.

        Returns:
            True if the task instance is done, False otherwise.
        """
        return self._status in (TaskStatus.COMPLETED, TaskStatus.FAILED, TaskStatus.CANCELLED)

    def get_progress(self) -> float:
        """
        Get the progress of the task instance.

        Returns:
            The progress of the task instance (0.0 to 1.0).
        """
        with self._lock:
            return self._progress

    def result(self) -> Any:
        """
        Get the result of the task instance. Blocks until the task is done.

        Returns:
            The result of the task instance.
        """
        if self._future:
            self._future.result()  # Wait for completion
        return self._result


# ============================================================================
# Task Registry - Query Functions
# ============================================================================

def list_tasks() -> list[Task]:
    """
    List all registered task definitions.

    Returns:
        A list of all task definitions in the global registry.
    """ 
    return _task_registry


def get_task(task_id: str, app_id: str) -> Task:
    """
    Get a task by its ID.

    Args:
        task_id: The ID of the task to get.
        app_id: The ID of the app to get the task for.
        
    Returns:
        The Task with the given ID, or None if not found.
    """
    ...


def get_task_instance_controller_from_env() -> TaskInstanceController:
    """
    Get a task instance controller from the environment.
    
    Used by platform executors to get a controller connected to the
    remote task instance.

    Returns:
        A task instance controller configured from environment variables.
    """
    ...


# ============================================================================
# Decorator - Task Registration
# ============================================================================

def task(func=None, *, task_name: str = None):
    """
    Decorator to convert a function into a Task.
    
    The decorated function becomes callable and returns a TaskInstance when invoked.
    The Task is automatically registered in the global registry.
    
    Usage:
        @task
        def my_task(x: int) -> int:
            return x + 1
        
        @task(task_name="custom_name")
        def another_task(x: int) -> int:
            return x * 2
        
        # Call the task to create an instance
        instance = my_task(5)

    Args:
        task_name: Optional custom name for the task. 
                  If not provided, uses the function name.

    Returns:
        A Task object that can be called to create TaskInstances.
    """
    def decorator(f: Callable[[dict[str, Any]], Any]) -> Task:
        # Create a Task instance with the function and optional name
        return Task(f, name=task_name)
    
    if func is None:
        # Called with arguments: @task(task_name="...")
        return decorator
    else:
        # Called without arguments: @task
        return decorator(func)


# ============================================================================
# Executor - Task Instance Execution
# ============================================================================

class LocalThreadTaskInstanceExecutor:
    """
    Executor that runs task instances in a local thread pool.
    
    This is the default executor for local development and testing.
    Tasks are executed asynchronously in background threads.
    
    Attributes:
        max_workers: Maximum number of concurrent task executions
        executor: The underlying ThreadPoolExecutor
    """

    def __init__(self, max_workers: int=4):
        """
        Initialize the local thread executor.
        
        Args:
            max_workers: Maximum number of concurrent tasks (default: 4)
        """
        self.max_workers = max_workers
        self.executor = ThreadPoolExecutor(max_workers=max_workers)

    def execute(self, task_instance: TaskInstance, block: bool=False) -> Future:
        """
        Execute a task instance asynchronously.

        Args:
            task_instance: The task instance to execute.
            block: If True, wait for the task to complete before returning.

        Returns:
            A Future representing the task execution.
        """
        future = self.executor.submit(task_instance._execute)
        task_instance._future = future
        if block:
            future.result()  # Wait for completion

        return future


# ============================================================================
# Executor Initialization
# ============================================================================

# Initialize the global executor based on environment
# 
# Two modes:
# 1. LOCAL (default): Uses LocalThreadTaskInstanceExecutor for local execution
# 2. PLATFORM: Would use PlatformTaskInstanceExecutor for remote execution
#
if os.getenv("NUMEROUS_EXECUTOR") == "NUMEROUS_PLATFORM_EXECUTOR":
    # Platform executor (not implemented yet)
    # 
    # When implemented, PlatformTaskInstanceExecutor will:
    # 1. Read task name and inputs from environment variables
    # 2. Execute the appropriate task function in a thread pool
    # 3. Monitor for stop signals via API calls
    # 4. Push progress and status updates to the platform API
    # 5. Upload the final result/output when complete
    #
    # For now, fall back to local executor
    print("Warning: PlatformExecutor not implemented, using LocalThreadExecutor")
    executor = LocalThreadTaskInstanceExecutor()
else:
    # Local executor for development
    executor = LocalThreadTaskInstanceExecutor()


# ============================================================================
# Example Usage
# ============================================================================

if __name__ == "__main__":
    """
    Example demonstrating the task execution API.
    
    This example shows:
    - Defining a task with @task decorator
    - Running tasks with progress tracking
    - Stopping tasks mid-execution
    - Querying tasks and instances


    Registering the task in numerous.toml:
    [[tasks]]
      name = "Task Test"
      python_task_file = "task.py" # This is the file that contains the task function

    Because we use the python task file, the platform will interpret this task as a python task,
     and execute it by importing the file and look for the task with the name "Task Test". 
    If found it will execute the task in the Python interpreter.
    """
    import time

    # Optionally add name to the task if different from the function name
    @task(task_name="Task Test")
    def test(x: int, task_controller: TaskInstanceController=None) -> int:
        """Example task that demonstrates progress tracking and cancellation."""
        NUM_STEPS = 10
        for i in range(NUM_STEPS):
            time.sleep(.1)
            task_controller.set_progress(i / NUM_STEPS)
            if task_controller.should_stop():
                print("Task instance stopped")
                return x + 1
        return x + 1

    # Create and run a task instance
    # Supports multiple calling styles:
    task_instance = test(1)       # Positional
    # task_instance = test(x=1)   # Keyword
    # task_instance = test({'x': 1})  # Dictionary

    # Monitor progress and demonstrate cancellation
    while not task_instance.is_done:
        time.sleep(.1)
        print("Progress: {:.2f}%".format(task_instance.get_progress()*100))
        
        # Stop the task when it reaches 50% progress
        if task_instance.get_progress() > 0.5:
            task_instance.stop()
            print("Stopping task instance")

    # Get the result (blocks until complete)
    print("Result:", task_instance.result())
    print("Final status:", task_instance.status.value)

    # List all registered tasks
    print("\nTasks:")
    for task in list_tasks():
        print(f"  - {task.name}")

    # List all instances of the test task
    print("\nTask instances:")
    for instance in test.list_instances():
        print(f"  - {instance.id} (status: {instance.status.value})")