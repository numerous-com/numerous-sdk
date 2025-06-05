from typing import Dict, Optional, Type, Any, Callable
from abc import ABC, abstractmethod

# Forward declaration for type hinting
if False: # TYPE_CHECKING
    from ..future import Future
    from ..task import TaskInstance # Not directly used here, but good for context

from .local import LocalExecutionBackend
from .remote import RemoteExecutionBackend # Import the new backend

_backends: Dict[str, 'ExecutionBackend'] = {}
_default_backend_name: Optional[str] = None

class ExecutionBackend(ABC):
    """Abstract base class for task execution backends."""

    @abstractmethod
    def execute(
        self, 
        target_callable: Callable[..., Any], # The function to execute (e.g., TaskInstance._task_runner)
        future: 'Future',                  # The future object to be updated by the backend
        args: tuple, 
        kwargs: dict
    ) -> None:
        """Execute the target_callable.

        This method should arrange for the callable to be run, and upon its completion
        (or failure), it must update the provided `future` object with the result
        or exception.

        For local backends (like a thread pool), this might involve submitting the
        callable to the pool and attaching callbacks to update the future.
        For remote backends, this would involve serializing the call, sending it to
        a remote service, and then managing the future based on responses from that service.
        """
        pass

    # @abstractmethod
    # def get_session(self, name: str) -> 'Session': # Session might be backend-specific
    #     """Get or create a session managed by this backend."""
    #     pass

    @abstractmethod
    def cancel_task_instance(self, instance_id: str, session_id: Optional[str] = None) -> bool:
        """Attempt to cancel a task instance managed by this backend."""
        pass

    def startup(self) -> None:
        """Called when the backend is initialized, for any setup."""
        pass

    def shutdown(self) -> None:
        """Called when the SDK or application is shutting down, for cleanup."""
        pass


def register_backend(name: str, backend_instance: ExecutionBackend) -> None:
    """Registers an execution backend instance."""
    global _default_backend_name
    if name in _backends:
        # TODO: logging.warning(f"Backend '{name}' is being overridden.")
        pass 
    _backends[name] = backend_instance
    if _default_backend_name is None:
        _default_backend_name = name

def get_backend(name: Optional[str] = None) -> Optional[ExecutionBackend]:
    """Retrieves a registered backend instance.

    If name is None, returns the default backend (the first one registered, or 'local').
    """
    if name is None:
        name = _default_backend_name
    if name is None and 'local' in _backends: # Fallback to local if no default set explicitly
        name = 'local'
    
    backend = _backends.get(name)
    if backend is None and name == 'local':
        # Auto-initialize local backend if requested and not yet registered
        local_backend = LocalExecutionBackend()
        local_backend.startup()
        register_backend('local', local_backend)
        return local_backend
        
    return backend

def set_default_backend(name: str) -> None:
    """Sets the default backend to be used if no specific backend is requested."""
    global _default_backend_name
    if name not in _backends:
        # If trying to set a non-registered backend as default, and it's 'local', initialize it.
        if name == 'local':
            get_backend('local') # This will initialize and register the local backend
        else:
            # TODO: Define specific BackendError in exceptions.py
            raise ValueError(f"Backend '{name}' not registered. Cannot set as default.")
    _default_backend_name = name

# Auto-register the local backend by default when this module is imported.
# This ensures 'local' is always available unless explicitly overridden or removed.
# Deferring actual instantiation until get_backend('local') is called. 

__all__ = [
    "ExecutionBackend",
    "register_backend",
    "get_backend",
    "set_default_backend",
    "LocalExecutionBackend",
    "RemoteExecutionBackend", # Add to __all__
] 