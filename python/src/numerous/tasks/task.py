"""Task management API - Functional Interface."""

from __future__ import annotations

import inspect
import os
import threading
import uuid
from concurrent.futures import Future, ThreadPoolExecutor
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from functools import wraps
from typing import Any, Callable, Optional, Protocol, TypeVar, Union, overload

from typing_extensions import ParamSpec


# Type variables for generic function types
P = ParamSpec("P")
R = TypeVar("R")


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

    def __init__(self, instance_state: TaskInstanceState) -> None:
        self._state = instance_state
        self._should_stop = False

    def should_stop(self) -> bool:
        """Check if the task should stop."""
        return self._should_stop

    def request_stop(self) -> None:
        """Request the task to stop."""
        self._should_stop = True

    def set_progress(self, progress: float) -> None:
        """Set the task progress."""
        set_progress(self._state, progress)

    def set_status(self, status: Union[str, TaskStatus]) -> None:
        """Set the task status."""
        if isinstance(status, str):
            with self._state.lock:
                self._state.status = TaskStatus(status)
        else:
            with self._state.lock:
                self._state.status = status

    def set_output(self, output: dict[str, Any]) -> None:
        """Set the task output."""
        with self._state.lock:
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
    future: Optional[Future[Any]] = None
    controller: Optional[TaskController] = None
    lock: threading.Lock = field(default_factory=threading.Lock, repr=False)

    def is_done(self) -> bool:
        """Check if task is complete."""
        return self.status in (
            TaskStatus.COMPLETED,
            TaskStatus.FAILED,
            TaskStatus.CANCELLED,
        )

    def stop(self) -> None:
        """Stop the task."""
        if self.controller:
            self.controller.request_stop()

    def get_progress(self) -> float:
        """Get the task progress."""
        with self.lock:
            return self.progress

    def get_status(self) -> TaskStatus:
        """Get the task status."""
        with self.lock:
            return self.status

    def get_output(self) -> Optional[dict[str, Any]]:
        """Get the task output."""
        with self.lock:
            return self.output

    def logs(self) -> list[str]:
        """Get task logs."""
        # Local mode: no logs collected
        return []


# Global state (minimized)
class InMemoryTaskStore:
    def __init__(self) -> None:
        self._task_definitions: dict[str, TaskDefinition] = {}
        self._task_instances: dict[str, TaskInstanceState] = {}
        self.lock = threading.Lock()

    def register_task(self, task_def: TaskDefinition) -> TaskDefinition:
        """Register a task definition."""
        with self.lock:
            self._task_definitions[task_def.id] = task_def
        return task_def

    def get_task_definition(self, task_id: str) -> Optional[TaskDefinition]:
        """Get a task definition by ID."""
        with self.lock:
            return self._task_definitions.get(task_id)

    def list_task_definitions(self) -> list[TaskDefinition]:
        """List all task definitions."""
        with self.lock:
            return list(self._task_definitions.values())

    def register_instance(self, state: TaskInstanceState) -> TaskInstanceState:
        """Register a task instance."""
        with self.lock:
            self._task_instances[state.id] = state
        return state

    def get_task_instance(self, instance_id: str) -> Optional[TaskInstanceState]:
        """Get a task instance by ID."""
        with self.lock:
            return self._task_instances.get(instance_id)

    def list_task_instances(
        self, task_id: Optional[str] = None
    ) -> list[TaskInstanceState]:
        """List task instances, optionally filtered by task_id."""
        with self.lock:
            if task_id:
                return [
                    inst
                    for inst in self._task_instances.values()
                    if inst.task_id == task_id
                ]
            return list(self._task_instances.values())


_store = InMemoryTaskStore()


class TaskExecutor(Protocol):
    def submit(self, task_def: TaskDefinition, state: TaskInstanceState) -> Future[Any]:
        """Submit a task for execution."""
        ...


class LocalThreadTaskExecutor:
    def __init__(self, max_workers: int = 4) -> None:
        self._executor = ThreadPoolExecutor(max_workers=max_workers)

    def submit(self, task_def: TaskDefinition, state: TaskInstanceState) -> Future[Any]:
        """Submit a task for execution."""
        return self._executor.submit(execute_task, task_def, state)


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
        created_at=datetime.now().astimezone(),
    )


def set_progress(state: TaskInstanceState, progress: float) -> TaskInstanceState:
    """Return new state with updated progress."""
    if not 0 <= progress <= 1:
        progress_error = "Progress must be between 0 and 1"
        raise ValueError(progress_error)
    with state.lock:
        state.progress = progress
    return state


def set_status(state: TaskInstanceState, status: TaskStatus) -> TaskInstanceState:
    """Return new state with updated status."""
    with state.lock:
        state.status = status
    return state


def set_result(state: TaskInstanceState, result: Any) -> TaskInstanceState:  # noqa: ANN401
    """Return new state with result."""
    with state.lock:
        state.result = result
        state.status = TaskStatus.COMPLETED
        state.progress = 1.0
    return state


def set_error(state: TaskInstanceState, error: Exception) -> TaskInstanceState:
    """Return new state with error."""
    with state.lock:
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
) -> Any:  # noqa: ANN401
    """
    Execute a task function with its inputs and controller.

    This is where side effects happen - the actual function execution.
    """
    # Update status to running
    with state.lock:
        state.status = TaskStatus.RUNNING

    try:
        # Prepare arguments
        sig = inspect.signature(task_def.func)
        kwargs = dict(state.inputs)

        # Inject controller if function expects it
        if "task_controller" in sig.parameters:
            if state.controller is None:
                state.controller = TaskController(state)
            kwargs["task_controller"] = state.controller

        # Execute the function
        result = task_def.func(**kwargs)

    except Exception as e:
        set_error(state, e)
        raise
    else:
        set_result(state, result)
        return result


