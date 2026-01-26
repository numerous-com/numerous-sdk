from __future__ import annotations

import base64
import json
from dataclasses import dataclass
from datetime import datetime
from typing import TYPE_CHECKING, Any
from unittest.mock import Mock

import pytest

from numerous._client.graphql.client import Client as GQLClient
from numerous._client.graphql.deployment_tasks import (
    DeploymentTasks,
    DeploymentTasksTasks,
)
from numerous._client.graphql.enums import TaskWorkloadStatus
from numerous._client.graphql.task_instance import (
    TaskInstance,
    TaskInstanceTaskInstance,
)
from numerous._client.graphql.task_instances import (
    TaskInstances,
    TaskInstancesTaskInstances,
)
from numerous._client.graphql.task_start import TaskStart, TaskStartTaskStart
from numerous.tasks.controller import PlatformTaskController
from numerous.tasks.executor import PlatformExecutor
from numerous.tasks.store import PlatformTaskStore
from numerous.tasks.types import (
    TaskDefinition,
    TaskInstanceState,
    TaskStatus,
    TaskWorkload,
)


if TYPE_CHECKING:
    from numerous.tasks._client import Client


ORGANIZATION_SLUG = "test-org"
ORGANIZATION_ID = "test-org-id"
DEPLOY_ID = "test-deploy-123"
FIRST_TASK_ID = "task-1"
SECOND_TASK_ID = "task-2"
TASK_RESULT = 123
FIRST_INSTANCE_ID = "instance-1"
SECOND_INSTANCE_ID = "instance-2"
THIRD_INSTANCE_ID = "instance-3"
HEADERS_WITH_AUTHORIZATION = {"headers": {"Authorization": "Bearer token"}}

EXPECTED_PROGRESS_FRACTION = 0.5
EXPECTED_PROGRESS_PERCENT = 50.0


def _encode_base64(data: dict[str, Any]) -> str:
    json_str = json.dumps(data)
    return base64.b64encode(json_str.encode("utf-8")).decode("utf-8")


@dataclass
class TaskInstanceResponseData:
    instance_id: str
    task_id: str
    status: TaskWorkloadStatus = TaskWorkloadStatus.PENDING
    exit_code: int | None = None
    input_data: str | None = None
    output_data: str | None = None
    progress_value: float | None = None
    created_at: str | None = None

    def to_dict(self) -> dict[str, Any]:  # noqa: D102
        return {
            "id": self.instance_id,
            "task": {"id": self.task_id},
            "workload": {
                "status": self.status,
                "exitCode": self.exit_code,
            },
            "input": self.input_data,
            "output": self.output_data,
            "progress": {"value": self.progress_value},
            "createdAt": self.created_at or "2024-01-01T00:00:00Z",
        }


def _deployment_tasks_response(tasks: list[tuple[str, list[str]]]) -> DeploymentTasks:
    task_list = [
        DeploymentTasksTasks.model_validate({"id": task_id, "command": command})
        for task_id, command in tasks
    ]
    return DeploymentTasks.model_validate({"tasks": task_list})


def _task_instances_response(
    instances: list[tuple[str, str, TaskWorkloadStatus, int | None]],
) -> TaskInstances:
    instance_list = [
        TaskInstancesTaskInstances.model_validate(
            TaskInstanceResponseData(
                instance_id=inst_id,
                task_id=task_id,
                status=status,
                exit_code=exit_code,
            ).to_dict()
        )
        for inst_id, task_id, status, exit_code in instances
    ]
    return TaskInstances.model_validate({"taskInstances": instance_list})


def _task_instance_response(
    data: TaskInstanceResponseData,
) -> TaskInstance:
    instance = TaskInstanceTaskInstance.model_validate(data.to_dict())
    return TaskInstance.model_validate({"taskInstance": instance})


def _task_start_response(instance_id: str) -> TaskStart:
    task_start = TaskStartTaskStart.model_validate({"id": instance_id})
    return TaskStart.model_validate({"taskStart": task_start})


