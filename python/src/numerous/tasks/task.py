"""Task management API - Functional Interface."""

from __future__ import annotations

import inspect
import os
import uuid
from datetime import datetime
from functools import wraps
from typing import TYPE_CHECKING, Any, Callable, Optional, TypeVar, Union, overload

from numerous.tasks.executor import LocalThreadTaskExecutor, PlatformExecutor
from numerous.tasks.store import InMemoryTaskStore, PlatformTaskStore, TaskStore
from numerous.tasks.types import (
    TaskDefinition,
    TaskInstanceState,
    TaskStatus,
    TaskWorkload,
)


if TYPE_CHECKING:
    from concurrent.futures import Future

    from numerous.tasks._client import Client
    from numerous.tasks.executor import TaskExecutor

# Type variables for generic function types
R = TypeVar("R")


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
    """Set task progress, delegating to controller if available."""
    if state.controller:
        state.controller.set_progress(progress)
    return state


def set_status(state: TaskInstanceState, status: TaskStatus) -> TaskInstanceState:
    """Set task status, delegating to controller if available."""
    if state.controller:
        state.controller.set_status(status)
    return state


def request_stop(state: TaskInstanceState) -> TaskInstanceState:
    """Request the task to stop."""
    if state.controller:
        state.controller.request_stop()
    return state


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
        return state.future.result()
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


class _ExecutorManager:
    _instance: Optional[TaskExecutor] = None

    @classmethod
    def get_executor(cls) -> TaskExecutor:
        if cls._instance is None:
            if os.getenv("NUMEROUS_EXECUTOR") == "NUMEROUS_PLATFORM_EXECUTOR":
                cls._instance = _get_platform_executor()
            else:
                cls._instance = LocalThreadTaskExecutor(max_workers=4)
        return cls._instance


def _get_executor() -> TaskExecutor:
    return _ExecutorManager.get_executor()


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
