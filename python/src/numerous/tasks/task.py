"""
Task management API - Functional Interface.
"""
from __future__ import annotations

from concurrent.futures import ThreadPoolExecutor, Future
from typing import Any, Callable, Optional, TypeVar, ParamSpec, Protocol
from datetime import datetime
from enum import Enum
from dataclasses import dataclass, field, replace
from functools import wraps
import os
import uuid
import inspect
import threading


# Type variables for generic function types
P = ParamSpec('P')
R = TypeVar('R')


class TaskStatus(Enum):
    """Task instance status."""
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


class TaskWorkload(Enum):
    """Task instance workload type."""
    LOCAL = "local"
    REMOTE = "remote"


@dataclass(frozen=True)
class TaskDefinition:
    """Immutable task definition."""
    id: str
    name: str
    func: Callable[..., Any]
    app_id: Optional[str] = None
    app_version_id: Optional[str] = None
    command: Optional[list[str]] = None


class TaskController:
    """
    Unified controller interface injected into task functions.
    
    Provides:
    - should_stop()
    - request_stop()
    - set_progress(progress)
    - set_status(status)
    - set_output(output)
    """
    def __init__(self, instance_state: "TaskInstanceState"):
        self._state = instance_state
        self._should_stop = False

    def should_stop(self) -> bool:
        return self._should_stop

    def request_stop(self) -> None:
        self._should_stop = True

    def set_progress(self, progress: float) -> None:
        set_progress(self._state, progress)

    def set_status(self, status: str | "TaskStatus") -> None:
        if isinstance(status, str):
            with self._state._lock:
                self._state.status = TaskStatus(status)
        else:
            with self._state._lock:
                self._state.status = status

    def set_output(self, output: dict[str, Any]) -> None:
        with self._state._lock:
            self._state.output = output


@dataclass
class TaskInstanceState:
    """
    Mutable task instance state.
    
    This is the only mutable structure, containing runtime state.
    All state changes go through pure functions.
    """
    id: str
    task_id: str
    status: TaskStatus
    progress: float
    inputs: dict[str, Any]
    workload: TaskWorkload
    created_at: datetime
    output: Optional[dict[str, Any]] = None
    result: Any = None
    error: Optional[Exception] = None
    future: Optional[Future] = None
    controller: Optional[TaskController] = None
    _lock: threading.Lock = field(default_factory=threading.Lock, repr=False)
    
    def is_done(self) -> bool:
        """Check if task is complete."""
        return self.status in (TaskStatus.COMPLETED, TaskStatus.FAILED, TaskStatus.CANCELLED)

    def stop(self) -> None:
        if self.controller:
            self.controller.request_stop()

    def get_progress(self) -> float:
        with self._lock:
            return self.progress

    def get_status(self) -> TaskStatus:
        with self._lock:
            return self.status

    def get_output(self) -> Optional[dict[str, Any]]:
        with self._lock:
            return self.output

    def logs(self) -> list[str]:
        # Local mode: no logs collected
        return []


# Global state (minimized)
class InMemoryTaskStore:
    def __init__(self):
        self._task_definitions: dict[str, TaskDefinition] = {}
        self._task_instances: dict[str, TaskInstanceState] = {}
        self._lock = threading.Lock()

    def register_task(self, task_def: TaskDefinition) -> TaskDefinition:
        with self._lock:
            self._task_definitions[task_def.id] = task_def
        return task_def

    def get_task_definition(self, task_id: str) -> Optional[TaskDefinition]:
        with self._lock:
            return self._task_definitions.get(task_id)

    def list_task_definitions(self) -> list[TaskDefinition]:
        with self._lock:
            return list(self._task_definitions.values())

    def register_instance(self, state: TaskInstanceState) -> TaskInstanceState:
        with self._lock:
            self._task_instances[state.id] = state
        return state

    def get_task_instance(self, instance_id: str) -> Optional[TaskInstanceState]:
        with self._lock:
            return self._task_instances.get(instance_id)

    def list_task_instances(self, task_id: Optional[str] = None) -> list[TaskInstanceState]:
        with self._lock:
            if task_id:
                return [inst for inst in self._task_instances.values() if inst.task_id == task_id]
            return list(self._task_instances.values())


_store = InMemoryTaskStore()


class TaskExecutor(Protocol):
    def submit(self, task_def: TaskDefinition, state: TaskInstanceState) -> Future: ...


class LocalThreadTaskExecutor:
    def __init__(self, max_workers: int = 4):
        self._executor = ThreadPoolExecutor(max_workers=max_workers)

    def submit(self, task_def: TaskDefinition, state: TaskInstanceState) -> Future:
        future = self._executor.submit(execute_task, task_def, state)
        return future


