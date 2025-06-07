import pytest
import threading
import time
from unittest.mock import Mock, patch, MagicMock
from typing import Optional

from numerous.tasks import (
    task, Task, TaskInstance, TaskConfig, TaskControl, Session, Future, 
    TaskStatus, LocalFuture, TaskError, MaxInstancesReachedError,
    SessionNotFoundError, SessionError, TaskCancelledError, BackendError,
    TaskDefinitionError, ExecutionBackend, LocalExecutionBackend,
    set_task_control_handler, LocalTaskControlHandler, PoCMockRemoteTaskControlHandler
)


class TestTaskDecorator:
    """Test the @task decorator and Task class."""
    
    def test_task_decorator_without_params(self):
        """Test @task decorator without parameters."""
        @task
        def simple_task():
            return "hello"
        
        assert isinstance(simple_task, Task)
        assert simple_task.name == "simple_task"
        assert simple_task.config.max_parallel == 1
        assert simple_task.config.size == "small"
        assert not simple_task.expects_task_control
    
    def test_task_decorator_with_params(self):
        """Test @task decorator with parameters."""
        @task(name="custom_name", max_parallel=3, size="large")
        def configured_task():
            return "configured"
        
        assert isinstance(configured_task, Task)
        assert configured_task.name == "custom_name"
        assert configured_task.config.max_parallel == 3
        assert configured_task.config.size == "large"
    
    def test_task_with_task_control_parameter(self):
        """Test task function that expects TaskControl."""
        @task
        def task_with_control(tc: TaskControl, value: int):
            tc.log("Processing value", level="info")
            return value * 2
        
        assert task_with_control.expects_task_control
    
    def test_task_without_task_control_parameter(self):
        """Test task function that doesn't expect TaskControl."""
        @task
        def task_without_control(value: int):
            return value * 2
        
        assert not task_without_control.expects_task_control
    
    def test_task_config_defaults(self):
        """Test TaskConfig default values."""
        config = TaskConfig()
        assert config.max_parallel == 1
        assert config.size == "small"
        assert config.name is None


class TestTaskExecution:
    """Test task execution functionality."""
    
    def test_simple_task_execution(self):
        """Test executing a simple task."""
        @task
        def add_numbers(a: int, b: int):
            return a + b
        
        with Session() as session:
            result = add_numbers(5, 3)
            assert result == 8
    
    def test_task_with_control_execution(self):
        """Test executing a task that uses TaskControl."""
        @task
        def task_with_logging(tc: TaskControl, message: str):
            tc.log(f"Processing: {message}")
            tc.update_progress(50.0, "halfway")
            tc.update_status("processing")
            return f"Processed: {message}"
        
        with Session() as session:
            result = task_with_logging("test message")
            assert result == "Processed: test message"
    
    def test_task_instance_creation(self):
        """Test creating task instances."""
        @task
        def test_task():
            return "instance"
        
        with Session() as session:
            instance = test_task.instance()
            assert isinstance(instance, TaskInstance)
            assert instance.task_definition == test_task
            assert instance.session == session
            assert instance.id is not None
    
    def test_task_instance_start_and_future(self):
        """Test starting task instance and getting future."""
        @task
        def async_task(value: int):
            return value * 2
        
        with Session() as session:
            instance = async_task.instance()
            future = instance.start(5)
            assert isinstance(future, Future)
            result = future.result()
            assert result == 10
    
    def test_max_parallel_enforcement(self):
        """Test that max_parallel limits are enforced."""
        @task(max_parallel=1)
        def slow_task():
            time.sleep(0.1)
            return "done"
        
        with Session() as session:
            # Start first instance
            instance1 = slow_task.instance()
            future1 = instance1.start()
            
            # Try to start second instance - should fail
            instance2 = slow_task.instance()
            with pytest.raises(MaxInstancesReachedError):
                instance2.start()
            
            # Wait for first to complete
            future1.result()
            
            # Now second instance should be able to start
            future2 = instance2.start()
            assert future2.result() == "done"
    
    def test_task_direct_call_with_max_parallel_gt_1(self):
        """Test that direct call fails for tasks with max_parallel > 1."""
        @task(max_parallel=2)
        def multi_task():
            return "multi" 
        
        with Session() as session:
            with pytest.raises(TypeError, match="max_parallel > 1"):
                multi_task()


