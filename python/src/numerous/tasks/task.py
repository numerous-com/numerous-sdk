from typing import Callable, Generic, TypeVar, Optional, Any, Dict, TYPE_CHECKING
from dataclasses import dataclass, field
import functools
import uuid
import os
import inspect
import warnings

if TYPE_CHECKING:
    from .session import Session

from .control import TaskControl
from .future import Future, LocalFuture # For local execution
from .exceptions import TaskError, MaxInstancesReachedError, SessionNotFoundError
from .backends import get_backend # Will be used later

# Define a generic type for the result of the task function
R = TypeVar('R') # Result type

@dataclass
class TaskConfig:
    """Configuration for a Task definition."""
    max_parallel: int = 1
    size: str = "small"  # Example: small, medium, large - for backend resource hint
    name: Optional[str] = None # Optional explicit name for the task
    # TODO: Add other config options: timeout, retries, etc.

# This is the user-facing function signature (without TaskControl)
UserCallable = TypeVar('UserCallable', bound=Callable[..., Any])

# This is the internal function signature (with TaskControl as first arg)
InternalCallable = Callable[..., R]

class Task(Generic[UserCallable, R]):
    """Represents a callable Numerous Task definition.

    This class is typically created by the @task decorator.
    It wraps the user's original function and handles the injection of TaskControl,
    instance creation, and interaction with the execution backend.
    """
    def __init__(self, func: UserCallable, config: TaskConfig, expects_task_control: bool):
        functools.update_wrapper(self, func) # Copies name, docstring, etc.
        self._original_func: UserCallable = func
        self.config = config
        self.name = config.name if config.name else func.__name__
        self.expects_task_control = expects_task_control # Store the flag

        # Store signature for validation and potential future use
        self._sig = inspect.signature(self._original_func)

    def _wrap_original_func(self, user_func: UserCallable) -> InternalCallable:
        """Wraps the user's function. The actual TaskControl injection is conditional in _task_runner."""
        @functools.wraps(user_func)
        def wrapped_func_placeholder(task_control_placeholder: TaskControl, *args: Any, **kwargs: Any) -> Any:
            # This wrapper's signature might be misleading if not used carefully.
            # The core logic is now in _task_runner.
            if self.expects_task_control:
                return user_func(task_control_placeholder, *args, **kwargs)
            return user_func(*args, **kwargs)
        # The type InternalCallable implies TaskControl is always first.
        # This might need a more complex type or a re-evaluation of _internal_func's purpose.
        # For now, _task_runner directly calls _original_func.
        return user_func # type: ignore # Returning original_func directly if _internal_func is not strictly used by backend

    def _execute_direct(self, *args: Any, **kwargs: Any) -> R:
        """Execute the task directly as a function call without backend infrastructure.
        
        This is used for Stage 1 direct execution.
        Creates a lightweight TaskControl if needed and executes the function directly.
        """
        if self.expects_task_control:
            # Create a lightweight TaskControl for direct execution
            from .control import create_direct_execution_task_control
            task_control = create_direct_execution_task_control(self.name)
            return self._original_func(task_control, *args, **kwargs)
        else:
            # Execute function directly without TaskControl
            return self._original_func(*args, **kwargs)

    def instance(self) -> 'TaskInstance[UserCallable, R]':
        """Creates a new TaskInstance for this task definition.

        Requires an active Session context.
        """
        from .session import Session
        current_session = Session.current()
        if not current_session:
            raise SessionNotFoundError(
                "Cannot create a task instance without an active session. "
                "Use 'with Session() as session:' or ensure a session is active."
            )
        return TaskInstance(self, current_session)

    def __call__(self, *args: Any, **kwargs: Any) -> R:
        """Allows calling the task definition directly.

        If max_parallel=1, this will start the task and block for its result.
        If max_parallel > 1, this will raise a TypeError suggesting to use .instance().start().
        
        For Stage 1 direct execution: If no Session is active, executes directly as a function call.
        If a Session is active, uses TaskInstance for proper tracking and backend execution.
        """
        if self.config.max_parallel != 1:
            raise TypeError(
                f"Task '{self.name}' has max_parallel > 1. "
                f"Use .instance().start() for explicit instance management and to get a Future."
            )
        
        # Check if there's an active session and warn about direct execution
        from .session import Session
        current_session = Session.current()
        
        if current_session is not None:
            warnings.warn(
                f"Task '{self.name}' is executing directly despite active session "
                f"'{current_session.name}'. Direct execution bypasses session tracking "
                f"and task management. Use task.instance().start() for session-managed execution.",
                UserWarning,
                stacklevel=2
            )
        
        # For Stage 1, always use direct execution for simplicity and performance
        # In later stages, we can add logic to conditionally use TaskInstance when needed
        return self._execute_direct(*args, **kwargs)
        
    @property
    def definition_name(self) -> str:
        return self.name


