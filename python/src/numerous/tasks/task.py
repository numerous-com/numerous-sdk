"""Task management API - Functional Interface."""

from __future__ import annotations

import inspect
import os
import threading
import time
import uuid
from concurrent.futures import Future, ThreadPoolExecutor
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from functools import wraps
from typing import (
    TYPE_CHECKING,
    Any,
    Callable,
    Optional,
    Protocol,
    TypeVar,
    Union,
    overload,
)

from typing_extensions import ParamSpec

from numerous.tasks.serialization import (
    deserialize_task_inputs,
    deserialize_task_output,
    serialize_task_inputs,
    serialize_task_output,
)


if TYPE_CHECKING:
    from numerous._client.graphql.fragments import TaskInstanceData
    from numerous.tasks._client import Client


# Type variables for generic function types
P = ParamSpec("P")
R = TypeVar("R")


# ============================================================================
# Exceptions
# ============================================================================


class TaskInstanceNotFoundError(Exception):
    """Raised when a task instance cannot be found on the platform."""

    def __init__(self, instance_id: str) -> None:
        self.instance_id = instance_id
        super().__init__(f"Task instance not found: {instance_id}")


# ============================================================================
# Core Types
# ============================================================================


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


class TaskController(Protocol):
    """Controller interface injected into task functions."""

    def should_stop(self) -> bool:
        """Check if the task should stop."""
        ...

    def request_stop(self) -> None:
        """Request the task to stop."""
        ...

    def set_progress(self, progress: float) -> None:
        """Set the task progress (0.0 to 1.0)."""
        ...

    def set_status(self, status: Union[str, TaskStatus]) -> None:
        """Set the task status."""
        ...

    def set_output(self, output: dict[str, Any]) -> None:
        """Set the task output."""
        ...