@pytest.fixture
def gql() -> Mock:
    return Mock(GQLClient)


@pytest.fixture
def client(gql: Mock) -> Client:
    from numerous._client.graphql_client import GraphQLClient

    return GraphQLClient(gql, ORGANIZATION_ID, "token")


@pytest.fixture
def store(client: Client) -> PlatformTaskStore:
    return PlatformTaskStore(client, ORGANIZATION_SLUG, DEPLOY_ID)


@pytest.fixture
def executor(client: Client, store: PlatformTaskStore) -> PlatformExecutor:
    return PlatformExecutor(
        client,
        ORGANIZATION_SLUG,
        DEPLOY_ID,
        store,
        poll_interval=0.1,  # Short interval for tests
    )


def test_platform_store_list_task_definitions(
    gql: Mock,
    store: PlatformTaskStore,
) -> None:
    expected_tasks = [
        (FIRST_TASK_ID, ["python", "task1.py"]),
        (SECOND_TASK_ID, ["python", "task-2.py"]),
    ]
    gql.deployment_tasks.return_value = _deployment_tasks_response(expected_tasks)

    definitions = store.list_task_definitions()

    gql.deployment_tasks.assert_called_once_with(
        organization_slug=ORGANIZATION_SLUG,
        deploy_id=DEPLOY_ID,
        **HEADERS_WITH_AUTHORIZATION,
    )
    expected_task_count = 2
    assert len(definitions) == expected_task_count
    assert definitions[0].id == FIRST_TASK_ID
    assert definitions[0].name == FIRST_TASK_ID
    assert definitions[0].command == ["python", "task1.py"]
    assert definitions[1].id == SECOND_TASK_ID
    assert definitions[1].command == ["python", "task-2.py"]


def test_platform_store_get_task_definition(
    gql: Mock,
    store: PlatformTaskStore,
) -> None:
    expected_tasks = [
        (FIRST_TASK_ID, ["python", "task1.py"]),
        (SECOND_TASK_ID, ["python", "task-2.py"]),
    ]
    gql.deployment_tasks.return_value = _deployment_tasks_response(expected_tasks)

    definition = store.get_task_definition(FIRST_TASK_ID)

    assert definition is not None
    assert definition.id == FIRST_TASK_ID
    assert definition.command == ["python", "task1.py"]


def test_platform_store_get_task_definition_not_found(
    gql: Mock, store: PlatformTaskStore
) -> None:
    gql.deployment_tasks.return_value = _deployment_tasks_response([])

    definition = store.get_task_definition("non-existent-task")

    assert definition is None


def test_platform_store_get_task_instance(gql: Mock, store: PlatformTaskStore) -> None:
    # Backend returns progress as percentage (0-100), SDK converts to fraction (0-1)
    gql.task_instance.return_value = _task_instance_response(
        TaskInstanceResponseData(
            instance_id=FIRST_INSTANCE_ID,
            task_id=FIRST_TASK_ID,
            status=TaskWorkloadStatus.RUNNING,
            progress_value=EXPECTED_PROGRESS_PERCENT,
        )
    )

    instance = store.get_task_instance(FIRST_INSTANCE_ID)

    gql.task_instance.assert_called_once_with(
        task_instance_id=FIRST_INSTANCE_ID, **HEADERS_WITH_AUTHORIZATION
    )
    assert instance is not None
    assert instance.id == FIRST_INSTANCE_ID
    assert instance.task_id == FIRST_TASK_ID
    assert instance.status == TaskStatus.RUNNING
    assert instance.progress == EXPECTED_PROGRESS_FRACTION
    assert instance.workload == TaskWorkload.REMOTE


def test_platform_store_get_task_instance_not_found(
    gql: Mock, store: PlatformTaskStore
) -> None:
    # GraphQLClient returns None when the wrapper returns taskInstance=null
    gql.task_instance.return_value = TaskInstance.model_validate({"taskInstance": None})

    instance = store.get_task_instance("non-existent-instance")

    assert instance is None


