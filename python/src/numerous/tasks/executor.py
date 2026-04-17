"""Task executor implementations."""

from __future__ import annotations

import inspect
import threading
import time
from concurrent.futures import Future, ThreadPoolExecutor
from typing import TYPE_CHECKING, Any, Protocol

from numerous.tasks.context import _current_controller, _current_inputs
from numerous.tasks.controller import LocalTaskController
from numerous.tasks.serialization import serialize_task_inputs
from numerous.tasks.types import (
    TaskDefinition,
    TaskInstanceNotFoundError,
    TaskInstanceState,
    TaskStatus,
)


if TYPE_CHECKING:
    from numerous.tasks._client import Client
    from numerous.tasks.store import PlatformTaskStore


class TaskExecutor(Protocol):
    """Protocol for task execution."""

    def submit(self, task_def: TaskDefinition, state: TaskInstanceState) -> Future[Any]:
        """Submit a task for execution."""
        ...


def _set_finished(state: TaskInstanceState) -> TaskInstanceState:
    with state.lock:
        if state.controller is not None:
            if state.controller.should_stop():
                state.status = TaskStatus.CANCELLED
            else:
                state.status = TaskStatus.COMPLETED
    return state


def _set_error(state: TaskInstanceState, error: Exception) -> TaskInstanceState:
    with state.lock:
        state.error = error
        state.status = TaskStatus.FAILED
    return state


def _execute_local_task(
    task_def: TaskDefinition,
    state: TaskInstanceState,
) -> Any:  # noqa: ANN401
    with state.lock:
        state.status = TaskStatus.RUNNING

    if state.controller is None:
        state.controller = LocalTaskController(state)

    controller_token = _current_controller.set(state.controller)
    inputs_token = _current_inputs.set(state.inputs)

    try:
        sig = inspect.signature(task_def.func)
        kwargs = dict(state.inputs)

        if "task_controller" in sig.parameters:
            kwargs["task_controller"] = state.controller

        task_def.func(**kwargs)

    except Exception as e:
        _set_error(state, e)
        raise
    else:
        _set_finished(state)
        return state.result
    finally:
        _current_controller.reset(controller_token)
        _current_inputs.reset(inputs_token)


class LocalThreadTaskExecutor:
    """Executor that runs tasks in a thread pool."""

    def __init__(self, max_workers: int = 4) -> None:
        self._executor = ThreadPoolExecutor(max_workers=max_workers)

    def submit(self, task_def: TaskDefinition, state: TaskInstanceState) -> Future[Any]:
        """Submit a task for execution."""
        return self._executor.submit(_execute_local_task, task_def, state)


# ============================================================================
# Platform Implementation
# ============================================================================


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
            args=(state.id, future, state),
            daemon=True,
            name=f"TaskPoll-{state.id[:8]}",
        )
        poll_thread.start()

        return future

    def _poll_completion(
        self, instance_id: str, future: Future[Any], state: TaskInstanceState
    ) -> None:
        while True:
            instance = self._store.get_task_instance(instance_id)

            if instance is None:
                future.set_exception(TaskInstanceNotFoundError(instance_id))
                return

            with state.lock:
                state.status = instance.status
                state.progress = instance.progress
                state.result = instance.result
                state.error = instance.error
                state.output = instance.output

            if instance.is_done():
                if instance.error:
                    future.set_exception(instance.error)
                else:
                    future.set_result(instance.result)
                return

            time.sleep(self._poll_interval)
