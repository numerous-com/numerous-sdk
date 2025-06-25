"""
API-connected backend for tasks.

This backend allows local task execution while using the remote API for:
- Fetching task inputs
- Reporting task outputs  
- Sending progress updates and logs

Activated automatically when API environment variables are detected.
"""

import os
import json
import base64
import logging
from typing import Dict, Any, Optional
from dataclasses import dataclass

from .control import TaskControlHandler, set_task_control_handler
from .exceptions import BackendError

logger = logging.getLogger(__name__)


@dataclass
class APIConfig:
    """Configuration for API backend connection."""
    api_url: str
    access_token: str
    organization_id: Optional[str] = None
    task_instance_id: Optional[str] = None
    
    @classmethod
    def from_environment(cls) -> Optional['APIConfig']:
        """Create API config from environment variables."""
        api_url = os.getenv('NUMEROUS_API_URL')
        access_token = os.getenv('NUMEROUS_API_ACCESS_TOKEN')
        
        if not api_url or not access_token:
            return None
            
        return cls(
            api_url=api_url,
            access_token=access_token,
            organization_id=os.getenv('NUMEROUS_ORGANIZATION_ID'),
            task_instance_id=os.getenv('NUMEROUS_TASK_INSTANCE_ID')
        )


class APITaskControlHandler(TaskControlHandler):
    """TaskControl handler that sends updates to remote API."""
    
    def __init__(self, api_client, task_instance_id: str):
        self.api_client = api_client
        self.task_instance_id = task_instance_id
    
    def log(self, task_control, message: str, level: str, **extra_data: Any) -> None:
        """Send log message to API."""
        try:
            # For now, log locally as well
            log_level = getattr(logging, level.upper(), logging.INFO)
            logger.log(log_level, f"[Task {self.task_instance_id}] {message}", extra=extra_data)
            
            # TODO: Send to API logging endpoint
            # self.api_client.send_log(self.task_instance_id, message, level, extra_data)
            
        except Exception as e:
            logger.error(f"Failed to send log to API: {e}")
    
    def update_progress(self, task_control, progress: float, status: Optional[str]) -> None:
        """Send progress update to API."""
        try:
            logger.debug(f"[Task {self.task_instance_id}] Progress: {progress}%, Status: {status}")
            
            # TODO: Send to API progress endpoint
            # self.api_client.update_progress(self.task_instance_id, progress, status)
            
        except Exception as e:
            logger.error(f"Failed to send progress to API: {e}")
    
    def update_status(self, task_control, status: str) -> None:
        """Send status update to API."""
        try:
            logger.debug(f"[Task {self.task_instance_id}] Status: {status}")
            
            # TODO: Send to API status endpoint  
            # self.api_client.update_status(self.task_instance_id, status)
            
        except Exception as e:
            logger.error(f"Failed to send status to API: {e}")
    
    def request_stop(self, task_control) -> None:
        """Handle stop request (local for now)."""
        super().request_stop(task_control)
        logger.info(f"[Task {self.task_instance_id}] Stop requested via API handler")