_executor: Optional[TaskExecutor] = None


# ============================================================================
# Pure Functions - State Transformations
# ============================================================================

def create_task_definition(
    func: Callable[..., Any],
    name: Optional[str] = None,
    app_id: Optional[str] = None,
) -> TaskDefinition:
    """Create a task definition from a function."""
    return TaskDefinition(
        id=str(uuid.uuid4()),
        name=name or func.__name__,
        func=func,
        app_id=app_id,
    )


def create_task_instance(
    task_def: TaskDefinition,
    inputs: dict[str, Any],
    workload: TaskWorkload = TaskWorkload.LOCAL,
) -> TaskInstanceState:
    """Create a new task instance state."""
    return TaskInstanceState(
        id=str(uuid.uuid4()),
        task_id=task_def.id,
        status=TaskStatus.PENDING,
        progress=0.0,
        inputs=inputs,
        workload=workload,
        created_at=datetime.now(),
    )


def set_progress(state: TaskInstanceState, progress: float) -> TaskInstanceState:
    """Return new state with updated progress."""
    if not 0 <= progress <= 1:
        raise ValueError("Progress must be between 0 and 1")
    with state._lock:
        state.progress = progress
    return state


def set_status(state: TaskInstanceState, status: TaskStatus) -> TaskInstanceState:
    """Return new state with updated status."""
    with state._lock:
        state.status = status
    return state


def set_result(state: TaskInstanceState, result: Any) -> TaskInstanceState:
    """Return new state with result."""
    with state._lock:
        state.result = result
        state.status = TaskStatus.COMPLETED
        state.progress = 1.0
    return state


def set_error(state: TaskInstanceState, error: Exception) -> TaskInstanceState:
    """Return new state with error."""
    with state._lock:
        state.error = error
        state.status = TaskStatus.FAILED
    return state


def request_stop(state: TaskInstanceState) -> TaskInstanceState:
    """Request the task to stop."""
    if state.controller:
        state.controller.request_stop()
    return state


# ============================================================================
# Task Execution - Side Effects Isolated Here
# ============================================================================

def execute_task(
    task_def: TaskDefinition,
    state: TaskInstanceState,
) -> Any:
    """
    Execute a task function with its inputs and controller.
    
    This is where side effects happen - the actual function execution.
    """
    # Update status to running
    with state._lock:
        state.status = TaskStatus.RUNNING
    
    try:
        # Prepare arguments
        sig = inspect.signature(task_def.func)
        kwargs = dict(state.inputs)
        
        # Inject controller if function expects it
        if 'task_controller' in sig.parameters:
            if state.controller is None:
                state.controller = TaskController(state)
            kwargs['task_controller'] = state.controller
        
        # Execute the function
        result = task_def.func(**kwargs)
        set_result(state, result)
        return result
        
    except Exception as e:
        set_error(state, e)
        raise


def submit_task_execution(
    task_def: TaskDefinition,
    state: TaskInstanceState,
    executor: TaskExecutor,
) -> Future:
    """Submit task for async execution and return future."""
    future = executor.submit(task_def, state)
    state.future = future
    return future


# ============================================================================
# Task Registry - Stateful Operations
# ============================================================================

def register_task(task_def: TaskDefinition) -> TaskDefinition:
    """Register a task definition globally."""
    return _store.register_task(task_def)


def get_task_definition(task_id: str) -> Optional[TaskDefinition]:
    """Get a task definition by ID."""
    return _store.get_task_definition(task_id)


def list_task_definitions() -> list[TaskDefinition]:
    """List all registered task definitions."""
    return _store.list_task_definitions()


def register_instance(state: TaskInstanceState) -> TaskInstanceState:
    """Register a task instance."""
    return _store.register_instance(state)


def get_task_instance(instance_id: str) -> Optional[TaskInstanceState]:
    """Get a task instance by ID."""
    return _store.get_task_instance(instance_id)


def list_task_instances(task_id: Optional[str] = None) -> list[TaskInstanceState]:
    """List all task instances, optionally filtered by task_id."""
    return _store.list_task_instances(task_id)


# ============================================================================
# High-Level API - Composable Functions
# ============================================================================

def run_task(
    task_def: TaskDefinition,
    inputs: dict[str, Any],
    block: bool = False,
    workload: TaskWorkload = TaskWorkload.LOCAL,
) -> TaskInstanceState:
    """
    High-level function to create and execute a task.
    
    This composes the lower-level functions into a useful operation.
    """
    # Create instance
    state = create_task_instance(task_def, inputs, workload)
    register_instance(state)
    
    # Get executor
    executor = _get_executor()
    
    # Submit for execution
    future = submit_task_execution(task_def, state, executor)
    
    # Optionally block
    if block:
        future.result()
    
    return state