def submit_task_execution(
    task_def: TaskDefinition,
    state: TaskInstanceState,
    executor: TaskExecutor,
) -> Future[Any]:
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
    block: bool = False,  # noqa: FBT001,FBT002
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


def wait_for_completion(state: TaskInstanceState) -> Any:  # noqa: ANN401
    """Wait for task completion and return result."""
    if state.future:
        state.future.result()
    return state.result


# ============================================================================
# Decorator - Functional Task Registration
# ============================================================================


@overload
def task(
    func: Callable[..., R],
    *,
    name: Optional[str] = None,
    app_id: Optional[str] = None,
) -> Callable[..., TaskInstanceState]: ...


@overload
def task(
    func: None = None,
    *,
    name: Optional[str] = None,
    app_id: Optional[str] = None,
) -> Callable[[Callable[..., R]], Callable[..., TaskInstanceState]]: ...


def task(
    func: Optional[Callable[..., R]] = None,
    *,
    name: Optional[str] = None,
    app_id: Optional[str] = None,
) -> Union[
    Callable[..., TaskInstanceState],
    Callable[[Callable[..., R]], Callable[..., TaskInstanceState]],
]:
    """
    Convert a function into a task.

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

    def decorator(f: Callable[..., R]) -> Callable[..., TaskInstanceState]:
        # Create and register task definition
        task_def = create_task_definition(f, name=name, app_id=app_id)
        register_task(task_def)

        @wraps(f)
        def wrapper(*args: Any, **kwargs: Any) -> TaskInstanceState:  # noqa: ANN401
            # Convert args/kwargs to inputs dict
            sig = inspect.signature(f)
            params = list(sig.parameters.keys())

            # Remove task_controller from params
            if "task_controller" in params:
                params.remove("task_controller")

            # Build inputs
            inputs = {}
            for i, arg in enumerate(args):
                if i < len(params):
                    inputs[params[i]] = arg
            inputs.update(kwargs)

            # Run the task
            return run_task(task_def, inputs)

        # Attach task definition and helper methods to wrapper
        wrapper.task_def = task_def  # type: ignore[attr-defined]
        wrapper.list_instances = lambda: list_task_instances(task_def.id)  # type: ignore[attr-defined]

        return wrapper

    if func is None:
        # Called with arguments: @task(name="...")
        return decorator
    # Called without arguments: @task
    return decorator(func)


# ============================================================================
# Executor Management
# ============================================================================


class ExecutorManager:
    """Manages the global task executor instance."""

    _instance: Optional[TaskExecutor] = None

    @classmethod
    def get_executor(cls) -> TaskExecutor:
        """Get or create the executor instance."""
        if cls._instance is None:
            if os.getenv("NUMEROUS_EXECUTOR") == "NUMEROUS_PLATFORM_EXECUTOR":
                # Placeholder for platform executor
                platform_executor_error = "PlatformExecutor not implemented"
                raise NotImplementedError(platform_executor_error)
            cls._instance = LocalThreadTaskExecutor(max_workers=4)
        return cls._instance

    @classmethod
    def set_executor(cls, executor: TaskExecutor) -> None:
        """Set a custom executor."""
        cls._instance = executor


def _get_executor() -> TaskExecutor:
    """Get or create the global executor."""
    return ExecutorManager.get_executor()


def set_executor(executor: TaskExecutor) -> None:
    """Set a custom executor."""
    ExecutorManager.set_executor(executor)


# ============================================================================
# Example Usage
# ============================================================================

if __name__ == "__main__":
    """
    Example usage of the functional task API.

    Registering the task in numerous.toml:
    [[tasks]]
      name = "Task Test"
       command = "numerous-executor task.py"  # File that contains the task function

    Because we use the python task file, the platform will interpret this task
    as a python task,
     and execute it by importing the file and look for the task with the name
     "Task Test".
    If found it will execute the task in the Python interpreter.
    """
    import time

    PROGRESS_THRESHOLD = 0.5

    @task
    def compute(x: int, task_controller: Optional[TaskController] = None) -> int:
        """Perform example computation."""
        num_steps = 10
        for i in range(num_steps):
            time.sleep(0.1)

            # Update progress
            if task_controller:
                task_controller.set_progress(i / num_steps)

            # Check for stop signal
            if task_controller and task_controller.should_stop():
                print("Task stopped by request")  # noqa: T201
                return x

        return x + 1

    # Run the task
    print("Starting task...")  # noqa: T201
    instance = compute(5)

    # Monitor progress
    while not instance.is_done():
        time.sleep(0.1)
        print(f"Progress: {instance.progress * 100:.1f}%")  # noqa: T201

        # Stop task if progress > 50%
        if instance.progress > PROGRESS_THRESHOLD:
            print("Requesting stop...")  # noqa: T201
            request_stop(instance)

    # Get result
    result = wait_for_completion(instance)
    print(f"Result: {result}")  # noqa: T201

    # List all tasks
    print("\nAll tasks:")  # noqa: T201
    for task_def in list_task_definitions():
        print(f"  - {task_def.name} (id: {task_def.id})")  # noqa: T201

    # List instances for this task
    print("\nTask instances:")  # noqa: T201
    for inst in list_task_instances():
        print(  # noqa: T201
            f"  - {inst.id}: {inst.status.value} (progress: {inst.progress * 100:.0f}%)"
        )
