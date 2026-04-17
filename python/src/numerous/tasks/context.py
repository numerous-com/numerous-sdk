"""Task execution context — provides access to controller and inputs."""

from __future__ import annotations

import contextvars
import os
from datetime import datetime
from typing import TYPE_CHECKING, Any

from numerous.tasks._get_client import get_client
from numerous.tasks.serialization import deserialize_task_inputs
from numerous.tasks.types import TaskInstanceState, TaskStatus, TaskWorkload


if TYPE_CHECKING:
    from numerous.tasks.controller import PlatformTaskController, TaskController


_current_controller: contextvars.ContextVar[TaskController | None] = (
    contextvars.ContextVar("_current_controller", default=None)
)

_current_inputs: contextvars.ContextVar[dict[str, Any] | None] = contextvars.ContextVar(
    "_current_inputs", default=None
)


def get_task_controller() -> TaskController:
    """Get the task controller for the current task execution."""
    controller = _current_controller.get()
    if controller is not None:
        return controller

    instance_id = os.getenv("NUMEROUS_TASK_INSTANCE_ID")
    if instance_id is None:
        msg = (
            "No task controller available. "
            "Must be called within a task execution context, "
            "or NUMEROUS_TASK_INSTANCE_ID must be set (platform mode)."
        )
        raise RuntimeError(msg)

    controller = _create_platform_controller(instance_id)
    _current_controller.set(controller)
    return controller


def get_task_inputs() -> dict[str, Any]:
    """Get the inputs for the current task execution."""
    inputs = _current_inputs.get()
    if inputs is not None:
        return inputs

    input_data = os.getenv("TASK_DATA_INPUT")
    inputs = deserialize_task_inputs(input_data) if input_data else {}

    _current_inputs.set(inputs)
    return inputs


def _create_platform_controller(instance_id: str) -> PlatformTaskController:
    """Create a PlatformTaskController from environment variables."""
    from numerous.tasks.controller import PlatformTaskController

    client = get_client()

    backend_instance = client.task_instance(instance_id)
    if backend_instance is None:
        msg = f"Task instance {instance_id} not found"
        raise RuntimeError(msg)

    state = TaskInstanceState(
        id=instance_id,
        task_id=backend_instance.task.id,
        status=TaskStatus.RUNNING,
        progress=0.0,
        inputs=get_task_inputs(),
        workload=TaskWorkload.REMOTE,
        created_at=datetime.now().astimezone(),
    )

    controller = PlatformTaskController(state, client)
    state.controller = controller
    return controller
