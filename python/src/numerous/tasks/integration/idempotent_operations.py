"""
Idempotent API operations for task and instance registration.

Provides upsert operations that safely handle task definition registration
and instance creation with proper versioning and session scoping.
"""

import logging
import time
from typing import Dict, Any, Optional
from dataclasses import dataclass

logger = logging.getLogger(__name__)


@dataclass
class TaskDefinition:
    """Task definition data structure for API registration."""
    name: str
    version: str
    function_name: str
    module: str
    parameters: Dict[str, Any]
    metadata: Dict[str, Any]


@dataclass
class TaskInstance:
    """Task instance data structure for API registration."""
    id: str
    session_id: str
    task_name: str
    task_version: str
    status: str
    created_at: float


class IdempotentOperations:
    """
    Handles idempotent task and instance operations with the API.
    
    Provides upsert operations that:
    - Register task definitions with versioning
    - Create session-scoped instances
    - Handle conflicts gracefully
    """
    
    def __init__(self, graphql_client):
        """
        Initialize with GraphQL client.
        
        Args:
            graphql_client: GraphQL client for API communication
        """
        self.client = graphql_client
    
    async def upsert_task_definition(
        self, 
        task_name: str, 
        task_definition: TaskDefinition
    ) -> Dict[str, Any]:
        """
        Upsert a task definition with the API.
        
        Args:
            task_name: Name of the task
            task_definition: Task definition data
            
        Returns:
            Dictionary containing task registration result
            
        Raises:
            RuntimeError: If API communication fails
        """
        try:
            # TODO: Implement GraphQL mutation for task definition upsert
            # For now, simulate the operation
            logger.info(f"[IdempotentOps] Upserting task definition: {task_name} v{task_definition.version}")
            
            # Simulate API call
            result = {
                "task_name": task_name,
                "version": task_definition.version,
                "status": "registered",
                "created": True,  # Would be False if task already existed
                "timestamp": time.time()
            }
            
            logger.debug(f"[IdempotentOps] Task definition upsert result: {result}")
            return result
            
        except Exception as e:
            logger.error(f"Failed to upsert task definition {task_name}: {e}")
            raise RuntimeError(f"Failed to upsert task definition: {e}") from e
    
    async def upsert_task_instance(
        self,
        instance_id: str,
        session_id: str,
        task_name: str,
        task_version: str,
        inputs: Optional[Dict[str, Any]] = None
    ) -> TaskInstance:
        """
        Upsert a task instance with the API.
        
        Args:
            instance_id: Unique instance identifier
            session_id: Session ID for scoping
            task_name: Name of the task
            task_version: Version of the task
            inputs: Optional task inputs
            
        Returns:
            TaskInstance object
            
        Raises:
            RuntimeError: If API communication fails
        """
        try:
            # TODO: Implement GraphQL mutation for instance upsert
            # For now, simulate the operation
            logger.info(f"[IdempotentOps] Upserting task instance: {instance_id} (session: {session_id})")
            
            # Simulate checking if instance exists
            existing_instance = await self._get_existing_instance(instance_id)
            
            if existing_instance:
                logger.debug(f"[IdempotentOps] Instance {instance_id} already exists, returning existing")
                return existing_instance
            
            # Create new instance
            instance = TaskInstance(
                id=instance_id,
                session_id=session_id,
                task_name=task_name,
                task_version=task_version,
                status="PENDING",
                created_at=time.time()
            )
            
            logger.debug(f"[IdempotentOps] Created new task instance: {instance}")
            return instance
            
        except Exception as e:
            logger.error(f"Failed to upsert task instance {instance_id}: {e}")
            raise RuntimeError(f"Failed to upsert task instance: {e}") from e
    
    async def _get_existing_instance(self, instance_id: str) -> Optional[TaskInstance]:
        """
        Check if a task instance already exists.
        
        Args:
            instance_id: Instance ID to check
            
        Returns:
            TaskInstance if exists, None otherwise
        """
        try:
            # TODO: Implement GraphQL query to check existing instance
            # For now, simulate that instances don't exist (always create new)
            logger.debug(f"[IdempotentOps] Checking for existing instance: {instance_id}")
            return None
            
        except Exception as e:
            logger.warning(f"Failed to check existing instance {instance_id}: {e}")
            return None
    
    def generate_instance_id(self, session_id: str, task_name: str) -> str:
        """
        Generate a predictable instance ID for session-scoped instances.
        
        Args:
            session_id: Session ID
            task_name: Task name
            
        Returns:
            Generated instance ID
        """
        timestamp = int(time.time() * 1000)  # Millisecond precision
        return f"{session_id}:{task_name}:{timestamp}"
    
    def validate_session_ownership(self, instance_id: str, session_id: str) -> bool:
        """
        Validate that an instance belongs to the specified session.
        
        Args:
            instance_id: Instance ID to validate
            session_id: Expected session ID
            
        Returns:
            True if instance belongs to session
        """
        # For session-scoped instances, the session ID is part of the instance ID
        return instance_id.startswith(f"{session_id}:")


# GraphQL mutation templates (to be implemented)
UPSERT_TASK_DEFINITION_MUTATION = """
mutation UpsertTaskDefinition($input: UpsertTaskDefinitionInput!) {
    upsertTaskDefinition(input: $input) {
        taskName
        version
        status
        created
        timestamp
    }
}
"""

UPSERT_TASK_INSTANCE_MUTATION = """
mutation UpsertTaskInstance($input: UpsertTaskInstanceInput!) {
    upsertTaskInstance(input: $input) {
        id
        sessionId
        taskName
        taskVersion
        status
        createdAt
    }
}
"""

GET_TASK_INSTANCE_QUERY = """
query GetTaskInstance($instanceId: ID!) {
    getTaskInstance(id: $instanceId) {
        id
        sessionId
        taskName
        status
        createdAt
    }
}
""" 