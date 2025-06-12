"""
MockExecutionBackend for Task 1.0: Mock Backend Unit Testing.

Provides a comprehensive mock implementation of ExecutionBackend that:
- Simulates task execution without actual threading/processing
- Tracks all execution calls and state changes
- Allows predefined results and exceptions for testing
- Provides full control over execution timing and behavior
"""

import time
import threading
from typing import Callable, Any, Optional, Dict, List, Union
from dataclasses import dataclass, field
from enum import Enum
import logging
from uuid import uuid4

# Import the base backend interface
from numerous.tasks.backends import ExecutionBackend
from numerous.tasks.future import LocalFuture, TaskStatus
from numerous.tasks.control import TaskControl
from numerous.tasks.exceptions import BackendError

logger = logging.getLogger(__name__)


class MockExecutionMode(Enum):
    """Mock execution modes for different testing scenarios."""
    IMMEDIATE = "immediate"  # Complete immediately with result
    DELAYED = "delayed"      # Complete after specified delay
    MANUAL = "manual"        # Require manual completion via complete_task()
    FAILURE = "failure"      # Fail immediately with exception


@dataclass
class MockTaskExecution:
    """Represents a mock task execution with full state tracking."""
    instance_id: str
    target_callable: Callable[..., Any]
    future: LocalFuture
    args: tuple
    kwargs: dict
    started_at: float = field(default_factory=time.time)
    completed_at: Optional[float] = None
    result: Any = None
    exception: Optional[Exception] = None
    status: TaskStatus = TaskStatus.PENDING
    execution_mode: MockExecutionMode = MockExecutionMode.IMMEDIATE
    delay_seconds: float = 0.0
    call_count: int = 0


