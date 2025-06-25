class TaskError(Exception):
    """Base class for exceptions related to Numerous Tasks."""
    pass

class MaxInstancesReachedError(TaskError):
    """Raised when attempting to start a task instance would exceed max_parallel."""
    pass

class SessionNotFoundError(TaskError):
    """Raised when a task operation requires an active session but none is found."""
    pass

class SessionError(TaskError):
    """Base class for errors related to Session management (e.g., re-entering active session)."""
    pass

class TaskCancelledError(TaskError):
    """Raised when an operation is attempted on a cancelled task, or result() is called."""
    pass

class BackendError(TaskError):
    """Raised when there is an issue with the task execution backend."""
    pass

class TaskDefinitionError(TaskError):
    """Raised when there is an issue with how a task is defined (e.g., signature)."""
    pass


class TaskExecutionConflictError(TaskError):
    """Raised when attempting to start a task execution that conflicts with an existing active execution."""
    
    def __init__(self, message: str, active_execution_id: str = None, conflicting_instance_id: str = None):
        super().__init__(message)
        self.active_execution_id = active_execution_id
        self.conflicting_instance_id = conflicting_instance_id


class SessionOwnershipError(TaskError):
    """Raised when a task instance does not belong to the specified session."""
    
    def __init__(self, message: str, session_id: str = None, instance_id: str = None):
        super().__init__(message)
        self.session_id = session_id
        self.instance_id = instance_id 