"""Core task types and data structures."""

from __future__ import annotations

import threading
from dataclasses import dataclass, field
from enum import Enum
from typing import TYPE_CHECKING, Any, Callable, Optional


if TYPE_CHECKING:
    from concurrent.futures import Future
    from datetime import datetime

    from numerous.tasks.controller import TaskController


class TaskInstanceNotFoundError(Exception):
    """Raised when a task instance cannot be found on the platform."""

    def __init__(self, instance_id: str) -> None:
        self.instance_id = instance_id
        super().__init__(f"Task instance not found: {instance_id}")


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