class TaskInstance(Generic[UserCallable, R]):
    """Represents a single execution instance of a Task.
    """
    def __init__(self, task_definition: Task[UserCallable, R], session: 'Session'):
        self.task_definition = task_definition
        self.session = session
        self.id: str = str(uuid.uuid4()) # Unique ID for this instance
        self.task_control = TaskControl(instance_id=self.id, task_definition_name=self.task_definition.definition_name)
        self._future: Optional[LocalFuture[R]] = None # For now, only LocalFuture
        self._status: str = "pending" # TODO: Use TaskStatus enum from future.py
        
        self.session.add_task_instance(self) # Register with session

        # Get backend from environment variable or session/global config
        # For Phase 1, we'll default to a local backend directly.
        backend_name = os.environ.get("NUMEROUS_TASK_BACKEND", "local")
        self.backend = get_backend(backend_name)
        if self.backend is None: # Should not happen if local is default
            raise TaskError(f"Backend '{backend_name}' not found or configured.")

    @property
    def task_definition_name(self) -> str:
        return self.task_definition.definition_name

    @property
    def is_running(self) -> bool:
        return self._future is not None and not self._future.done
    
    @property
    def status(self) -> str:
        if self._future:
            return self._future.status
        return self._status # Initial status before future is created

    def _task_runner(self, *args: Any, **kwargs: Any) -> R:
        """Internal method that actually executes the task's original function
        with the TaskControl object if expected.
        
        This method is called by the execution backend and handles TaskControl injection.
        """
        try:
            if self.task_definition.expects_task_control:
                # Inject TaskControl as the first argument
                return self.task_definition._original_func(self.task_control, *args, **kwargs)
            else:
                # Execute function directly without TaskControl injection
                return self.task_definition._original_func(*args, **kwargs)
        except Exception as e:
            # Let the exception propagate to be handled by the backend/future
            raise

    def start(self, *args: Any, **kwargs: Any) -> Future[R]:
        """Starts the task execution.

        Returns a Future object that can be used to get the result.
        """
        if self._future is not None:
            raise TaskError("Task instance has already been started.")

        # Check concurrency limits within the session
        running_count = self.session.get_running_instances_count(self.task_definition_name)
        if running_count >= self.task_definition.config.max_parallel:
            raise MaxInstancesReachedError(
                f"Task '{self.task_definition_name}': cannot start new instance. "
                f"Max parallel ({self.task_definition.config.max_parallel}) reached for this session."
            )
        
        self._status = "starting" # TODO: Use TaskStatus enum
        self._future = LocalFuture() # For now, create LocalFuture directly.
                                    # Later, self.backend.execute(...) will return a Future.
        
        # The backend's execute method is responsible for running the task.
        # It should run `self._task_runner` with `*args` and `**kwargs`.
        # and manage setting the future's result or exception.
        try:
            self.backend.execute(
                target_callable=self._task_runner, 
                future=self._future, # Pass the future for the backend to update
                args=args, 
                kwargs=kwargs
            )
            self._status = self._future.status # Update status from future after backend starts it
        except Exception as e:
            self._future.set_exception(e) # Ensure future reflects failure if backend.execute fails
            self._status = self._future.status
            self.session.remove_task_instance(self) # Clean up from session if start fails badly
            raise
        
        return self._future

    def stop(self) -> None:
        """Requests the task to stop gracefully."""
        self.task_control.request_stop()
        # TODO: If future is present and backend supports it, call future.cancel() or backend.cancel(self.id)
        # For local thread backend, setting should_stop is the primary mechanism.

    @property
    def result(self) -> R:
        """Synchronously gets the result of the task execution.
        Blocks until the task is complete.
        """
        if self._future is None:
            raise TaskError("Task has not been started yet.")
        return self._future.result()

# The @task decorator
def task(
    _func: Optional[UserCallable] = None, 
    *, 
    name: Optional[str] = None,
    max_parallel: int = 1,
    size: str = "small",
    # ... other TaskConfig parameters ...
) -> Callable[[UserCallable], Task[UserCallable, Any]] | Task[UserCallable, Any]:
    """Decorator to define a Python function as a Numerous Task.

    Args (decorator parameters):
        name: Optional explicit name for the task. Defaults to the function name.
        max_parallel: Max number of concurrent instances in a session (default: 1).
        size: Resource size hint for the backend (e.g., "small", "medium", "large").

    Returns:
        A Task object that can be used to create and start instances.
    """
    task_config = TaskConfig(name=name, max_parallel=max_parallel, size=size)
    
    def decorator(func_to_wrap: UserCallable) -> Task[UserCallable, Any]:
        expects_task_control = False
        sig = inspect.signature(func_to_wrap)
        params = list(sig.parameters.values())

        if params: # Function has parameters
            first_actual_param_index = 0
            # Check if it's a method (first param is 'self' or 'cls')
            if params[0].name in ('self', 'cls'):
                if len(params) > 1: # Method has more params after self/cls
                    first_actual_param_index = 1
                else: # Method only has self/cls, so no other params to check for TaskControl
                    params = [] # Treat as if no checkable params for TaskControl

            # Check if a parameter exists at first_actual_param_index and if its annotation is TaskControl
            if params and first_actual_param_index < len(params):
                # Ensure the parameter at first_actual_param_index has an annotation
                if params[first_actual_param_index].annotation == TaskControl:
                    expects_task_control = True
        
        # The old validation raising TypeError is removed.
        # TaskControl is now optionally injected based on expects_task_control.
        
        return Task(func_to_wrap, task_config, expects_task_control)

    if _func is None: # Decorator called with parentheses: @task(...)
        return decorator
    else: # Decorator called without parentheses: @task
        return decorator(_func) 