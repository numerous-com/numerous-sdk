"""
RemoteTaskControlHandler for Task 4.2: Integration Testing with Local API.

This handler communicates with the Numerous API using the existing GraphQL client
infrastructure to report task progress, status updates, and logging.

Updated for Task 4.6: Includes idempotent task and instance registration with local versioning.
"""

import logging
import os
import time
import asyncio
from typing import Optional, Any
from dataclasses import dataclass

from numerous.tasks.control import TaskControlHandler, TaskControl
from numerous.tasks.versioning import get_task_version, extract_task_definition
from .idempotent_operations import IdempotentOperations, TaskDefinition

logger = logging.getLogger(__name__)


@dataclass
class TaskExecutionLog:
    """Data structure for task execution logs sent to API."""
    task_instance_id: str
    task_name: str
    message: str
    level: str
    timestamp: float
    extra_data: dict


@dataclass
class TaskProgressUpdate:
    """Data structure for task progress updates sent to API."""
    task_instance_id: str
    task_name: str
    progress: float
    status: Optional[str] = None


class RemoteTaskControlHandler(TaskControlHandler):
    """
    TaskControl handler that communicates with the Numerous API.
    
    This handler uses the existing GraphQL client infrastructure to:
    - Send log messages to the API
    - Report progress and status updates
    - Handle stop requests via API communication
    
    IMPORTANT: This handler requires a working API connection and will fail
    immediately if the API is not available. Use LocalExecutionBackend or
    a different handler if you want local-only execution.
    
    For integration testing, assumes API is available via NUMEROUS_API_URL
    environment variable (defaults to localhost:8080 for local testing).
    """
    
    def __init__(self, session_id: Optional[str] = None):
        """
        Initialize the remote handler.
        
        Args:
            session_id: Optional session ID for grouping task executions
        """
        self.session_id = session_id or f"session_{int(time.time())}"
        self._client = None
        self._idempotent_ops = None
        self._registered_tasks = set()  # Track registered tasks to avoid duplicate registration
        self._initialize_client()
    
    def _initialize_client(self):
        """Initialize the GraphQL client using existing organization client."""
        try:
            # Use existing centralized client (follows SDK patterns)
            from numerous.organization import get_client
            self._client = get_client()
            
            # Initialize idempotent operations
            self._idempotent_ops = IdempotentOperations(self._client)
            
            logger.info(f"RemoteTaskControlHandler initialized with organization client (session: {self.session_id})")
        except Exception as e:
            logger.error(f"Failed to initialize GraphQL client: {e}")
            raise RuntimeError(
                f"RemoteTaskControlHandler requires API connection but connection failed: {e}. "
                "Ensure the Numerous API is running and accessible, or use a different task control handler."
            ) from e
    
    @property
    def is_connected(self) -> bool:
        """Check if handler is connected to API."""
        return self._client is not None
    
    async def ensure_task_registered(self, task_control: TaskControl) -> None:
        """
        Ensure task definition is registered with the API using idempotent operations.
        
        Args:
            task_control: TaskControl instance containing task information
            
        Raises:
            RuntimeError: If task registration fails
        """
        task_name = task_control.task_definition_name
        
        # Skip if already registered in this session
        if task_name in self._registered_tasks:
            return
        
        try:
            # Get task function from TaskControl if available
            task_func = getattr(task_control, '_task_func', None)
            explicit_version = getattr(task_control, '_explicit_version', None)
            
            if task_func:
                # Extract task definition and generate version
                task_definition_dict = extract_task_definition(task_func)
                version = get_task_version(task_func, explicit_version)
            else:
                # Fallback for cases where task function is not available
                task_definition_dict = {
                    "function_name": task_name,
                    "module": "unknown",
                    "doc": "",
                    "parameters": {}
                }
                version = f"local-{hash(task_name) % 100000000:08x}"
            
            # Create TaskDefinition object
            task_definition = TaskDefinition(
                name=task_name,
                version=version,
                function_name=task_definition_dict["function_name"],
                module=task_definition_dict["module"],
                parameters=task_definition_dict.get("parameters", {}),
                metadata={
                    "doc": task_definition_dict.get("doc", ""),
                    "max_parallel": task_definition_dict.get("max_parallel"),
                    "resource_sizing": task_definition_dict.get("resource_sizing"),
                    "timeout": task_definition_dict.get("timeout"),
                }
            )
            
            # Register task definition with API
            result = await self._idempotent_ops.upsert_task_definition(task_name, task_definition)
            
            # Mark as registered
            self._registered_tasks.add(task_name)
            
            logger.info(f"[RemoteTC] Task {task_name} registered with version {version}")
            
        except Exception as e:
            logger.error(f"Failed to register task {task_name}: {e}")
            raise RuntimeError(f"Failed to register task {task_name}: {e}") from e
    
    async def ensure_instance_registered(self, task_control: TaskControl) -> str:
        """
        Ensure task instance is registered with the API using idempotent operations.
        
        Args:
            task_control: TaskControl instance
            
        Returns:
            Instance ID
            
        Raises:
            RuntimeError: If instance registration fails
        """
        try:
            # Generate session-scoped instance ID
            instance_id = self._idempotent_ops.generate_instance_id(
                self.session_id, 
                task_control.task_definition_name
            )
            
            # Get task version (same logic as in ensure_task_registered)
            task_func = getattr(task_control, '_task_func', None)
            explicit_version = getattr(task_control, '_explicit_version', None)
            
            if task_func:
                version = get_task_version(task_func, explicit_version)
            else:
                version = f"local-{hash(task_control.task_definition_name) % 100000000:08x}"
            
            # Register instance with API
            instance = await self._idempotent_ops.upsert_task_instance(
                instance_id=instance_id,
                session_id=self.session_id,
                task_name=task_control.task_definition_name,
                task_version=version
            )
            
            logger.info(f"[RemoteTC] Instance {instance_id} registered for task {task_control.task_definition_name}")
            return instance_id
            
        except Exception as e:
            logger.error(f"Failed to register instance for task {task_control.task_definition_name}: {e}")
            raise RuntimeError(f"Failed to register instance: {e}") from e
    
    def _run_async(self, coro):
        """Helper to run async operations in sync context."""
        try:
            loop = asyncio.get_event_loop()
        except RuntimeError:
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)
        
        return loop.run_until_complete(coro)
    
    def log(self, task_control: TaskControl, message: str, level: str, **extra_data: Any) -> None:
        """
        Send log message to API.
        
        Args:
            task_control: TaskControl instance
            message: Log message
            level: Log level (debug, info, warning, error)
            **extra_data: Additional data to include with log
            
        Raises:
            RuntimeError: If API communication fails
        """
        import time
        
        try:
            # Ensure task and instance are registered before logging
            self._run_async(self.ensure_task_registered(task_control))
            
            # Ensure instance is registered and get instance ID
            if not hasattr(task_control, '_remote_instance_id'):
                instance_id = self._run_async(self.ensure_instance_registered(task_control))
                task_control._remote_instance_id = instance_id
            else:
                instance_id = task_control._remote_instance_id
            
            # Create log entry
            log_entry = TaskExecutionLog(
                task_instance_id=instance_id,
                task_name=task_control.task_definition_name,
                message=message,
                level=level,
                timestamp=time.time(),
                extra_data=extra_data
            )
            
            # TODO: Send to API using GraphQL mutation
            # For now, log locally with API prefix to simulate API communication
            logger.info(f"[RemoteTC][LOG] Sending to API: {log_entry}")
            
        except Exception as e:
            logger.error(f"Failed to send log to API: {e}")
            raise RuntimeError(f"RemoteTaskControlHandler failed to send log to API: {e}") from e
    
    def update_progress(self, task_control: TaskControl, progress: float, status: Optional[str]) -> None:
        """
        Send progress update to API.
        
        Args:
            task_control: TaskControl instance
            progress: Progress percentage (0.0-100.0)
            status: Optional status message
            
        Raises:
            RuntimeError: If API communication fails
        """
        try:
            # Ensure task and instance are registered before updating progress
            self._run_async(self.ensure_task_registered(task_control))
            
            # Ensure instance is registered and get instance ID
            if not hasattr(task_control, '_remote_instance_id'):
                instance_id = self._run_async(self.ensure_instance_registered(task_control))
                task_control._remote_instance_id = instance_id
            else:
                instance_id = task_control._remote_instance_id
            
            # Create progress update
            progress_update = TaskProgressUpdate(
                task_instance_id=instance_id,
                task_name=task_control.task_definition_name,
                progress=progress,
                status=status
            )
            
            # TODO: Send to API using GraphQL mutation
            # For now, log locally with API prefix to simulate API communication
            logger.debug(f"[RemoteTC][PROGRESS] Sending to API: {progress_update}")
            
        except Exception as e:
            logger.error(f"Failed to send progress update to API: {e}")
            raise RuntimeError(f"RemoteTaskControlHandler failed to send progress update to API: {e}") from e
    
    def update_status(self, task_control: TaskControl, status: str) -> None:
        """
        Send status update to API.
        
        Args:
            task_control: TaskControl instance
            status: New status message
            
        Raises:
            RuntimeError: If API communication fails
        """
        try:
            # Ensure task and instance are registered before updating status
            self._run_async(self.ensure_task_registered(task_control))
            
            # Ensure instance is registered and get instance ID
            if not hasattr(task_control, '_remote_instance_id'):
                instance_id = self._run_async(self.ensure_instance_registered(task_control))
                task_control._remote_instance_id = instance_id
            else:
                instance_id = task_control._remote_instance_id
            
            # TODO: Send to API using GraphQL mutation
            # For now, log locally with API prefix to simulate API communication
            logger.debug(f"[RemoteTC][STATUS] Sending to API - Task {task_control.task_definition_name}/{instance_id}: {status}")
            
        except Exception as e:
            logger.error(f"Failed to send status update to API: {e}")
            raise RuntimeError(f"RemoteTaskControlHandler failed to send status update to API: {e}") from e
    
    def request_stop(self, task_control: TaskControl) -> None:
        """
        Handle stop request via API communication.
        
        Args:
            task_control: TaskControl instance
            
        Raises:
            RuntimeError: If API communication fails
        """
        # Always set the internal stop flag first
        super().request_stop(task_control)
        
        try:
            # TODO: Send stop request to API
            # For now, log locally with API prefix to simulate API communication
            logger.info(f"[RemoteTC][STOP] Sending to API - Stop requested for task {task_control.task_definition_name}/{task_control.instance_id}")
            
        except Exception as e:
            logger.error(f"Failed to send stop request to API: {e}")
            raise RuntimeError(f"RemoteTaskControlHandler failed to send stop request to API: {e}") from e
    
    def check_stop_requested(self, task_control: TaskControl) -> bool:
        """
        Check with API if stop has been requested for this task.
        
        Args:
            task_control: TaskControl instance
            
        Returns:
            True if stop has been requested
            
        Raises:
            RuntimeError: If API communication fails
        """
        try:
            # TODO: Query API for stop status
            # For now, return local state (simulating API query)
            return task_control.should_stop
            
        except Exception as e:
            logger.error(f"Failed to check stop status with API: {e}")
            raise RuntimeError(f"RemoteTaskControlHandler failed to check stop status with API: {e}") from e 