class MockExecutionBackend(ExecutionBackend):
    """
    Mock execution backend for comprehensive unit testing.
    
    Features:
    - Simulates task execution without actual threading
    - Tracks all execution calls and state changes
    - Configurable execution modes (immediate, delayed, manual, failure)
    - Predefined results and exceptions for testing scenarios
    - Full state inspection for test assertions
    """
    
    def __init__(self):
        """Initialize the mock backend with clean state."""
        self._executions: Dict[str, MockTaskExecution] = {}
        self._predefined_results: Dict[str, Any] = {}
        self._predefined_exceptions: Dict[str, Exception] = {}
        self._default_mode = MockExecutionMode.IMMEDIATE
        self._default_delay = 0.0
        self._lock = threading.Lock()
        self._is_started = False
        self._execution_history: List[MockTaskExecution] = []
        
        logger.info("MockExecutionBackend initialized")
    
    def startup(self) -> None:
        """Called when the backend is initialized."""
        with self._lock:
            self._is_started = True
            logger.info("MockExecutionBackend started up")
    
    def shutdown(self) -> None:
        """Called when the SDK is shutting down."""
        with self._lock:
            self._is_started = False
            # Complete any pending executions with cancellation
            for execution in self._executions.values():
                if execution.status in [TaskStatus.PENDING, TaskStatus.RUNNING]:
                    execution.status = TaskStatus.CANCELLED
                    execution.completed_at = time.time()
                    execution.future.set_exception(BackendError("Backend shutdown"))
            logger.info("MockExecutionBackend shut down")
    
    def execute(
        self, 
        target_callable: Callable[..., Any],
        future: LocalFuture,
        args: tuple, 
        kwargs: dict
    ) -> None:
        """
        Execute the target_callable according to configured mock behavior.
        
        This is the core method that simulates task execution based on:
        - Predefined results/exceptions for the task
        - Configured execution mode (immediate, delayed, manual, failure)
        - Default backend behavior
        """
        if not isinstance(future, LocalFuture):
            raise TypeError("MockExecutionBackend requires a LocalFuture object")
        
        # Extract task instance information
        task_instance = getattr(target_callable, '__self__', None)
        if task_instance is None:
            raise BackendError("target_callable is not a bound method of TaskInstance")
        
        instance_id = getattr(task_instance, 'id', str(uuid4()))
        task_name = getattr(task_instance, 'task_definition_name', 'unknown_task')
        
        with self._lock:
            if not self._is_started:
                future.set_exception(BackendError("Backend not started"))
                return
            
            # Create execution record
            execution = MockTaskExecution(
                instance_id=instance_id,
                target_callable=target_callable,
                future=future,
                args=args,
                kwargs=kwargs,
                execution_mode=self._get_execution_mode(task_name),
                delay_seconds=self._get_execution_delay(task_name)
            )
            
            self._executions[instance_id] = execution
            self._execution_history.append(execution)
            execution.call_count += 1
            
            logger.info(f"MockBackend executing task {task_name}/{instance_id} in mode {execution.execution_mode.value}")
        
        # Set future to running
        future.set_running()
        execution.status = TaskStatus.RUNNING
        
        # Execute based on mode
        if execution.execution_mode == MockExecutionMode.IMMEDIATE:
            self._complete_immediately(execution)
        elif execution.execution_mode == MockExecutionMode.DELAYED:
            self._complete_delayed(execution)
        elif execution.execution_mode == MockExecutionMode.FAILURE:
            self._complete_with_failure(execution)
        # MANUAL mode requires explicit completion via complete_task()
    
    def cancel_task_instance(self, instance_id: str, session_id: Optional[str] = None) -> bool:
        """Attempt to cancel a task instance."""
        with self._lock:
            execution = self._executions.get(instance_id)
            if execution is None:
                return False
            
            if execution.status in [TaskStatus.PENDING, TaskStatus.RUNNING]:
                execution.status = TaskStatus.CANCELLED
                execution.completed_at = time.time()
                execution.future.set_exception(BackendError(f"Task {instance_id} cancelled"))
                logger.info(f"MockBackend cancelled task {instance_id}")
                return True
            
            return False
    
    # Configuration methods for testing
    
    def set_default_execution_mode(self, mode: MockExecutionMode, delay: float = 0.0) -> None:
        """Set the default execution mode for all tasks."""
        with self._lock:
            self._default_mode = mode
            self._default_delay = delay
            logger.info(f"MockBackend default mode set to {mode.value} with {delay}s delay")
    
    def set_task_result(self, task_name: str, result: Any) -> None:
        """Set a predefined result for a specific task."""
        with self._lock:
            self._predefined_results[task_name] = result
            logger.info(f"MockBackend set predefined result for {task_name}")
    
    def set_task_exception(self, task_name: str, exception: Exception) -> None:
        """Set a predefined exception for a specific task."""
        with self._lock:
            self._predefined_exceptions[task_name] = exception
            logger.info(f"MockBackend set predefined exception for {task_name}")
    
    def complete_task(self, instance_id: str, result: Any = None, exception: Exception = None) -> bool:
        """Manually complete a task (for MANUAL mode)."""
        with self._lock:
            execution = self._executions.get(instance_id)
            if execution is None or execution.status != TaskStatus.RUNNING:
                return False
            
            execution.completed_at = time.time()
            
            if exception is not None:
                execution.exception = exception
                execution.status = TaskStatus.FAILED
                execution.future.set_exception(exception)
                logger.info(f"MockBackend manually completed task {instance_id} with exception")
            else:
                execution.result = result
                execution.status = TaskStatus.COMPLETED
                execution.future.set_result(result)
                logger.info(f"MockBackend manually completed task {instance_id} with result")
            
            return True
    
    # State inspection methods for testing
    
    def get_execution(self, instance_id: str) -> Optional[MockTaskExecution]:
        """Get execution details for a task instance."""
        with self._lock:
            return self._executions.get(instance_id)
    
    def get_all_executions(self) -> List[MockTaskExecution]:
        """Get all execution records."""
        with self._lock:
            return list(self._execution_history)
    
    def get_pending_executions(self) -> List[MockTaskExecution]:
        """Get all pending/running executions."""
        with self._lock:
            return [e for e in self._executions.values() 
                   if e.status in [TaskStatus.PENDING, TaskStatus.RUNNING]]
    
    def get_execution_count(self) -> int:
        """Get total number of executions attempted."""
        with self._lock:
            return len(self._execution_history)
    
    def clear_state(self) -> None:
        """Clear all execution state (for test cleanup)."""
        with self._lock:
            self._executions.clear()
            self._execution_history.clear()
            self._predefined_results.clear()
            self._predefined_exceptions.clear()
            logger.info("MockBackend state cleared")
    
    # Private helper methods
    
    def _get_execution_mode(self, task_name: str) -> MockExecutionMode:
        """Determine execution mode for a task."""
        # Check if we have predefined behavior
        if task_name in self._predefined_exceptions:
            return MockExecutionMode.FAILURE
        return self._default_mode
    
    def _get_execution_delay(self, task_name: str) -> float:
        """Determine execution delay for a task."""
        return self._default_delay
    
    def _complete_immediately(self, execution: MockTaskExecution) -> None:
        """Complete execution immediately."""
        task_name = getattr(execution.target_callable.__self__, 'task_definition_name', 'unknown')
        
        try:
            if task_name in self._predefined_results:
                result = self._predefined_results[task_name]
                execution.result = result
                execution.status = TaskStatus.COMPLETED
                execution.completed_at = time.time()
                execution.future.set_result(result)
            else:
                # Actually call the target callable for real execution
                result = execution.target_callable(*execution.args, **execution.kwargs)
                execution.result = result
                execution.status = TaskStatus.COMPLETED
                execution.completed_at = time.time()
                execution.future.set_result(result)
        except Exception as e:
            execution.exception = e
            execution.status = TaskStatus.FAILED
            execution.completed_at = time.time()
            execution.future.set_exception(e)
    
    def _complete_delayed(self, execution: MockTaskExecution) -> None:
        """Complete execution after delay (simulated)."""
        def delayed_completion():
            time.sleep(execution.delay_seconds)
            self._complete_immediately(execution)
        
        # Start delayed completion in background thread
        thread = threading.Thread(target=delayed_completion, daemon=True)
        thread.start()
    
    def _complete_with_failure(self, execution: MockTaskExecution) -> None:
        """Complete execution with predefined failure."""
        task_name = getattr(execution.target_callable.__self__, 'task_definition_name', 'unknown')
        
        exception = self._predefined_exceptions.get(task_name, BackendError(f"Mock failure for {task_name}"))
        execution.exception = exception
        execution.status = TaskStatus.FAILED
        execution.completed_at = time.time()
        execution.future.set_exception(exception) 