from dataclasses import dataclass, field
from typing import Optional, Dict, Any
import logging
from abc import ABC, abstractmethod

# Get a logger specific to task control activities
# This helps differentiate task logs if users configure numerous.tasks logging.
task_logger = logging.getLogger("numerous.tasks.control")

# --- TaskControl Handlers ---
_current_task_control_handler: Optional['TaskControlHandler'] = None

class TaskControlHandler(ABC):
    """Abstract base class for TaskControl handlers.
    Defines how TaskControl operations (logging, status updates) are processed.
    """
    @abstractmethod
    def log(self, task_control: 'TaskControl', message: str, level: str, **extra_data: Any) -> None:
        pass

    @abstractmethod
    def update_progress(self, task_control: 'TaskControl', progress: float, status: Optional[str]) -> None:
        pass

    @abstractmethod
    def update_status(self, task_control: 'TaskControl', status: str) -> None:
        pass

    @abstractmethod
    def request_stop(self, task_control: 'TaskControl') -> None:
        # For local, it sets task_control.should_stop.
        # For remote, this might involve an API call if the handler manages the should_stop signal remotely.
        # However, typically tc.should_stop is set by an external actor on the TC object itself.
        # So the handler's role here might be more about propagating this request if needed.
        task_control._should_stop_internal = True # Default internal update

class LocalTaskControlHandler(TaskControlHandler):
    """Handles TaskControl operations for local execution (e.g., local logging)."""
    def log(self, task_control: 'TaskControl', message: str, level: str, **extra_data: Any) -> None:
        log_level = getattr(logging, level.upper(), logging.INFO)
        full_message = f"[Task {task_control.task_definition_name}/{task_control.instance_id}] {message}"
        task_logger.log(log_level, full_message, extra=extra_data)

    def update_progress(self, task_control: 'TaskControl', progress: float, status: Optional[str]) -> None:
        # In local mode, TaskControl directly updates its own fields.
        # The handler could additionally log this if desired.
        task_logger.debug(f"[Task {task_control.task_definition_name}/{task_control.instance_id}] Progress: {progress}%, Status: {status if status else 'N/A'}")

    def update_status(self, task_control: 'TaskControl', status: str) -> None:
        task_logger.debug(f"[Task {task_control.task_definition_name}/{task_control.instance_id}] Status updated: {status}")
    
    def request_stop(self, task_control: 'TaskControl') -> None:
        super().request_stop(task_control) # Sets _should_stop_internal
        task_logger.info(f"[Task {task_control.task_definition_name}/{task_control.instance_id}] Stop requested (via LocalHandler).")


class PoCMockRemoteTaskControlHandler(TaskControlHandler):
    """Handles TaskControl for PoC remote execution - prints to stdout."""
    def log(self, task_control: 'TaskControl', message: str, level: str, **extra_data: Any) -> None:
        print(f"[PoCMockRemoteTC][LOG][{level.upper()}] Task {task_control.task_definition_name}/{task_control.instance_id}: {message}", flush=True)
        if extra_data:
            print(f"  Extra data: {extra_data}", flush=True)

    def update_progress(self, task_control: 'TaskControl', progress: float, status: Optional[str]) -> None:
        print(f"[PoCMockRemoteTC][PROGRESS] Task {task_control.task_definition_name}/{task_control.instance_id}: Progress={progress:.2f}% Status='{status if status else ""}'", flush=True)

    def update_status(self, task_control: 'TaskControl', status: str) -> None:
        print(f"[PoCMockRemoteTC][STATUS] Task {task_control.task_definition_name}/{task_control.instance_id}: Status='{status}'", flush=True)

    def request_stop(self, task_control: 'TaskControl') -> None:
        super().request_stop(task_control) # Sets _should_stop_internal
        print(f"[PoCMockRemoteTC][STOP_REQUEST] Task {task_control.task_definition_name}/{task_control.instance_id}: Stop requested.", flush=True)

def get_task_control_handler() -> TaskControlHandler:
    global _current_task_control_handler
    if _current_task_control_handler is None:
        _current_task_control_handler = LocalTaskControlHandler()
    return _current_task_control_handler

def set_task_control_handler(handler: Optional[TaskControlHandler]) -> None:
    """Sets the global task control handler. 
    Pass None to reset to default (LocalTaskControlHandler).
    This should be called by the execution environment (e.g., task runner) at startup.
    """
    global _current_task_control_handler
    if handler is None:
        _current_task_control_handler = LocalTaskControlHandler() # Default back to local
    else:
        _current_task_control_handler = handler

# --- End TaskControl Handlers ---

@dataclass
class TaskControl:
    """Control object injected into a running task.
    Delegates operations to a configured TaskControlHandler.
    """
    instance_id: str = field(init=True)
    task_definition_name: str = field(init=True)
    # should_stop is now managed by the handler or set externally
    _should_stop_internal: bool = field(default=False, init=False, repr=False)
    progress: float = field(default=0.0, init=False) 
    status: str = field(default="", init=False)
    _handler: TaskControlHandler = field(init=False, repr=False)

    def __post_init__(self):
        self._handler = get_task_control_handler()

    @property
    def should_stop(self) -> bool:
        return self._should_stop_internal

    def request_stop(self) -> None:
        """Requests the task to stop. 
           Called by the SDK/Session, or backend runner for the specific instance.
        """
        # This method on TaskControl itself directly sets the flag for its own instance.
        # The handler's request_stop might be for a broader or different stop mechanism.
        self._should_stop_internal = True
        # Optionally, notify the handler that a stop was requested on this instance
        # self._handler.request_stop(self) # This might be redundant if handler.request_stop sets the flag
        # Let's assume external actors (like TaskInstance.stop() or runner) call this method directly on TC.
        # If the handler needs to *initiate* a stop, it can, but this is for reacting to a stop signal.
        if isinstance(self._handler, LocalTaskControlHandler): # Keep existing log for local
            self._handler.log(self, "Stop requested for instance.", "info")
        elif isinstance(self._handler, PoCMockRemoteTaskControlHandler):
             self._handler.log(self, "Stop requested for instance (will be checked by task loop).", "info")

    def update_progress(self, progress_val: float, status_msg: Optional[str] = None) -> None:
        if not 0.0 <= progress_val <= 100.0:
            pass 
        self.progress = progress_val # Update local state for immediate read if needed
        if status_msg is not None:
            self.status = status_msg # Update local state
        self._handler.update_progress(self, progress_val, status_msg)

    def update_status(self, status_msg: str) -> None:
        self.status = status_msg # Update local state
        self._handler.update_status(self, status_msg)

    def log(self, message: str, level: str = "info", **extra_data: Any) -> None:
        self._handler.log(self, message, level, **extra_data)

    # TODO: Consider removing the general TODO for structured logging as this method serves that purpose. 


def create_direct_execution_task_control(task_definition_name: str) -> TaskControl:
    """Create a TaskControl instance optimized for Stage 1 direct execution.
    
    Creates a TaskControl with a simple instance ID for direct function calls.
    
    Args:
        task_definition_name: Name of the task definition
        
    Returns:
        TaskControl instance for direct execution
    """
    # Use a simple counter-based ID for direct execution instead of UUID
    import time
    instance_id = f"direct_{int(time.time() * 1000000) % 1000000:06d}"
    
    return TaskControl(
        instance_id=instance_id,
        task_definition_name=task_definition_name
    ) 