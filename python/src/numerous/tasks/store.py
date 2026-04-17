"""Task store implementations for task and instance storage."""

from __future__ import annotations

import threading
from datetime import datetime
from typing import TYPE_CHECKING, Any, Optional, Protocol

from numerous.tasks.controller import PlatformTaskController
from numerous.tasks.serialization import (
    deserialize_task_inputs,
    deserialize_task_output,
)
from numerous.tasks.types import (
    TaskDefinition,
    TaskInstanceState,
    TaskStatus,
    TaskWorkload,
)


if TYPE_CHECKING:
    from numerous._client.graphql.fragments import TaskInstanceData
    from numerous.tasks._client import Client


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
