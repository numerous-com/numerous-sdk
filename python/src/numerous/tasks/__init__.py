# numerous.tasks - Define, launch, and interact with tasks

"""
This package provides the core functionality for defining, launching,
and interacting with Numerous Tasks.

Key Abstractions:
- @task: Decorator to define a Python function as a Numerous Task.
- TaskControl: Object injected into tasks for status reporting and control.
- Session: Manages the lifecycle and context of task instances.
- Future: Represents the result of an asynchronous task execution.
- ExecutionBackend: Abstract base class for task execution backends.
"""

from .task import task, Task, TaskInstance, TaskConfig
from .control import TaskControl, set_task_control_handler, LocalTaskControlHandler, PoCMockRemoteTaskControlHandler
from .session import Session
from .future import Future, LocalFuture, TaskStatus, TaskCancelledError
from .exceptions import (
    TaskError,
    MaxInstancesReachedError,
    BackendError,
    TaskDefinitionError,
    SessionNotFoundError,
    SessionError
)
from .backends import (
    ExecutionBackend,
    register_backend,
    get_backend,
    set_default_backend
)
from .backends.local import LocalExecutionBackend
from .backends.remote import RemoteExecutionBackend

# Local backend is registered by default.
_local_backend = LocalExecutionBackend()
register_backend("local", _local_backend)
set_default_backend("local")

__all__ = [
    # Core task definition and instance management
    "task",
    "Task",
    "TaskInstance",
    "TaskConfig",
    
    # Control and context
    "TaskControl",
    "Session",
    
    # Futures and status
    "Future",
    "LocalFuture",
    "TaskStatus",
    
    # Exceptions
    "TaskError",
    "MaxInstancesReachedError",
    "SessionNotFoundError",
    "SessionError",
    "TaskCancelledError",
    "BackendError",
    "TaskDefinitionError",

    # Backends & Handlers (Expose for extensibility / testing)
    "ExecutionBackend",
    "register_backend",
    "get_backend",
    "set_default_backend",
    "LocalExecutionBackend",
    "set_task_control_handler", # For runner or advanced setup
    "LocalTaskControlHandler",  # For type hinting or direct use if needed
    "PoCMockRemoteTaskControlHandler", # For PoC testing
    "RemoteExecutionBackend",
] 