def test_platform_store_get_task_instance_completed(
    gql: Mock, store: PlatformTaskStore
) -> None:
    output_data = {"result": TASK_RESULT}
    output_json = _encode_base64(output_data)
    gql.task_instance.return_value = _task_instance_response(
        TaskInstanceResponseData(
            instance_id=FIRST_INSTANCE_ID,
            task_id=FIRST_TASK_ID,
            status=TaskWorkloadStatus.RUNNING,
            exit_code=0,
            output_data=output_json,
        )
    )

    instance = store.get_task_instance(FIRST_INSTANCE_ID)

    assert instance is not None
    assert instance.status == TaskStatus.COMPLETED
    assert instance.output == output_data
    assert instance.result == TASK_RESULT


def test_platform_store_get_task_instance_failed(
    gql: Mock, store: PlatformTaskStore
) -> None:
    gql.task_instance.return_value = _task_instance_response(
        TaskInstanceResponseData(
            instance_id=FIRST_INSTANCE_ID,
            task_id=FIRST_TASK_ID,
            status=TaskWorkloadStatus.RUNNING,
            exit_code=1,
        )
    )

    instance = store.get_task_instance(FIRST_INSTANCE_ID)

    assert instance is not None
    assert instance.status == TaskStatus.FAILED


def test_platform_store_list_task_instances(
    gql: Mock,
    store: PlatformTaskStore,
) -> None:
    expected_instances = [
        (FIRST_INSTANCE_ID, FIRST_TASK_ID, TaskWorkloadStatus.RUNNING, None),
        (SECOND_INSTANCE_ID, FIRST_TASK_ID, TaskWorkloadStatus.RUNNING, 0),
        (THIRD_INSTANCE_ID, FIRST_TASK_ID, TaskWorkloadStatus.RUNNING, 1),
    ]
    gql.task_instances.return_value = _task_instances_response(expected_instances)

    instances = store.list_task_instances(task_id=FIRST_TASK_ID)

    gql.task_instances.assert_called_once_with(
        organization_slug=ORGANIZATION_SLUG,
        deploy_id=DEPLOY_ID,
        task_id=FIRST_TASK_ID,
        **HEADERS_WITH_AUTHORIZATION,
    )
    expected_instance_count = 3
    assert len(instances) == expected_instance_count
    assert instances[0].id == FIRST_INSTANCE_ID
    assert instances[0].task_id == FIRST_TASK_ID
    assert instances[0].status == TaskStatus.RUNNING
    assert instances[1].id == SECOND_INSTANCE_ID
    assert instances[1].task_id == FIRST_TASK_ID
    assert instances[1].status == TaskStatus.COMPLETED
    assert instances[2].id == THIRD_INSTANCE_ID
    assert instances[2].task_id == FIRST_TASK_ID
    assert instances[2].status == TaskStatus.FAILED


def test_platform_store_list_task_instances_requires_task_id(
    store: PlatformTaskStore,
) -> None:
    with pytest.raises(ValueError, match="Platform store requires task_id"):
        store.list_task_instances(task_id=None)


def test_platform_store_register_task_is_noop(store: PlatformTaskStore) -> None:
    task_def = TaskDefinition(
        id=FIRST_TASK_ID,
        name=FIRST_TASK_ID,
        func=lambda: None,
    )

    result = store.register_task(task_def)

    assert result == task_def


def test_platform_store_register_instance_is_noop(store: PlatformTaskStore) -> None:
    state = TaskInstanceState(
        id=FIRST_INSTANCE_ID,
        task_id=FIRST_TASK_ID,
        status=TaskStatus.PENDING,
        progress=0.0,
        inputs={},
        workload=TaskWorkload.REMOTE,
        created_at=datetime.now().astimezone(),
    )

    result = store.register_instance(state)

    assert result == state


