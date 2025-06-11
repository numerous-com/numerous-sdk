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


class TestTaskControlAdvanced:
    """Test advanced TaskControl functionality and integration."""
    
    def test_task_control_with_mock_handler(self):
        """Test TaskControl with mocked handler to verify method calls."""
        mock_handler = Mock()
        set_task_control_handler(mock_handler)
        
        try:
            tc = TaskControl(instance_id="test_id", task_definition_name="test_task")
            
            # Test logging
            tc.log("test message", level="info", extra="data")
            mock_handler.log.assert_called_with(tc, "test message", "info", extra="data")
            
            # Test progress update
            tc.update_progress(75.0, "progress status")
            mock_handler.update_progress.assert_called_with(tc, 75.0, "progress status")
            
            # Test status update
            tc.update_status("new status")
            mock_handler.update_status.assert_called_with(tc, "new status")
            
            # Test stop request
            tc.request_stop()
            # Note: request_stop might not call handler, just check the internal state
            assert tc.should_stop
            
        finally:
            # Reset to default handler
            set_task_control_handler(None)
    
    def test_task_control_progress_validation(self):
        """Test TaskControl progress value validation and edge cases."""
        tc = TaskControl(instance_id="test_id", task_definition_name="test_task")
        
        # Test valid progress values
        tc.update_progress(0.0)
        assert tc.progress == 0.0
        
        tc.update_progress(50.5)
        assert tc.progress == 50.5
        
        tc.update_progress(100.0)
        assert tc.progress == 100.0
        
        # Test progress without status
        tc.update_progress(25.0)
        assert tc.progress == 25.0
        # Status should remain unchanged from previous call
        assert tc.status == ""
    
    def test_task_control_in_actual_task_execution(self):
        """Test TaskControl integration with actual task execution."""
        control_calls = []
        
        @task
        def task_with_comprehensive_control(tc: TaskControl, steps: int):
            control_calls.append(("start", tc.instance_id))
            
            for i in range(steps):
                if tc.should_stop:
                    control_calls.append(("stopped", i))
                    return f"stopped_at_step_{i}"
                
                progress = (i + 1) / steps * 100
                tc.update_progress(progress, f"Step {i+1} of {steps}")
                tc.log(f"Completed step {i+1}", level="debug")
                control_calls.append(("step", i+1, progress))
                
                if i == steps // 2:
                    tc.update_status("halfway_complete")
                    control_calls.append(("halfway", tc.status))
            
            tc.update_status("completed")
            control_calls.append(("finished", tc.status))
            return f"completed_{steps}_steps"
        
        with Session() as session:
            result = task_with_comprehensive_control(5)
            assert result == "completed_5_steps"
            
            # Verify all control calls were made
            start_calls = [call for call in control_calls if call[0] == "start"]
            assert len(start_calls) == 1
            assert ("halfway", "halfway_complete") in control_calls
            assert ("finished", "completed") in control_calls
            assert len([call for call in control_calls if call[0] == "step"]) == 5
    
    def test_task_control_stop_behavior_in_running_task(self):
        """Test TaskControl stop behavior in a running task."""
        stop_check_count = 0
        
        @task  
        def stoppable_task(tc: TaskControl, max_iterations: int):
            nonlocal stop_check_count
            for i in range(max_iterations):
                stop_check_count += 1
                if tc.should_stop:
                    tc.log(f"Task stopped at iteration {i}", level="info")
                    return f"stopped_at_{i}"
                
                tc.update_progress(i / max_iterations * 100)
                time.sleep(0.01)  # Small delay to allow stop signal
            
            return f"completed_{max_iterations}"
        
        with Session() as session:
            instance = stoppable_task.instance()
            future = instance.start(100)
            
            # Let it run for a bit
            time.sleep(0.05)
            
            # Request stop
            instance.stop()
            
            result = future.result()
            assert result.startswith("stopped_at_")
            assert stop_check_count > 0  # Verify stop was checked
    
    def test_task_control_multiple_status_updates(self):
        """Test multiple status updates in sequence."""
        @task
        def multi_status_task(tc: TaskControl):
            statuses = ["initializing", "processing", "validating", "finalizing", "complete"]
            
            for i, status in enumerate(statuses):
                tc.update_status(status)
                tc.update_progress((i + 1) / len(statuses) * 100)
                time.sleep(0.01)
            
            return tc.status
        
        with Session() as session:
            result = multi_status_task()
            assert result == "complete"
    
    def test_task_control_logging_with_different_levels(self):
        """Test TaskControl logging with different log levels and extra data."""
        logged_messages = []
        
        # Mock handler to capture log calls
        class TestHandler(LocalTaskControlHandler):
            def log(self, task_control, message, level, **kwargs):
                logged_messages.append((message, level, kwargs))
        
        test_handler = TestHandler()
        set_task_control_handler(test_handler)
        
        try:
            @task
            def logging_task(tc: TaskControl):
                tc.log("Debug message", level="debug", component="data_loader")
                tc.log("Info message", level="info", user_id=123)
                tc.log("Warning message", level="warning", retry_count=3)
                tc.log("Error message", level="error", error_code="E001")
                return "logged"
            
            with Session() as session:
                result = logging_task()
                assert result == "logged"
                
                # Verify all log messages were captured
                assert len(logged_messages) == 4
                
                debug_msg = next(msg for msg in logged_messages if msg[1] == "debug")
                assert debug_msg[0] == "Debug message"
                assert debug_msg[2]["component"] == "data_loader"
                
                info_msg = next(msg for msg in logged_messages if msg[1] == "info")
                assert info_msg[0] == "Info message"
                assert info_msg[2]["user_id"] == 123
                
                warning_msg = next(msg for msg in logged_messages if msg[1] == "warning")
                assert warning_msg[0] == "Warning message"
                assert warning_msg[2]["retry_count"] == 3
                
                error_msg = next(msg for msg in logged_messages if msg[1] == "error")
                assert error_msg[0] == "Error message"
                assert error_msg[2]["error_code"] == "E001"
        
        finally:
            set_task_control_handler(None)
    
    def test_task_control_state_persistence(self):
        """Test that TaskControl state persists throughout task execution."""
        state_snapshots = []
        
        @task
        def state_tracking_task(tc: TaskControl):
            # Initial state
            state_snapshots.append({
                'progress': tc.progress,
                'status': tc.status,
                'should_stop': tc.should_stop
            })
            
            # Update progress
            tc.update_progress(30.0, "processing")
            state_snapshots.append({
                'progress': tc.progress,
                'status': tc.status,
                'should_stop': tc.should_stop
            })
            
            # Update status only
            tc.update_status("validating")
            state_snapshots.append({
                'progress': tc.progress,
                'status': tc.status,
                'should_stop': tc.should_stop
            })
            
            # Update progress without status
            tc.update_progress(80.0)
            state_snapshots.append({
                'progress': tc.progress,
                'status': tc.status,
                'should_stop': tc.should_stop
            })
            
            return "state_tracked"
        
        with Session() as session:
            result = state_tracking_task()
            assert result == "state_tracked"
            
            # Verify state progression
            assert len(state_snapshots) == 4
            
            # Initial state
            assert state_snapshots[0]['progress'] == 0.0
            assert state_snapshots[0]['status'] == ""
            assert not state_snapshots[0]['should_stop']
            
            # After first update
            assert state_snapshots[1]['progress'] == 30.0
            assert state_snapshots[1]['status'] == "processing"
            
            # After status-only update
            assert state_snapshots[2]['progress'] == 30.0  # Should remain unchanged
            assert state_snapshots[2]['status'] == "validating"
            
            # After progress-only update
            assert state_snapshots[3]['progress'] == 80.0
            assert state_snapshots[3]['status'] == "validating"  # Should remain unchanged
    
    def test_task_control_error_handling_in_updates(self):
        """Test TaskControl behavior when handler methods raise exceptions."""
        class FailingHandler(LocalTaskControlHandler):
            def update_progress(self, task_control, progress, status=None):
                raise RuntimeError("Handler failed")
            
            def log(self, task_control, message, level, **kwargs):
                if level == "error":
                    raise RuntimeError("Logging failed")
                super().log(task_control, message, level, **kwargs)
        
        failing_handler = FailingHandler()
        set_task_control_handler(failing_handler)
        
        try:
            @task
            def error_prone_task(tc: TaskControl):
                # This should handle the handler exception gracefully
                try:
                    tc.update_progress(50.0, "halfway")
                except RuntimeError:
                    pass  # Handler error, but task continues
                
                try:
                    tc.log("This should work", level="info")
                    tc.log("This should fail", level="error")
                except RuntimeError:
                    pass  # Handler error, but task continues
                
                return "completed_despite_errors"
            
            with Session() as session:
                result = error_prone_task()
                assert result == "completed_despite_errors"
        
        finally:
            set_task_control_handler(None)


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