class LocalTaskController:
    """
    Local implementation of TaskController for in-process execution.

    Updates only the local TaskInstanceState without any backend synchronization.
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
        """Set the task status (useful for local debugging)."""
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


class PlatformTaskController:
    """
    Platform implementation of TaskController.

    Syncs task state changes to the backend via GraphQL mutations.
    """

    def __init__(
        self,
        instance_state: TaskInstanceState,
        client: Client,
    ) -> None:
        self._state = instance_state
        self._client = client
        self._should_stop = False

    def should_stop(self) -> bool:
        """Check if the task should stop."""
        return self._should_stop

    def request_stop(self) -> None:
        """Request the task to stop via backend."""
        self._should_stop = True
        self._client.task_stop(task_instance_id=self._state.id)

    def set_progress(self, progress: float) -> None:
        """Report task progress to backend."""
        if not 0 <= progress <= 1:
            msg = "Progress must be between 0 and 1"
            raise ValueError(msg)

        progress_percent = progress * 100

        self._client.task_instance_update_progress(
            task_instance_id=self._state.id,
            value=progress_percent,
            message=None,
        )

    def set_status(self, _: Union[str, TaskStatus]) -> None:
        """Status is managed by the backend and cannot be set directly."""
        msg = (
            "Cannot set status in platform mode. "
            "Status is determined by the backend based on task execution state."
        )
        raise NotImplementedError(msg)

    def set_output(self, output: dict[str, Any]) -> None:
        """Set task output in backend."""
        output_json = serialize_task_output(output)
        self._client.task_instance_set_output(
            task_instance_id=self._state.id,
            value=output_json,
        )


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


# ============================================================================
# Task Store - Storage Interface
# ============================================================================


class TaskStore(Protocol):
    """Protocol for task storage operations."""

    def register_task(self, task_def: TaskDefinition) -> TaskDefinition:
        """Register a task definition."""
        ...

    def get_task_definition(self, task_id: str) -> Optional[TaskDefinition]:
        """Get a task definition by ID."""
        ...

    def list_task_definitions(self) -> list[TaskDefinition]:
        """List all task definitions."""
        ...

    def register_instance(self, state: TaskInstanceState) -> TaskInstanceState:
        """Register a task instance."""
        ...

    def get_task_instance(self, instance_id: str) -> Optional[TaskInstanceState]:
        """Get a task instance by ID."""
        ...

    def list_task_instances(
        self, task_id: Optional[str] = None
    ) -> list[TaskInstanceState]:
        """List task instances, optionally filtered by task_id."""
        ...


class InMemoryTaskStore:
    """In-memory task store for local execution."""

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


def _platform_task_placeholder() -> None:
    msg = (
        "This task function should not be called directly. "
        "Task execution happens on the platform backend."
    )
    raise RuntimeError(msg)


class PlatformTaskStore:
    """
    Platform task store that queries backend for all data.

    This store does not allow local registration of tasks or instances,
    as all data comes from the backend.
    """

    def __init__(
        self,
        client: Client,
        organization_slug: str,
        deploy_id: str,
    ) -> None:
        self._client = client
        self._org_slug = organization_slug
        self._deploy_id = deploy_id

    def register_task(self, task_def: TaskDefinition) -> TaskDefinition:
        """
        No-op in platform mode.

        Tasks are defined during app deployment and managed by the backend.
        """
        return task_def

    def get_task_definition(self, task_id: str) -> Optional[TaskDefinition]:
        """Get a task definition by ID from backend."""
        definitions = self.list_task_definitions()

        for task_def in definitions:
            if task_def.id == task_id:
                return task_def

        return None

    def list_task_definitions(self) -> list[TaskDefinition]:
        """List all task definitions from backend."""
        result = self._client.deployment_tasks(
            organization_slug=self._org_slug,
            deploy_id=self._deploy_id,
        )

        definitions = []
        for backend_task in result.tasks:
            task_def = TaskDefinition(
                id=backend_task.id,
                name=backend_task.id,
                func=_platform_task_placeholder,
                app_id=None,
                app_version_id=None,
                command=backend_task.command,
            )
            definitions.append(task_def)

        return definitions

    def register_instance(self, state: TaskInstanceState) -> TaskInstanceState:
        """
        No-op in platform mode.

        Instances are created by the platform executor via taskStart mutation.
        """
        return state

    def get_task_instance(self, instance_id: str) -> Optional[TaskInstanceState]:
        """Get a task instance by ID from backend."""
        instance = self._client.task_instance(instance_id)
        if not instance:
            return None

        return self._backend_to_state(instance)

    def list_task_instances(
        self, task_id: Optional[str] = None
    ) -> list[TaskInstanceState]:
        """List task instances from backend."""
        if task_id is None:
            msg = (
                "Platform store requires task_id to list instances. "
                "Cannot list all instances across all tasks."
            )
            raise ValueError(msg)

        result = self._client.task_instances(
            organization_slug=self._org_slug,
            deploy_id=self._deploy_id,
            task_id=task_id,
        )

        return [self._backend_to_state(inst) for inst in result.task_instances]

    def _backend_to_state(
        self, backend_instance: TaskInstanceData
    ) -> TaskInstanceState:
        inputs = {}
        if backend_instance.input:
            inputs = deserialize_task_inputs(backend_instance.input)

        output = {}
        result = None
        if backend_instance.output:
            output = deserialize_task_output(backend_instance.output)
            result = output.get("result")

        status = self._map_status(
            backend_instance.workload.status.value,
            backend_instance.workload.exit_code,
        )

        # Parse progress (backend uses 0-100, we use 0-1)
        progress = 0.0
        if backend_instance.progress.value is not None:
            progress = backend_instance.progress.value / 100

        created_at = self._parse_created_at(backend_instance.created_at)

        state = TaskInstanceState(
            id=backend_instance.id,
            task_id=backend_instance.task.id,
            status=status,
            progress=progress,
            inputs=inputs,
            output=output,
            result=result,
            workload=TaskWorkload.REMOTE,
            created_at=created_at,
        )
        state.controller = PlatformTaskController(state, self._client)
        return state

    def _parse_created_at(self, value: Any) -> datetime:  # noqa: ANN401
        if isinstance(value, datetime):
            return value

        if isinstance(value, str):
            try:
                return datetime.fromisoformat(value.replace("Z", "+00:00"))
            except (ValueError, AttributeError):
                pass

        return datetime.now().astimezone()

    def _map_status(self, backend_status: str, exit_code: Optional[int]) -> TaskStatus:
        if exit_code is not None:
            return TaskStatus.COMPLETED if exit_code == 0 else TaskStatus.FAILED

        mapping = {
            "PENDING": TaskStatus.PENDING,
            "RUNNING": TaskStatus.RUNNING,
            "STOPPED": TaskStatus.CANCELLED,
            "ERROR": TaskStatus.FAILED,
            "UNKNOWN": TaskStatus.FAILED,
        }
        return mapping.get(backend_status, TaskStatus.FAILED)


# ============================================================================
# Task Executor
# ============================================================================


class TaskExecutor(Protocol):
    """Protocol for task execution."""

    def submit(self, task_def: TaskDefinition, state: TaskInstanceState) -> Future[Any]:
        """Submit a task for execution."""
        ...


class LocalThreadTaskExecutor:
    """Executor that runs tasks in a thread pool."""

    def __init__(self, max_workers: int = 4) -> None:
        self._executor = ThreadPoolExecutor(max_workers=max_workers)

    def submit(self, task_def: TaskDefinition, state: TaskInstanceState) -> Future[Any]:
        """Submit a task for execution."""
        return self._executor.submit(execute_task, task_def, state)


class PlatformExecutor:
    """Executor that starts tasks on the Numerous platform via GraphQL."""

    def __init__(
        self,
        client: Client,
        organization_slug: str,
        deploy_id: str,
        store: PlatformTaskStore,
        poll_interval: float = 5.0,
    ) -> None:
        self._client = client
        self._org_slug = organization_slug
        self._deploy_id = deploy_id
        self._store = store
        self._poll_interval = poll_interval

    def submit(self, task_def: TaskDefinition, state: TaskInstanceState) -> Future[Any]:
        """Start a task on the Numerous platform."""
        input_json = None
        if state.inputs:
            input_json = serialize_task_inputs(state.inputs)

        result = self._client.task_start(
            organization_slug=self._org_slug,
            deploy_id=self._deploy_id,
            task_name=task_def.name,
            input_data=input_json,
        )

        state.id = result.id

        future: Future[Any] = Future()

        poll_thread = threading.Thread(
            target=self._poll_completion,
            args=(state.id, future),
            daemon=True,
            name=f"TaskPoll-{state.id[:8]}",
        )
        poll_thread.start()

        return future

    def _poll_completion(self, instance_id: str, future: Future[Any]) -> None:
        while True:
            instance = self._store.get_task_instance(instance_id)

            if instance is None:
                future.set_exception(TaskInstanceNotFoundError(instance_id))
                return

            if instance.is_done():
                if instance.error:
                    future.set_exception(instance.error)
                else:
                    future.set_result(instance.result)
                return

            time.sleep(self._poll_interval)


# ============================================================================
# Platform Context
# ============================================================================


class _PlatformContext:
    client: Client
    organization_slug: str
    deploy_id: str

    def __init__(self) -> None:
        from numerous.deployment.factory import get_deployment_id_from_env
        from numerous.tasks._get_client import get_client

        org_id = os.getenv("NUMEROUS_ORGANIZATION_ID")
        if not org_id:
            msg = "NUMEROUS_ORGANIZATION_ID environment variable is required"
            raise ValueError(msg)

        self.client = get_client()
        self.deploy_id = get_deployment_id_from_env()

        org_data = self.client.get_organization(org_id)
        if not org_data:
            msg = f"Could not retrieve organization for ID: {org_id}"
            raise ValueError(msg)

        self.organization_slug = org_data.slug


_platform_context: Optional[_PlatformContext] = None
_store: TaskStore

if os.getenv("NUMEROUS_EXECUTOR") == "NUMEROUS_PLATFORM_EXECUTOR":
    _platform_context = _PlatformContext()
    _store = PlatformTaskStore(
        _platform_context.client,
        _platform_context.organization_slug,
        _platform_context.deploy_id,
    )
else:
    _store = InMemoryTaskStore()


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
                state.controller = _create_task_controller(state)
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
                cls._instance = _get_platform_executor()
            else:
                cls._instance = LocalThreadTaskExecutor(max_workers=4)
        return cls._instance

    @classmethod
    def set_executor(cls, executor: TaskExecutor) -> None:
        """Set a custom executor."""
        cls._instance = executor


def _get_executor() -> TaskExecutor:
    """Get or create the global executor."""
    return ExecutorManager.get_executor()


def _get_platform_executor() -> TaskExecutor:
    if _platform_context is None:
        msg = "Platform context not initialized"
        raise RuntimeError(msg)

    if not isinstance(_store, PlatformTaskStore):
        msg = "Platform executor requires PlatformTaskStore"
        raise TypeError(msg)

    return PlatformExecutor(
        _platform_context.client,
        _platform_context.organization_slug,
        _platform_context.deploy_id,
        _store,
    )


def _create_task_controller(state: TaskInstanceState) -> TaskController:
    if os.getenv("NUMEROUS_EXECUTOR") == "NUMEROUS_PLATFORM_EXECUTOR":
        from numerous.tasks._get_client import get_client

        client = get_client()
        return PlatformTaskController(state, client)

    return LocalTaskController(state)


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