def test_platform_store_maps_backend_statuses(
    gql: Mock,
    store: PlatformTaskStore,
) -> None:
    test_cases = [
        (TaskWorkloadStatus.PENDING, None, TaskStatus.PENDING),
        (TaskWorkloadStatus.RUNNING, None, TaskStatus.RUNNING),
        (TaskWorkloadStatus.STOPPED, None, TaskStatus.CANCELLED),
        (TaskWorkloadStatus.ERROR, None, TaskStatus.FAILED),
        (TaskWorkloadStatus.UNKNOWN, None, TaskStatus.FAILED),
    ]

    for backend_status, exit_code, expected_status in test_cases:
        gql.task_instance.return_value = _task_instance_response(
            TaskInstanceResponseData(
                instance_id=FIRST_INSTANCE_ID,
                task_id=FIRST_TASK_ID,
                status=backend_status,
                exit_code=exit_code,
            )
        )

        instance = store.get_task_instance(FIRST_INSTANCE_ID)

        assert instance is not None
        assert instance.status == expected_status, f"Failed for {backend_status}"


def test_platform_store_parses_task_input(
    gql: Mock,
    store: PlatformTaskStore,
) -> None:
    input_data = {"x": 10, "y": 20}
    input_json = _encode_base64(input_data)
    gql.task_instance.return_value = _task_instance_response(
        TaskInstanceResponseData(
            instance_id=FIRST_INSTANCE_ID,
            task_id=FIRST_TASK_ID,
            input_data=input_json,
        )
    )

    instance = store.get_task_instance(FIRST_INSTANCE_ID)

    assert instance is not None
    assert instance.inputs == input_data


def test_platform_executor_submit_task(
    gql: Mock,
    executor: PlatformExecutor,
) -> None:
    task_def = TaskDefinition(
        id=FIRST_TASK_ID,
        name=FIRST_TASK_ID,
        func=lambda: None,
    )
    inputs = {"x": 5}
    state = TaskInstanceState(
        id="temp-id",
        task_id=FIRST_TASK_ID,
        status=TaskStatus.PENDING,
        progress=0.0,
        inputs=inputs,
        workload=TaskWorkload.REMOTE,
        created_at=datetime.now().astimezone(),
    )

    gql.task_start.return_value = _task_start_response(FIRST_INSTANCE_ID)
    output_data = {"result": TASK_RESULT}
    gql.task_instance.return_value = _task_instance_response(
        TaskInstanceResponseData(
            instance_id=FIRST_INSTANCE_ID,
            task_id=FIRST_TASK_ID,
            status=TaskWorkloadStatus.RUNNING,
            exit_code=0,
            output_data=_encode_base64(output_data),
        )
    )

    future = executor.submit(task_def, state)

    gql.task_start.assert_called_once_with(
        organization_slug=ORGANIZATION_SLUG,
        deploy_id=DEPLOY_ID,
        task_name=FIRST_TASK_ID,
        input=_encode_base64(inputs),
        **HEADERS_WITH_AUTHORIZATION,
    )
    assert state.id == FIRST_INSTANCE_ID
    assert future is not None

    # Wait for completion (the mock returns completed immediately)
    result = future.result(timeout=1.0)
    assert result == TASK_RESULT


def test_platform_executor_submit_task_without_input(
    gql: Mock,
    executor: PlatformExecutor,
) -> None:
    task_def = TaskDefinition(
        id=FIRST_TASK_ID,
        name=FIRST_TASK_ID,
        func=lambda: None,
    )
    state = TaskInstanceState(
        id="temp-id",
        task_id=FIRST_TASK_ID,
        status=TaskStatus.PENDING,
        progress=0.0,
        inputs={},
        workload=TaskWorkload.REMOTE,
        created_at=datetime.now().astimezone(),
    )

    gql.task_start.return_value = _task_start_response(FIRST_INSTANCE_ID)
    gql.task_instance.return_value = _task_instance_response(
        TaskInstanceResponseData(
            instance_id=FIRST_INSTANCE_ID,
            task_id=FIRST_TASK_ID,
            status=TaskWorkloadStatus.RUNNING,
            exit_code=0,
        )
    )

    future = executor.submit(task_def, state)

    gql.task_start.assert_called_once_with(
        organization_slug=ORGANIZATION_SLUG,
        deploy_id=DEPLOY_ID,
        task_name=FIRST_TASK_ID,
        input=None,
        **HEADERS_WITH_AUTHORIZATION,
    )
    assert future is not None


