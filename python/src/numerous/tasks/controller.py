"""Task controller implementations."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any, Protocol, Union

from numerous.tasks.serialization import serialize_task_output
from numerous.tasks.types import TaskInstanceState, TaskStatus


if TYPE_CHECKING:
    from numerous.tasks._client import Client


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
        if not 0 <= progress <= 1:
            msg = "Progress must be between 0 and 1"
            raise ValueError(msg)
        with self._state.lock:
            self._state.progress = progress

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