class TestSession:
    """Test Session functionality."""
    
    def test_session_creation(self):
        """Test creating a session."""
        session = Session(name="test_session")
        assert session.name == "test_session"
        assert session.id is not None
        assert not session._is_active
    
    def test_session_context_manager(self):
        """Test session as context manager."""
        session = Session()
        assert Session.current() is None
        
        with session:
            assert Session.current() == session
            assert session._is_active
        
        assert Session.current() is None
        assert not session._is_active
    
    def test_nested_sessions(self):
        """Test nested session contexts."""
        session1 = Session(name="outer")
        session2 = Session(name="inner")
        
        with session1:
            assert Session.current() == session1
            
            with session2:
                assert Session.current() == session2
            
            assert Session.current() == session1
    
    def test_session_re_entry_error(self):
        """Test that re-entering an active session raises error."""
        session = Session()
        
        with session:
            with pytest.raises(SessionError, match="already active"):
                with session:
                    pass
    
    def test_session_task_tracking(self):
        """Test that sessions track task instances."""
        @task
        def tracked_task():
            return "tracked"
        
        with Session() as session:
            instance = tracked_task.instance()
            assert session.get_running_instances_count("tracked_task") == 0
            
            future = instance.start()
            # Note: for very fast tasks, this might be 0 if already completed
            # So we just check the task was registered
            assert "tracked_task" in session.tasks
            
            future.result()  # Wait for completion


class TestFutureAndTaskStatus:
    """Test Future and TaskStatus functionality."""
    
    def test_local_future_success(self):
        """Test LocalFuture with successful result."""
        future = LocalFuture()
        assert future.status == TaskStatus.PENDING
        assert not future.done
        assert future.error is None
        
        # Simulate task execution
        future.set_running()
        assert future.status == TaskStatus.RUNNING
        
        future.set_result("success")
        assert future.status == TaskStatus.COMPLETED
        assert future.done
        assert future.result() == "success"
    
    def test_local_future_exception(self):
        """Test LocalFuture with exception."""
        future = LocalFuture()
        future.set_running()
        
        test_error = ValueError("test error")
        future.set_exception(test_error)
        
        assert future.status == TaskStatus.FAILED
        assert future.done
        assert future.error == test_error
        
        with pytest.raises(ValueError, match="test error"):
            future.result()
    
    def test_local_future_cancellation(self):
        """Test LocalFuture cancellation."""
        future = LocalFuture()
        
        cancelled = future.cancel()
        assert cancelled
        assert future.status == TaskStatus.CANCELLED
        assert future.done
        
        with pytest.raises(TaskCancelledError):
            future.result()
    
    def test_future_timeout(self):
        """Test Future timeout functionality."""
        future = LocalFuture()
        
        with pytest.raises(TimeoutError):
            future.result(timeout=0.1)