class APIConnectedBackend:
    """Backend that connects to remote API for data while executing locally."""
    
    def __init__(self, config: APIConfig):
        self.config = config
        self.api_client = None
        self._setup_api_client()
    
    def _setup_api_client(self):
        """Initialize API client."""
        try:
            from ..collections._get_client import get_client
            self.api_client = get_client()
            logger.info("API client initialized for connected backend")
        except Exception as e:
            logger.error(f"Failed to initialize API client: {e}")
            raise BackendError(f"Cannot connect to API: {e}")
    
    def fetch_task_inputs(self, task_instance_id: str) -> Dict[str, Any]:
        """Fetch task inputs from API."""
        if not self.api_client:
            raise BackendError("API client not initialized")
        
        try:
            # Use the same GraphQL query as the runner
            query = """
                query GetTaskInstance($id: ID!) {
                    getTaskInstance(id: $id) {
                        id
                        inputs
                        taskDefinitionName
                        status
                    }
                }
            """
            
            variables = {"id": task_instance_id}
            
            # Execute query using the client's event loop
            async def execute_query():
                response = await self.api_client._gql.execute(
                    query=query,
                    operation_name="GetTaskInstance", 
                    variables=variables,
                    headers=self.api_client._headers
                )
                return self.api_client._gql.get_data(response)
            
            result = self.api_client._loop.await_coro(execute_query())
            
            # Extract task instance data
            task_instance = result.get("getTaskInstance")
            if not task_instance:
                raise BackendError(f"Task instance {task_instance_id} not found")
            
            # Parse inputs from base64 JSON
            inputs_base64 = task_instance.get("inputs")
            if not inputs_base64:
                logger.info("No inputs provided for task instance")
                return {}
            
            inputs_json = base64.b64decode(inputs_base64).decode('utf-8')
            inputs_data = json.loads(inputs_json)
            
            logger.info(f"Fetched inputs for task {task_instance_id}: {list(inputs_data.keys())}")
            return inputs_data
            
        except Exception as e:
            logger.error(f"Failed to fetch task inputs: {e}")
            raise BackendError(f"Failed to fetch inputs: {e}")
    
    def report_task_result(self, task_instance_id: str, result: Any, error: Optional[Exception] = None):
        """Report task result to API."""
        if not self.api_client:
            raise BackendError("API client not initialized")
        
        try:
            status = "FAILED" if error else "COMPLETED"
            
            mutation = """
                mutation ReportTaskOutcome($input: ReportTaskOutcomeInput!) {
                    reportTaskOutcome(input: $input) {
                        id
                        status
                        completedAt
                    }
                }
            """
            
            mutation_input = {
                "taskInstanceId": task_instance_id,
                "status": status
            }
            
            # Add result or error
            if error:
                mutation_input["error"] = {
                    "errorType": type(error).__name__,
                    "message": str(error),
                    "traceback": None  # Could add traceback if needed
                }
            else:
                result_json = json.dumps(result)
                result_base64 = base64.b64encode(result_json.encode('utf-8')).decode('utf-8')
                mutation_input["result"] = result_base64
            
            variables = {"input": mutation_input}
            
            # Execute mutation
            async def execute_mutation():
                response = await self.api_client._gql.execute(
                    query=mutation,
                    operation_name="ReportTaskOutcome",
                    variables=variables,
                    headers=self.api_client._headers
                )
                return self.api_client._gql.get_data(response)
            
            result = self.api_client._loop.await_coro(execute_mutation())
            
            reported_task = result.get("reportTaskOutcome")
            if reported_task:
                logger.info(f"Successfully reported result for task {task_instance_id}")
            else:
                logger.warning(f"Unexpected response when reporting result for {task_instance_id}")
                
        except Exception as e:
            logger.error(f"Failed to report task result: {e}")
            raise BackendError(f"Failed to report result: {e}")
    
    def setup_task_control(self, task_instance_id: str):
        """Setup TaskControl handler for API communication."""
        handler = APITaskControlHandler(self.api_client, task_instance_id)
        set_task_control_handler(handler)
        logger.info(f"API TaskControl handler set for task {task_instance_id}")
    
    # Task Execution Layer Methods
    
    def upsert_task(self, name: str, version: str, function_name: str, module: str, parameters: dict = None, metadata: dict = None) -> dict:
        """Register or update a task definition (idempotent)."""
        if not self.api_client:
            raise BackendError("API client not initialized")
        
        try:
            mutation = """
                mutation UpsertTask($input: UpsertTaskInput!) {
                    upsertTask(input: $input) {
                        id
                        name
                        version
                        functionName
                        module
                        createdAt
                        updatedAt
                    }
                }
            """
            
            mutation_input = {
                "name": name,
                "version": version,
                "functionName": function_name,
                "module": module
            }
            
            if parameters:
                mutation_input["parameters"] = json.dumps(parameters)
            if metadata:
                mutation_input["metadata"] = json.dumps(metadata)
            
            variables = {"input": mutation_input}
            
            async def execute_mutation():
                response = await self.api_client._gql.execute(
                    query=mutation,
                    operation_name="UpsertTask",
                    variables=variables,
                    headers=self.api_client._headers
                )
                return self.api_client._gql.get_data(response)
            
            result = self.api_client._loop.await_coro(execute_mutation())
            return result.get("upsertTask")
            
        except Exception as e:
            logger.error(f"Failed to upsert task: {e}")
            raise BackendError(f"Failed to upsert task: {e}")
    
    def upsert_task_instance(self, instance_id: str, session_id: str, task_name: str, task_version: str, inputs: dict = None) -> dict:
        """Register or update a task instance (idempotent)."""
        if not self.api_client:
            raise BackendError("API client not initialized")
        
        try:
            mutation = """
                mutation UpsertTaskInstance($input: UpsertInstanceInput!) {
                    upsertInstance(input: $input) {
                        id
                        apiId
                        sessionId
                        taskDefinitionName
                        status
                        createdAt
                    }
                }
            """
            
            mutation_input = {
                "instanceId": instance_id,
                "sessionId": session_id,
                "taskName": task_name,
                "taskVersion": task_version
            }
            
            if inputs:
                inputs_json = json.dumps(inputs)
                inputs_base64 = base64.b64encode(inputs_json.encode('utf-8')).decode('utf-8')
                mutation_input["inputs"] = inputs_base64
            
            variables = {"input": mutation_input}
            
            async def execute_mutation():
                response = await self.api_client._gql.execute(
                    query=mutation,
                    operation_name="UpsertTaskInstance",
                    variables=variables,
                    headers=self.api_client._headers
                )
                return self.api_client._gql.get_data(response)
            
            result = self.api_client._loop.await_coro(execute_mutation())
            return result.get("upsertInstance")
            
        except Exception as e:
            logger.error(f"Failed to upsert task instance: {e}")
            raise BackendError(f"Failed to upsert task instance: {e}")
    
    def start_execution(self, task_instance_id: str, session_id: str, client_id: str, force: bool = False) -> dict:
        """Start a new task execution with conflict detection."""
        if not self.api_client:
            raise BackendError("API client not initialized")
        
        try:
            mutation = """
                mutation StartExecution($input: StartExecutionInput!) {
                    startExecution(input: $input) {
                        id
                        status
                        taskInstanceId
                        startedAt
                        isActive
                    }
                }
            """
            
            mutation_input = {
                "instanceId": task_instance_id,
                "sessionId": session_id,
                "force": force
            }
            
            variables = {"input": mutation_input}
            
            async def execute_mutation():
                response = await self.api_client._gql.execute(
                    query=mutation,
                    operation_name="StartExecution",
                    variables=variables,
                    headers=self.api_client._headers
                )
                return self.api_client._gql.get_data(response)
            
            result = self.api_client._loop.await_coro(execute_mutation())
            return result.get("startExecution")
            
        except Exception as e:
            if "ExecutionConflictError" in str(e):
                from .exceptions import TaskExecutionConflictError
                raise TaskExecutionConflictError(str(e))
            logger.error(f"Failed to start execution: {e}")
            raise BackendError(f"Failed to start execution: {e}")
    
    def force_start_execution(self, task_instance_id: str, session_id: str, client_id: str) -> dict:
        """Force start a new execution, killing any active ones."""
        return self.start_execution(task_instance_id, session_id, client_id, force=True)
    
    def check_execution_conflict(self, task_instance_id: str) -> dict:
        """Check if there are any execution conflicts for a task instance."""
        if not self.api_client:
            raise BackendError("API client not initialized")
        
        try:
            query = """
                query CheckExecutionConflict($instanceId: ID!) {
                    checkExecutionConflict(instanceId: $instanceId) {
                        conflictType
                        activeExecutionId
                        instanceId
                        message
                    }
                }
            """
            
            variables = {"instanceId": task_instance_id}
            
            async def execute_query():
                response = await self.api_client._gql.execute(
                    query=query,
                    operation_name="CheckExecutionConflict",
                    variables=variables,
                    headers=self.api_client._headers
                )
                return self.api_client._gql.get_data(response)
            
            result = self.api_client._loop.await_coro(execute_query())
            return result.get("checkExecutionConflict")
            
        except Exception as e:
            logger.error(f"Failed to check execution conflict: {e}")
            raise BackendError(f"Failed to check execution conflict: {e}")
    
    def force_stop_execution(self, execution_id: str, session_id: str, reason: str) -> dict:
        """Force stop an execution."""
        if not self.api_client:
            raise BackendError("API client not initialized")
        
        try:
            mutation = """
                mutation ForceStopExecution($executionId: ID!, $sessionId: ID!, $reason: String!) {
                    forceStopExecution(executionId: $executionId, sessionId: $sessionId, reason: $reason) {
                        id
                        status
                        completedAt
                        statusMessage
                    }
                }
            """
            
            variables = {
                "executionId": execution_id,
                "sessionId": session_id,
                "reason": reason
            }
            
            async def execute_mutation():
                response = await self.api_client._gql.execute(
                    query=mutation,
                    operation_name="ForceStopExecution",
                    variables=variables,
                    headers=self.api_client._headers
                )
                return self.api_client._gql.get_data(response)
            
            result = self.api_client._loop.await_coro(execute_mutation())
            return result.get("forceStopExecution")
            
        except Exception as e:
            logger.error(f"Failed to force stop execution: {e}")
            raise BackendError(f"Failed to force stop execution: {e}")
    
    def report_execution_progress(self, execution_id: str, progress: float, status_message: str = None) -> dict:
        """Report execution progress."""
        if not self.api_client:
            raise BackendError("API client not initialized")
        
        try:
            mutation = """
                mutation ReportExecutionProgress($input: ReportProgressInput!) {
                    reportExecutionProgress(input: $input) {
                        id
                        progress
                        statusMessage
                        updatedAt
                    }
                }
            """
            
            mutation_input = {
                "executionId": execution_id,
                "progress": progress
            }
            
            if status_message:
                mutation_input["message"] = status_message
            
            variables = {"input": mutation_input}
            
            async def execute_mutation():
                response = await self.api_client._gql.execute(
                    query=mutation,
                    operation_name="ReportExecutionProgress",
                    variables=variables,
                    headers=self.api_client._headers
                )
                return self.api_client._gql.get_data(response)
            
            result = self.api_client._loop.await_coro(execute_mutation())
            return result.get("reportExecutionProgress")
            
        except Exception as e:
            logger.error(f"Failed to report execution progress: {e}")
            raise BackendError(f"Failed to report execution progress: {e}")
    
    def complete_execution(self, execution_id: str, result: dict) -> dict:
        """Complete an execution with results."""
        if not self.api_client:
            raise BackendError("API client not initialized")
        
        try:
            mutation = """
                mutation CompleteExecution($executionId: ID!, $result: JSON!) {
                    completeExecution(executionId: $executionId, result: $result) {
                        id
                        status
                        result
                        completedAt
                    }
                }
            """
            
            result_json = json.dumps(result)
            
            variables = {
                "executionId": execution_id,
                "result": result_json
            }
            
            async def execute_mutation():
                response = await self.api_client._gql.execute(
                    query=mutation,
                    operation_name="CompleteExecution",
                    variables=variables,
                    headers=self.api_client._headers
                )
                return self.api_client._gql.get_data(response)
            
            result = self.api_client._loop.await_coro(execute_mutation())
            return result.get("completeExecution")
            
        except Exception as e:
            logger.error(f"Failed to complete execution: {e}")
            raise BackendError(f"Failed to complete execution: {e}")
    
    def fail_execution(self, execution_id: str, error: str) -> dict:
        """Fail an execution with error details."""
        if not self.api_client:
            raise BackendError("API client not initialized")
        
        try:
            mutation = """
                mutation FailExecution($executionId: ID!, $error: String!) {
                    failExecution(executionId: $executionId, error: $error) {
                        id
                        status
                        error
                        completedAt
                    }
                }
            """
            
            variables = {
                "executionId": execution_id,
                "error": error
            }
            
            async def execute_mutation():
                response = await self.api_client._gql.execute(
                    query=mutation,
                    operation_name="FailExecution",
                    variables=variables,
                    headers=self.api_client._headers
                )
                return self.api_client._gql.get_data(response)
            
            result = self.api_client._loop.await_coro(execute_mutation())
            return result.get("failExecution")
            
        except Exception as e:
            logger.error(f"Failed to fail execution: {e}")
            raise BackendError(f"Failed to fail execution: {e}")
    
    def validate_session_ownership(self, session_id: str, instance_id: str) -> bool:
        """Validate that a task instance belongs to the specified session."""
        if not self.api_client:
            raise BackendError("API client not initialized")
        
        try:
            query = """
                query ValidateSessionOwnership($sessionId: ID!, $instanceId: ID!) {
                    validateSessionOwnership(sessionId: $sessionId, instanceId: $instanceId)
                }
            """
            
            variables = {
                "sessionId": session_id,
                "instanceId": instance_id
            }
            
            async def execute_query():
                response = await self.api_client._gql.execute(
                    query=query,
                    operation_name="ValidateSessionOwnership",
                    variables=variables,
                    headers=self.api_client._headers
                )
                return self.api_client._gql.get_data(response)
            
            result = self.api_client._loop.await_coro(execute_query())
            return result.get("validateSessionOwnership", False)
            
        except Exception as e:
            if "SessionOwnershipError" in str(e):
                from .exceptions import SessionOwnershipError
                raise SessionOwnershipError(str(e), session_id, instance_id)
            logger.error(f"Failed to validate session ownership: {e}")
            raise BackendError(f"Failed to validate session ownership: {e}")
    
    def disconnect_client(self, client_id: str, session_id: str) -> dict:
        """Disconnect a client and kill its active executions."""
        if not self.api_client:
            raise BackendError("API client not initialized")
        
        try:
            mutation = """
                mutation DisconnectClient($clientId: String!, $sessionId: ID!) {
                    disconnectClient(clientId: $clientId, sessionId: $sessionId) {
                        killedExecutions
                        message
                    }
                }
            """
            
            variables = {
                "clientId": client_id,
                "sessionId": session_id
            }
            
            async def execute_mutation():
                response = await self.api_client._gql.execute(
                    query=mutation,
                    operation_name="DisconnectClient",
                    variables=variables,
                    headers=self.api_client._headers
                )
                return self.api_client._gql.get_data(response)
            
            result = self.api_client._loop.await_coro(execute_mutation())
            return result.get("disconnectClient")
            
        except Exception as e:
            logger.error(f"Failed to disconnect client: {e}")
            raise BackendError(f"Failed to disconnect client: {e}")
    
    def subscribe_to_execution_updates(self, execution_id: str):
        """Subscribe to real-time execution updates."""
        # TODO: Implement subscription functionality
        # This would use GraphQL subscriptions for real-time updates
        logger.warning("Subscription functionality not yet implemented")
        return []
    
    def subscribe_to_instance_updates(self, instance_id: str):
        """Subscribe to real-time task instance updates."""
        # TODO: Implement subscription functionality
        # This would use GraphQL subscriptions for real-time updates
        logger.warning("Subscription functionality not yet implemented")
        return []


