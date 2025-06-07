import os
from typing import Any, Callable, Optional
import logging
from abc import ABC, abstractmethod

from ..future import Future
from ..exceptions import BackendError

# Logger for the remote backend
logger = logging.getLogger(__name__)

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

class RemoteExecutionBackend(ExecutionBackend):
    """
    An execution backend that submits tasks to a remote Numerous Task service.
    """

    def __init__(self, api_url: Optional[str] = None, api_token: Optional[str] = None):
        """
        Initializes the RemoteExecutionBackend.

        Args:
            api_url: The base URL for the Numerous Task API.
            api_token: The authentication token for the API.
        """
        self.api_url = api_url or os.environ.get("NUMEROUS_API_URL")
        self.api_token = api_token # Or from os.environ.get("NUMEROUS_API_TOKEN")

        if not self.api_url:
            logger.warning("RemoteExecutionBackend initialized without an API URL.")

        logger.info(f"RemoteExecutionBackend initialized. API URL: {self.api_url}")

    def execute(
        self, 
        target_callable: Callable[..., Any],
        future: Future,
        args: tuple, 
        kwargs: dict
    ) -> None:
        """
        Submits the task for execution on the remote backend.

        The 'target_callable' in a remote context is usually not the function itself,
        but rather metadata about the task (like its registered name or ID) that
        the remote service can use to look up and run the actual code.
        """
        logger.info(f"RemoteExecutionBackend.execute called for task (metadata): {target_callable}")
        logger.info(f"Args: {args}, Kwargs: {kwargs}")
        
        if not self.api_url:
            future.set_exception(BackendError("Remote backend not configured with API URL. Cannot execute task."))
            return

        logger.warning("RemoteExecutionBackend.execute is not fully implemented.")
        future.set_exception(NotImplementedError("Remote task execution is not yet implemented."))

    def cancel_task_instance(self, instance_id: str, session_id: Optional[str] = None) -> bool:
        """
        Requests cancellation of a task instance on the remote backend.
        'instance_id' here is the SDK's internal TaskInstance.id. This might need to
        be mapped to a remote service's task ID.
        """
        logger.info(f"RemoteExecutionBackend.cancel_task_instance called for instance: {instance_id}")

        if not self.api_url:
            logger.error("Remote backend not configured with API URL. Cannot cancel task.")
            return False
            
        logger.warning("RemoteExecutionBackend.cancel_task_instance is not fully implemented.")
        raise NotImplementedError("Remote task cancellation is not yet implemented.")

    def get_status(self, future: Future):
        """
        (Optional/Helper) Polls the remote service for the status of a task associated with the future.
        This logic might also live within a custom RemoteFuture type.
        """
        pass

    def startup(self) -> None:
        """Called when the backend is initialized, for any setup (e.g., API client)."""
        logger.info("RemoteExecutionBackend starting up...")
        logger.info("RemoteExecutionBackend startup complete.")

    def shutdown(self) -> None:
        """Called when the SDK or application is shutting down, for cleanup."""
        logger.info("RemoteExecutionBackend shutting down...")
        logger.info("RemoteExecutionBackend shutdown complete.")

# Need to import os for os.environ.get
import os 