def test_platform_controller_set_progress(gql: Mock, client: Client) -> None:
    state = TaskInstanceState(
        id=FIRST_INSTANCE_ID,
        task_id=FIRST_TASK_ID,
        status=TaskStatus.RUNNING,
        progress=0.0,
        inputs={},
        workload=TaskWorkload.REMOTE,
        created_at=datetime.now().astimezone(),
    )
    controller = PlatformTaskController(state, client)

    controller.set_progress(0.5)

    # SDK uses fraction (0-1), backend expects percentage (0-100)
    gql.task_instance_update_progress.assert_called_once_with(
        task_instance_id=FIRST_INSTANCE_ID,
        value=50.0,
        message=None,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_platform_controller_set_progress_validates_range(client: Client) -> None:
    state = TaskInstanceState(
        id=FIRST_INSTANCE_ID,
        task_id=FIRST_TASK_ID,
        status=TaskStatus.RUNNING,
        progress=0.0,
        inputs={},
        workload=TaskWorkload.REMOTE,
        created_at=datetime.now().astimezone(),
    )
    controller = PlatformTaskController(state, client)

    with pytest.raises(ValueError, match="Progress must be between 0 and 1"):
        controller.set_progress(1.5)

    with pytest.raises(ValueError, match="Progress must be between 0 and 1"):
        controller.set_progress(-0.1)


def test_platform_controller_set_output(gql: Mock, client: Client) -> None:
    state = TaskInstanceState(
        id=FIRST_INSTANCE_ID,
        task_id=FIRST_TASK_ID,
        status=TaskStatus.RUNNING,
        progress=0.0,
        inputs={},
        workload=TaskWorkload.REMOTE,
        created_at=datetime.now().astimezone(),
    )
    controller = PlatformTaskController(state, client)

    output = {"some_field": "some_value"}
    controller.set_output(output)

    gql.task_instance_set_output.assert_called_once_with(
        task_instance_id=FIRST_INSTANCE_ID,
        value=_encode_base64(output),
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_platform_controller_request_stop(gql: Mock, client: Client) -> None:
    state = TaskInstanceState(
        id=FIRST_INSTANCE_ID,
        task_id=FIRST_TASK_ID,
        status=TaskStatus.RUNNING,
        progress=0.0,
        inputs={},
        workload=TaskWorkload.REMOTE,
        created_at=datetime.now().astimezone(),
    )
    controller = PlatformTaskController(state, client)

    controller.request_stop()

    assert controller.should_stop() is True
    gql.task_stop.assert_called_once_with(
        task_instance_id=FIRST_INSTANCE_ID, **HEADERS_WITH_AUTHORIZATION
    )


def test_platform_controller_should_stop_initial_state(client: Client) -> None:
    state = TaskInstanceState(
        id=FIRST_INSTANCE_ID,
        task_id=FIRST_TASK_ID,
        status=TaskStatus.RUNNING,
        progress=0.0,
        inputs={},
        workload=TaskWorkload.REMOTE,
        created_at=datetime.now().astimezone(),
    )
    controller = PlatformTaskController(state, client)

    assert controller.should_stop() is False


def test_platform_controller_set_status_raises_error(client: Client) -> None:
    state = TaskInstanceState(
        id=FIRST_INSTANCE_ID,
        task_id=FIRST_TASK_ID,
        status=TaskStatus.RUNNING,
        progress=0.0,
        inputs={},
        workload=TaskWorkload.REMOTE,
        created_at=datetime.now().astimezone(),
    )
    controller = PlatformTaskController(state, client)

    with pytest.raises(NotImplementedError, match="Cannot set status in platform mode"):
        controller.set_status(TaskStatus.COMPLETED)