def stop_task_instance(instance_id: str) -> Optional[TaskInstanceState]:
    """Stop a running task instance."""
    state = get_task_instance(instance_id)
    if state:
        return request_stop(state)
    return None


def wait_for_completion(state: TaskInstanceState) -> Any:
    """Wait for task completion and return result."""
    if state.future:
        state.future.result()
    return state.result


# ============================================================================
# Decorator - Functional Task Registration
# ============================================================================

def task(
    func: Optional[Callable[P, R]] = None,
    *,
    name: Optional[str] = None,
    app_id: Optional[str] = None,
) -> Callable[P, TaskInstanceState] | Callable[[Callable[P, R]], Callable[P, TaskInstanceState]]:
    """
    Decorator to convert a function into a task.
    
    Returns a new function that when called, creates and runs a task instance.
    
    Usage:
        @task
        def my_task(x: int) -> int:
            return x + 1
        
        @task(name="custom_name")
        def another_task(x: int) -> int:
            return x * 2
        
        # Calling the decorated function runs the task
        instance = my_task(5)
    """
    def decorator(f: Callable[P, R]) -> Callable[P, TaskInstanceState]:
        # Create and register task definition
        task_def = create_task_definition(f, name=name, app_id=app_id)
        register_task(task_def)
        
        @wraps(f)
        def wrapper(*args: P.args, **kwargs: P.kwargs) -> TaskInstanceState:
            # Convert args/kwargs to inputs dict
            sig = inspect.signature(f)
            params = list(sig.parameters.keys())
            
            # Remove task_controller from params
            if 'task_controller' in params:
                params.remove('task_controller')
            
            # Build inputs
            inputs = {}
            for i, arg in enumerate(args):
                if i < len(params):
                    inputs[params[i]] = arg
            inputs.update(kwargs)
            
            # Run the task
            return run_task(task_def, inputs)
        
        # Attach task definition and helper methods to wrapper
        wrapper.task_def = task_def  # type: ignore
        wrapper.list_instances = lambda: list_task_instances(task_def.id)  # type: ignore
        
        return wrapper
    
    if func is None:
        # Called with arguments: @task(name="...")
        return decorator
    else:
        # Called without arguments: @task
        return decorator(func)


# ============================================================================
# Executor Management
# ============================================================================

def _get_executor() -> TaskExecutor:
    """Get or create the global executor."""
    global _executor
    if _executor is None:
        if os.getenv("NUMEROUS_EXECUTOR") == "NUMEROUS_PLATFORM_EXECUTOR":
            # Placeholder for platform executor
            print("Warning: PlatformExecutor not implemented, using LocalThreadExecutor")
            _executor = LocalThreadTaskExecutor(max_workers=4)
        else:
            _executor = LocalThreadTaskExecutor(max_workers=4)
    return _executor


def set_executor(executor: TaskExecutor):
    """Set a custom executor."""
    global _executor
    _executor = executor


# ============================================================================
# Example Usage
# ============================================================================

if __name__ == "__main__":
    """
    Example usage of the functional task API.

    Registering the task in numerous.toml:
    [[tasks]]
      name = "Task Test"
      command = "numerous-executor task.py" # This is the file that contains the task function

    Because we use the python task file, the platform will interpret this task as a python task,
     and execute it by importing the file and look for the task with the name "Task Test". 
    If found it will execute the task in the Python interpreter.
    """
    import time

    @task
    def compute(x: int, task_controller=None) -> int:
        """Example task that does some computation."""
        num_steps = 10
        for i in range(num_steps):
            time.sleep(0.1)
            
            # Update progress
            if task_controller:
                task_controller.set_progress(i / num_steps)
            
            # Check for stop signal
            if task_controller and task_controller.should_stop():
                print("Task stopped by request")
                return x
        
        return x + 1

    # Run the task
    print("Starting task...")
    instance = compute(5)
    
    # Monitor progress
    while not instance.is_done():
        time.sleep(0.1)
        print(f"Progress: {instance.progress * 100:.1f}%")
        
        # Stop task if progress > 50%
        if instance.progress > 0.5:
            print("Requesting stop...")
            request_stop(instance)
    
    # Get result
    result = wait_for_completion(instance)
    print(f"Result: {result}")
    
    # List all tasks
    print("\nAll tasks:")
    for task_def in list_task_definitions():
        print(f"  - {task_def.name} (id: {task_def.id})")
    
    # List instances for this task
    print("\nTask instances:")
    for inst in compute.list_instances():
        print(f"  - {inst.id}: {inst.status.value} (progress: {inst.progress * 100:.0f}%)")
