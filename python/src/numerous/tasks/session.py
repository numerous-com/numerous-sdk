from typing import Dict, Optional, TYPE_CHECKING
from threading import local
import uuid
from .exceptions import SessionError # Import SessionError

# Conditional import for type hinting TaskInstance to avoid circular dependency
if TYPE_CHECKING:
    from .task import TaskInstance # noqa: F401 - Used for type hinting

# Thread-local storage for the current session context
_session_context = local()

class Session:
    """Manages a collection of task instances and provides a scope for their execution.

    Sessions are typically used to group tasks related to a specific user session,
    a web request, or a larger workflow. They help in tracking tasks and can
    be used by backends to enforce concurrency limits within the session's scope.
    """
    def __init__(self, name: Optional[str] = None, session_id: Optional[str] = None):
        self.name = name if name else f"session_{uuid.uuid4().hex[:8]}"
        self.id = session_id if session_id else str(uuid.uuid4())
        
        # Stores task instances: {task_definition_name: {instance_id: TaskInstance}}
        self.tasks: Dict[str, Dict[str, 'TaskInstance']] = {}
        self._is_active = False
        self._previous_session: Optional['Session'] = None

    @classmethod
    def current(cls) -> Optional['Session']:
        """Returns the currently active session in this context (thread)."""
        return getattr(_session_context, 'session', None)

    @classmethod
    def _set_current(cls, session: Optional['Session']) -> None:
        """Internal method to set the current session. Not for public use.
           Use the context manager (`with Session(...) as s:`) instead.
        """
        _session_context.session = session

    def __enter__(self) -> 'Session':
        """Enters the session context, making this session the active one."""
        if self._is_active:
            # TODO: Define specific SessionError in exceptions.py
            raise SessionError("Session is already active. Cannot re-enter.") 

        self._previous_session = Session.current()
        Session._set_current(self)
        self._is_active = True
        # TODO: Potentially notify backend that session has started if required
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Exits the session context, restoring the previous session if any."""
        if Session.current() is not self:
             # TODO: Define specific SessionError in exceptions.py
            raise SessionError(
                "Exiting a session that is not the current session. This indicates a bug."
            ) 
        Session._set_current(self._previous_session)
        self._is_active = False
        self._previous_session = None # Clear reference
        # TODO: Potentially notify backend that session has ended if required

    def add_task_instance(self, instance: 'TaskInstance') -> None:
        """Adds a task instance to this session for tracking."""
        task_def_name = instance.task_definition_name # Assuming TaskInstance has this property
        if task_def_name not in self.tasks:
            self.tasks[task_def_name] = {}
        self.tasks[task_def_name][instance.id] = instance

    def remove_task_instance(self, instance: 'TaskInstance') -> None:
        """Removes a task instance from this session (e.g., when it's done and reaped)."""
        task_def_name = instance.task_definition_name
        if task_def_name in self.tasks and instance.id in self.tasks[task_def_name]:
            del self.tasks[task_def_name][instance.id]
            if not self.tasks[task_def_name]: # Clean up empty task type dict
                del self.tasks[task_def_name]

    def get_running_instances_count(self, task_definition_name: str) -> int:
        """Counts currently running instances of a specific task type within this session."""
        if task_definition_name not in self.tasks:
            return 0
        return sum(
            1 for inst in self.tasks[task_definition_name].values()
            if inst.is_running # Assuming TaskInstance has an is_running property
        )

    # In a real implementation, Session.get(name) might interact with a backend
    # to retrieve or create a session that could be shared or managed by the backend.
    # For local mode, Session() constructor is enough for now. 