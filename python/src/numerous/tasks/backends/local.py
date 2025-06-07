import threading
from concurrent.futures import ThreadPoolExecutor, Future as ConcurrentFuture
from typing import Callable, Any, Optional, Dict, Set
import logging
from abc import ABC, abstractmethod
from ..future import LocalFuture, TaskStatus # Assuming LocalFuture is in ..future
from ..control import TaskControl # Assuming TaskControl is in ..control

logger = logging.getLogger(__name__)

# Forward declaration for type hinting
if False: # TYPE_CHECKING
    from ..future import Future
    from ..task import TaskInstance # Not directly used here, but good for context

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

class LocalExecutionBackend(ExecutionBackend):
    """Executes tasks locally using a thread pool."""

    def __init__(self, max_workers: Optional[int] = None):
        """
        Args:
            max_workers: Maximum number of threads in the pool. 
                         Defaults to None (Python's default, usually CPU count * 5).
        """
        self._executor = ThreadPoolExecutor(max_workers=max_workers)
        # Maps task_instance.id to the concurrent.futures.Future object
        self._active_concurrent_futures: Dict[str, ConcurrentFuture] = {}
        # Maps task_instance.id to its TaskControl object
        self._active_task_controls: Dict[str, TaskControl] = {}
        self._lock = threading.Lock()
        logger.info(f"LocalExecutionBackend initialized with max_workers={max_workers or 'default'}")

    def execute(
        self, 
        target_callable: Callable[..., Any], # This is TaskInstance._task_runner
        future: LocalFuture, # This is the numerous.tasks.future.LocalFuture
        args: tuple, 
        kwargs: dict
    ) -> None:
        """Submits the task to the thread pool for execution."""
        
        if not isinstance(future, LocalFuture):
            # TODO: use specific BackendError from exceptions.py
            raise TypeError("LocalExecutionBackend requires a LocalFuture object.")

        # target_callable is TaskInstance._task_runner, which is a bound method.
        # target_callable.__self__ gives us the TaskInstance object.
        task_instance = getattr(target_callable, '__self__', None)
        if task_instance is None or not hasattr(task_instance, 'id') or not hasattr(task_instance, 'task_control'):
            # This should ideally not happen if the SDK is used correctly.
            # TODO: use specific BackendError from exceptions.py
            err_msg = "target_callable is not a bound method of TaskInstance or TaskInstance is malformed."
            logger.error(err_msg)
            future.set_exception(RuntimeError(err_msg))
            return

        instance_id: str = task_instance.id
        task_control: TaskControl = task_instance.task_control

        future.set_running() # Mark our LocalFuture as running

        try:
            concurrent_future: ConcurrentFuture = self._executor.submit(target_callable, *args, **kwargs)
            
            with self._lock:
                self._active_concurrent_futures[instance_id] = concurrent_future
                self._active_task_controls[instance_id] = task_control

            def _on_done_callback(cf: ConcurrentFuture):
                try:
                    if cf.cancelled(): # This refers to concurrent_future's cancellation
                        # If concurrent_future was cancelled before running, reflect this.
                        # If it ran, TaskControl.should_stop would have been the way.
                        future.cancel() # Propagate cancellation to our LocalFuture
                    elif cf.exception():
                        future.set_exception(cf.exception()) # type: ignore
                    else:
                        future.set_result(cf.result())
                except Exception as e:
                    logger.error(f"Error in LocalFuture callback for task {instance_id}: {e}", exc_info=True)
                    if not future.done:
                        future.set_exception(e)
                finally:
                    with self._lock:
                        self._active_concurrent_futures.pop(instance_id, None)
                        self._active_task_controls.pop(instance_id, None)
            
            concurrent_future.add_done_callback(_on_done_callback)

        except Exception as e:
            logger.error(f"Failed to submit task {instance_id} to LocalExecutionBackend: {e}", exc_info=True)
            future.set_exception(e)
            # Clean up if it was added to maps before an error during submission itself (though unlikely here)
            with self._lock:
                self._active_concurrent_futures.pop(instance_id, None)
                self._active_task_controls.pop(instance_id, None)

    def cancel_task_instance(self, instance_id: str, session_id: Optional[str] = None) -> bool:
        """Requests a task instance to stop by signaling its TaskControl object."""
        with self._lock:
            task_control = self._active_task_controls.get(instance_id)
            concurrent_ft = self._active_concurrent_futures.get(instance_id)

            if task_control:
                logger.info(f"Requesting stop for task instance {instance_id} via TaskControl.")
                task_control.request_stop()
                
                # Optional: Attempt to cancel the concurrent.futures.Future as well.
                # This is only effective if the task hasn't started running yet.
                # If already running, it does nothing to the thread, but sets the future's state.
                if concurrent_ft and not concurrent_ft.done():
                    if concurrent_ft.cancel():
                        logger.info(f"concurrent.futures.Future for task {instance_id} was successfully cancelled (task likely not started).")
                    else:
                        logger.info(f"concurrent.futures.Future for task {instance_id} could not be cancelled (task likely already running or finished).")
                return True
            else:
                logger.warning(f"Task instance {instance_id} not found or no TaskControl available for cancellation.")
                return False

    def startup(self) -> None:
        logger.info("LocalExecutionBackend started.")

    def shutdown(self, wait: bool = True) -> None:
        logger.info(f"LocalExecutionBackend shutting down (wait={wait})...")
        self._executor.shutdown(wait=wait)
        logger.info("LocalExecutionBackend shutdown complete.")

# Example of how to register this backend (typically done in backends/__init__.py or by user)
# from . import register_backend
# register_backend("local", LocalExecutionBackend()) 