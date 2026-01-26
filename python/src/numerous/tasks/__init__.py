"""Task management API."""

__all__ = [
    # Core decorator
    "task",
    # Types
    "TaskController",
    "TaskDefinition",
    "TaskInstanceState",
    "TaskStatus",
    # Lifecycle
    "run_task",
    "wait_for_completion",
    "stop_task_instance",
    # Registry
    "register_task",
    "get_task_definition",
    "list_task_definitions",
    "get_task_instance",
    "list_task_instances",
    # Exceptions
    "TaskInstanceNotFoundError",
]

from numerous.tasks.controller import TaskController
from numerous.tasks.task import (
    get_task_definition,
    get_task_instance,
    list_task_definitions,
    list_task_instances,
    register_task,
    run_task,
    stop_task_instance,
    task,
    wait_for_completion,
)
from numerous.tasks.types import (
    TaskDefinition,
    TaskInstanceNotFoundError,
    TaskInstanceState,
    TaskStatus,
)
