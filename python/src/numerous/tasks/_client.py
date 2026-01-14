from __future__ import annotations

from typing import TYPE_CHECKING, Protocol


if TYPE_CHECKING:
    from numerous._client.graphql.deployment_tasks import DeploymentTasks
    from numerous._client.graphql.task_instance import TaskInstanceTaskInstance
    from numerous._client.graphql.task_instances import TaskInstances
    from numerous._client.graphql.task_start import TaskStartTaskStart
    from numerous.organization._client import OrganizationData


class Client(Protocol):
    def task_start(
        self,
        organization_slug: str,
        deploy_id: str,
        task_name: str,
        input_data: str | None = None,
    ) -> TaskStartTaskStart: ...

    def task_instance(
        self, task_instance_id: str
    ) -> TaskInstanceTaskInstance | None: ...

    def task_instance_update_progress(
        self,
        task_instance_id: str,
        value: float | None = None,
        message: str | None = None,
    ) -> None: ...

    def task_instance_set_output(
        self,
        task_instance_id: str,
        value: str,
    ) -> None: ...

    def get_organization(self, organization_id: str) -> OrganizationData | None: ...

    def deployment_tasks(
        self, organization_slug: str, deploy_id: str
    ) -> DeploymentTasks: ...

    def task_instances(
        self, organization_slug: str, deploy_id: str, task_id: str
    ) -> TaskInstances: ...

    def task_stop(self, task_instance_id: str) -> str: ...
