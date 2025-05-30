from typing import Callable, Generic, TypeVar, Optional, Any, Dict
from dataclasses import dataclass, field
import functools
import uuid
import os
import inspect

from .control import TaskControl
from .session import Session
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
    def __init__(self, func: UserCallable, config: TaskConfig):
        functools.update_wrapper(self, func) # Copies name, docstring, etc.
        self._original_func: UserCallable = func
        self._internal_func: InternalCallable = self._wrap_original_func(func)
        self.config = config
        self.name = config.name if config.name else func.__name__

        # Store signature for validation and potential future use
        self._sig = inspect.signature(self._original_func)

    def _wrap_original_func(self, user_func: UserCallable) -> InternalCallable:
        """Wraps the user's function to inject TaskControl as the first argument."""
        @functools.wraps(user_func)
        def wrapped_func(task_control: TaskControl, *args: Any, **kwargs: Any) -> Any:
            return user_func(*args, **kwargs) # Pass through, TaskControl handled by backend
        
        # Actual injection will be handled by the backend. For now, this is a placeholder
        # conceptually. The _task_runner in TaskInstance will correctly pass TaskControl.
        return wrapped_func # type: ignore

    def instance(self) -> 'TaskInstance[UserCallable, R]':
        """Creates a new TaskInstance for this task definition.

        Requires an active Session context.
        """
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
        Consider changing this to always return a Future or TaskInstance for consistency.
        """
        if self.config.max_parallel != 1:
            raise TypeError(
                f"Task '{self.name}' has max_parallel > 1. "
                f"Use .instance().start() for explicit instance management and to get a Future."
            )
        
        # For max_parallel=1, create an implicit instance and run it synchronously for now.
        # This matches the pathfinder but might be revised based on review feedback.
        task_instance = self.instance()
        return task_instance.start(*args, **kwargs).result()
        
    @property
    def definition_name(self) -> str:
        return self.name


class TaskInstance(Generic[UserCallable, R]):
    """Represents a single execution instance of a Task.
    """
    def __init__(self, task_definition: Task[UserCallable, R], session: Session):
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
        with the TaskControl object.
        """
        # This is where the TaskControl is explicitly passed to the user's code.
        # The self.task_definition._original_func is the user's actual code.
        try:
            # We need to call the user's original function with task_control and other args.
            # The _internal_func wrapper was more conceptual for the Task definition.
            
            # Bind arguments to the original function's signature to respect defaults, etc.
            # and to separate args from kwargs if the user function expects that.
            bound_args = self.task_definition._sig.bind_partial(*args, **kwargs)
            
            # The TaskControl is the first *logical* argument to the user's task
            # as they define it (e.g. def my_task(tc: TaskControl, x, y)).
            # However, the @task decorator hides this from the *call signature*.
            # So, when calling the _original_func, it does NOT expect TaskControl directly.
            # TaskControl is provided by the execution environment (backend).
            
            # The current pathfinder approach: _task_wrapper in old task.py took func, task_control
            # This means the backend.execute needs to be aware of this convention.
            
            # For now, let's assume the backend.execute will handle providing TaskControl
            # to the original function or a wrapped version of it.
            # So, this _task_runner will directly call original_func, assuming TaskControl is handled.
            # This aligns with the Task._wrap_original_func. 
            # OR, the backend execute calls THIS method, and this method provides tc. 
            # The latter seems more direct for local backend. 

            # Let's try this: the `_task_runner` IS the function the backend executes.
            # It receives args/kwargs meant for the *user's function*.
            # It then calls the *user's function* prepending the `task_control`.
            
            return self.task_definition._original_func(self.task_control, *args, **kwargs)
        except Exception as e:
            # The future should capture this exception
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
    config = TaskConfig(name=name, max_parallel=max_parallel, size=size)
    
    def decorator(func_to_wrap: UserCallable) -> Task[UserCallable, Any]:
        # Validate that the first argument of func_to_wrap is type-hinted as TaskControl
        sig = inspect.signature(func_to_wrap)
        params = list(sig.parameters.values())
        if not params or params[0].name == 'self' or params[0].name == 'cls': # Skip self/cls for methods
            if len(params) < 2 or params[1].annotation != TaskControl:
                 raise TypeError(
                    f"Task function '{func_to_wrap.__name__}' must have 'TaskControl' as its first "
                    f"argument (after self/cls if a method). Example: def {func_to_wrap.__name__}(tc: TaskControl, ...)."
                )
        elif params[0].annotation != TaskControl:
            raise TypeError(
                f"Task function '{func_to_wrap.__name__}' must have 'TaskControl' as its first "
                f"argument. Example: def {func_to_wrap.__name__}(tc: TaskControl, ...)."
            )

        # Modify the user's function signature to remove TaskControl for the Task object's __call__.
        # This is tricky because the actual call to the original function *must* include TaskControl.
        # The Task object itself, when called, should reflect the user's intended signature.
        
        # The Task object will store the original func and handle TaskControl injection.
        return Task(func_to_wrap, config)

    if _func is None:
        return decorator # Decorator called with parentheses: @task(...)
    else:
        return decorator(_func) # Decorator called without parentheses: @task 