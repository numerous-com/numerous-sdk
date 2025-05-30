from abc import ABC, abstractmethod
from typing import Any, Optional, Generic, TypeVar
import threading
import time
from .exceptions import TaskCancelledError # Defined in exceptions.py

T = TypeVar('T')

class TaskStatus:
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"

class Future(ABC, Generic[T]):
    """Base class for task futures.

    Represents the eventual result of an asynchronous task operation.
    Each backend should implement its own version.
    """

    @abstractmethod
    def result(self, timeout: Optional[float] = None) -> T:
        """Return the result of the task.

        If the task is not yet complete, this method should wait up to
        `timeout` seconds. If `timeout` is None, it waits indefinitely.
        Raises TimeoutError if the timeout is reached.
        Raises an exception if the task failed.
        """
        pass

    @property
    @abstractmethod
    def status(self) -> str:
        """Return the current status of the task (e.g., from TaskStatus)."""
        pass

    @property
    @abstractmethod
    def done(self) -> bool:
        """Return True if the task is completed, failed, or cancelled."""
        pass

    @property
    @abstractmethod
    def error(self) -> Optional[BaseException]:
        """Return the exception raised by the task if it failed, else None."""
        pass

    @abstractmethod
    def cancel(self) -> bool:
        """Attempt to cancel the task. 
        
        Returns True if cancellation was successful or already cancelled,
        False otherwise (e.g., if the task is already completed or cannot be cancelled).
        """
        pass

    # Optional: Add callback functionality
    # def add_done_callback(self, fn):
    #     pass


class LocalFuture(Future[T]):
    """A basic Future implementation for locally executed tasks (e.g., in threads)."""
    def __init__(self):
        self._result: Optional[T] = None
        self._exception: Optional[BaseException] = None
        self._status: str = TaskStatus.PENDING
        self._done_event = threading.Event()
        self._lock = threading.Lock()

    def result(self, timeout: Optional[float] = None) -> T:
        if not self._done_event.wait(timeout):
            raise TimeoutError("Timeout waiting for task result")
        
        with self._lock:
            if self._exception:
                raise self._exception
            if self._status == TaskStatus.CANCELLED:
                raise TaskCancelledError("Task was cancelled")
            return self._result # type: ignore[return-value] # Assuming result is set if no exception

    @property
    def status(self) -> str:
        with self._lock:
            return self._status

    @property
    def done(self) -> bool:
        return self._done_event.is_set()

    @property
    def error(self) -> Optional[BaseException]:
        with self._lock:
            return self._exception

    def cancel(self) -> bool:
        with self._lock:
            if self._status not in [TaskStatus.COMPLETED, TaskStatus.FAILED, TaskStatus.CANCELLED]:
                # This future itself doesn't trigger the stop in TaskControl.
                # The TaskInstance or backend managing the TaskControl should do that.
                # This merely marks the future as cancelled.
                self._status = TaskStatus.CANCELLED
                self._done_event.set() # Signal completion due to cancellation
                return True
            return self._status == TaskStatus.CANCELLED

    def set_running(self) -> None:
        with self._lock:
            if self._status == TaskStatus.PENDING:
                self._status = TaskStatus.RUNNING

    def set_result(self, result: T) -> None:
        with self._lock:
            if self._status == TaskStatus.RUNNING:
                self._result = result
                self._status = TaskStatus.COMPLETED
                self._done_event.set()
            # else: log warning or raise error if trying to set result on non-running/already done task

    def set_exception(self, exception: BaseException) -> None:
        with self._lock:
            if self._status in [TaskStatus.RUNNING, TaskStatus.PENDING]: # Can fail even before running
                self._exception = exception
                self._status = TaskStatus.FAILED
                self._done_event.set()
            # else: log warning or raise error 