class TestTaskControl:
    """Test TaskControl functionality."""
    
    def test_task_control_creation(self):
        """Test TaskControl object creation."""
        tc = TaskControl(instance_id="test_id", task_definition_name="test_task")
        assert tc.instance_id == "test_id"
        assert tc.task_definition_name == "test_task"
        assert not tc.should_stop
        assert tc.progress == 0.0
        assert tc.status == ""
    
    def test_task_control_progress_update(self):
        """Test updating task progress."""
        tc = TaskControl(instance_id="test_id", task_definition_name="test_task")
        
        tc.update_progress(25.0, "quarter done")
        assert tc.progress == 25.0
        assert tc.status == "quarter done"
        
        tc.update_progress(100.0)
        assert tc.progress == 100.0
    
    def test_task_control_status_update(self):
        """Test updating task status."""
        tc = TaskControl(instance_id="test_id", task_definition_name="test_task")
        
        tc.update_status("working")
        assert tc.status == "working"
    
    def test_task_control_stop_request(self):
        """Test requesting task to stop."""
        tc = TaskControl(instance_id="test_id", task_definition_name="test_task")
        
        assert not tc.should_stop
        tc.request_stop()
        assert tc.should_stop
    
    def test_task_control_logging(self):
        """Test task control logging."""
        tc = TaskControl(instance_id="test_id", task_definition_name="test_task")
        
        # This should not raise an exception
        tc.log("test message", level="info")
        tc.log("debug message", level="debug", extra_field="extra_value")


class TestTaskControlHandlers:
    """Test TaskControl handlers."""
    
    def test_local_task_control_handler(self):
        """Test LocalTaskControlHandler functionality."""
        handler = LocalTaskControlHandler()
        tc = TaskControl(instance_id="test_id", task_definition_name="test_task")
        
        # These should not raise exceptions
        handler.log(tc, "test message", "info")
        handler.update_progress(tc, 50.0, "halfway")
        handler.update_status(tc, "running")
        handler.request_stop(tc)
    
    def test_poc_mock_remote_handler(self):
        """Test PoCMockRemoteTaskControlHandler functionality."""
        handler = PoCMockRemoteTaskControlHandler()
        tc = TaskControl(instance_id="test_id", task_definition_name="test_task")
        
        # These should not raise exceptions
        handler.log(tc, "test message", "info")
        handler.update_progress(tc, 75.0, "almost done")
        handler.update_status(tc, "finishing")
        handler.request_stop(tc)
    
    def test_set_task_control_handler(self):
        """Test setting custom task control handler."""
        original_handler = TaskControl(instance_id="test", task_definition_name="test")._handler
        
        mock_handler = Mock(spec=LocalTaskControlHandler)
        set_task_control_handler(mock_handler)
        
        tc = TaskControl(instance_id="test_id", task_definition_name="test_task")
        assert tc._handler == mock_handler
        
        # Reset to default
        set_task_control_handler(None)
        tc2 = TaskControl(instance_id="test_id2", task_definition_name="test_task2")
        assert isinstance(tc2._handler, LocalTaskControlHandler)


class TestExceptions:
    """Test custom exceptions."""
    
    def test_session_not_found_error(self):
        """Test SessionNotFoundError when no session is active."""
        @task
        def no_session_task():
            return "fail"
        
        # No active session
        with pytest.raises(SessionNotFoundError):
            no_session_task.instance()
    
    def test_task_error_inheritance(self):
        """Test that all custom exceptions inherit from TaskError."""
        exceptions_to_test = [
            MaxInstancesReachedError,
            SessionNotFoundError,
            SessionError,
            TaskCancelledError,
            BackendError,
            TaskDefinitionError
        ]
        
        for exc_class in exceptions_to_test:
            assert issubclass(exc_class, TaskError)
            assert issubclass(exc_class, Exception)
    
    def test_task_instance_double_start(self):
        """Test error when starting task instance twice."""
        @task
        def double_start_task():
            return "once"
        
        with Session() as session:
            instance = double_start_task.instance()
            instance.start()
            
            with pytest.raises(TaskError, match="already been started"):
                instance.start()


