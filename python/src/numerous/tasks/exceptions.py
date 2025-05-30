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