# Global API backend instance
_api_backend: Optional[APIConnectedBackend] = None


def get_api_backend() -> Optional[APIConnectedBackend]:
    """Get the API backend instance if available."""
    global _api_backend
    
    if _api_backend is None:
        config = APIConfig.from_environment()
        if config:
            try:
                _api_backend = APIConnectedBackend(config)
                logger.info("API backend initialized from environment variables")
            except Exception as e:
                logger.error(f"Failed to initialize API backend: {e}")
                _api_backend = None
    
    return _api_backend


def is_api_mode() -> bool:
    """Check if API mode is enabled via environment variables."""
    return get_api_backend() is not None


def api_task_execution_wrapper(task_func, task_name: str, *args, **kwargs):
    """
    Wrapper that handles API-connected task execution.
    
    This function:
    1. Checks if API mode is enabled
    2. If yes, fetches inputs from API (ignoring passed args/kwargs)
    3. Executes task locally
    4. Reports results back to API
    """
    backend = get_api_backend()
    if not backend:
        # No API backend, execute normally
        return task_func(*args, **kwargs)
    
    # API mode enabled
    task_instance_id = backend.config.task_instance_id
    if not task_instance_id:
        logger.warning("API mode enabled but no NUMEROUS_TASK_INSTANCE_ID provided")
        return task_func(*args, **kwargs)
    
    try:
        # Setup API TaskControl handler
        backend.setup_task_control(task_instance_id)
        
        # Fetch inputs from API
        logger.info(f"Fetching inputs for task {task_name} (instance: {task_instance_id})")
        api_inputs = backend.fetch_task_inputs(task_instance_id)
        
        # Execute task with API inputs (ignore passed args/kwargs)
        logger.info(f"Executing task {task_name} with API inputs")
        result = task_func(**api_inputs)
        
        # Report success to API
        backend.report_task_result(task_instance_id, result)
        logger.info(f"Task {task_name} completed and result reported to API")
        
        return result
        
    except Exception as e:
        # Report error to API
        try:
            backend.report_task_result(task_instance_id, None, e)
        except:
            pass  # Don't let reporting errors mask the original error
        
        logger.error(f"Task {task_name} failed: {e}")
        raise 