class TestLocalExecutionBackend:
    """Test LocalExecutionBackend functionality."""
    
    def test_backend_initialization(self):
        """Test LocalExecutionBackend initialization."""
        backend = LocalExecutionBackend(max_workers=2)
        assert backend._executor is not None
        backend.shutdown()
    
    def test_backend_task_execution(self):
        """Test task execution through backend."""
        @task
        def backend_task(value: int):
            return value + 10
        
        with Session() as session:
            instance = backend_task.instance()
            future = instance.start(5)
            result = future.result()
            assert result == 15
    
    def test_backend_task_cancellation(self):
        """Test task cancellation through backend."""
        @task
        def cancellable_task(tc: TaskControl):
            for i in range(100):
                if tc.should_stop:
                    return "stopped"
                time.sleep(0.01)
            return "completed"
        
        with Session() as session:
            instance = cancellable_task.instance()
            future = instance.start()
            
            # Give task a moment to start
            time.sleep(0.02)
            
            # Request cancellation
            success = instance.backend.cancel_task_instance(instance.id)
            assert success
            
            # Check that task control was signaled
            assert instance.task_control.should_stop


class TestThreadSafety:
    """Test thread safety of various components."""
    
    def test_concurrent_task_execution(self):
        """Test multiple tasks running concurrently."""
        results = []
        
        @task
        def concurrent_task(task_id: int):
            time.sleep(0.1)  # Simulate work
            return f"result_{task_id}"
        
        def run_task(task_id):
            with Session() as session:
                result = concurrent_task(task_id)
                results.append(result)
        
        threads = []
        for i in range(5):
            thread = threading.Thread(target=run_task, args=(i,))
            threads.append(thread)
            thread.start()
        
        for thread in threads:
            thread.join()
        
        assert len(results) == 5
        assert all(f"result_{i}" in results for i in range(5))
    
    def test_concurrent_session_usage(self):
        """Test concurrent session usage."""
        session_results = []
        
        def use_session():
            with Session() as session:
                session_results.append(session.id)
        
        threads = []
        for i in range(3):
            thread = threading.Thread(target=use_session)
            threads.append(thread)
            thread.start()
        
        for thread in threads:
            thread.join()
        
        # Should have 3 different session IDs
        assert len(set(session_results)) == 3


class TestComplexScenarios:
    """Test complex usage scenarios."""
    
    def test_workflow_with_multiple_tasks(self):
        """Test a workflow with multiple interdependent tasks."""
        @task
        def prepare_data(data: str):
            return f"prepared_{data}"
        
        @task
        def process_data(tc: TaskControl, data: str):
            tc.update_progress(50.0, "processing")
            result = f"processed_{data}"
            tc.update_progress(100.0, "complete")
            return result
        
        @task
        def finalize_data(data: str):
            return f"finalized_{data}"
        
        with Session() as session:
            # Sequential workflow
            step1 = prepare_data("input")
            step2 = process_data(step1)
            step3 = finalize_data(step2)
            
            assert step3 == "finalized_processed_prepared_input"
    
    def test_error_handling_in_tasks(self):
        """Test error handling within tasks."""
        @task
        def failing_task(should_fail: bool):
            if should_fail:
                raise ValueError("Task failed as requested")
            return "success"
        
        with Session() as session:
            # Successful execution
            result = failing_task(False)
            assert result == "success"
            
            # Failed execution
            instance = failing_task.instance()
            future = instance.start(True)
            
            with pytest.raises(ValueError, match="Task failed as requested"):
                future.result()
    
    def test_long_running_task_with_control(self):
        """Test long-running task with progress updates and cancellation."""
        progress_updates = []
        
        @task
        def long_task(tc: TaskControl, iterations: int):
            for i in range(iterations):
                if tc.should_stop:
                    tc.log("Task stopped early", level="info")
                    return f"stopped_at_{i}"
                
                progress = (i + 1) / iterations * 100
                tc.update_progress(progress, f"iteration_{i}")
                progress_updates.append(progress)
                time.sleep(0.01)
            
            return f"completed_{iterations}"
        
        with Session() as session:
            instance = long_task.instance()
            future = instance.start(10)
            
            # Let it run a bit
            time.sleep(0.05)
            
            # Request stop
            instance.stop()
            
            result = future.result()
            assert result.startswith("stopped_at_")
            assert len(progress_updates) > 0 