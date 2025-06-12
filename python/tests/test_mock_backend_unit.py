"""
Task 1.0: Mock Backend Unit Testing - Comprehensive Test Suite.

Tests the mock backend implementation covering all execution paths:
- MockExecutionBackend functionality and state tracking
- MockTaskControlHandler logging and progress simulation  
- MockSessionManager in-memory state persistence
- Backend switching (Mock ↔ Local) capabilities
- Task cancellation and cleanup with mock backend
- Full unit test coverage using mock implementations
"""

import pytest
import time
import threading
from typing import Dict, Any
from unittest.mock import Mock, patch

from numerous.tasks import task, TaskControl, Session
from numerous.tasks.future import LocalFuture, TaskStatus
from numerous.tasks.exceptions import BackendError
from numerous.tasks.control import set_task_control_handler, get_task_control_handler

# Import our mock implementations
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__)))

from mocks.backend import MockExecutionBackend, MockExecutionMode, MockTaskExecution
from mocks.handler import MockTaskControlHandler, MockLogEntry, MockProgressUpdate
from mocks.session import MockSessionManager, MockSessionState
from mocks.fixtures import (
    mock_backend, mock_handler, mock_session_manager, mock_environment,
    mock_backend_immediate, mock_backend_delayed, mock_backend_manual, mock_backend_failure
)


# Test task definitions for mock testing
@task
def simple_mock_task(a: int, b: int) -> int:
    """Simple task for mock testing."""
    return a + b

@task  
def mock_task_with_control(tc: TaskControl, value: int, multiplier: int = 2) -> int:
    """Task with TaskControl for testing mock handler."""
    tc.log("Starting computation", level="info")
    tc.update_progress(25.0, "initialized")
    
    result = value * multiplier
    
    tc.update_progress(75.0, "processing")
    tc.log(f"Computed result: {result}", level="debug")
    
    tc.update_progress(100.0, "completed")
    tc.update_status("finished")
    
    return result

@task
def failing_mock_task(should_fail: bool = True) -> str:
    """Task that can fail for testing error handling."""
    if should_fail:
        raise ValueError("Mock task failure")
    return "success"

@task  
def long_running_mock_task(tc: TaskControl, steps: int, check_stop: bool = True) -> str:
    """Task that can be interrupted for cancellation testing."""
    tc.log(f"Starting long task with {steps} steps", level="info")
    
    for i in range(steps):
        if check_stop and tc.should_stop:
            tc.log(f"Task stopped at step {i}", level="warning")
            return f"stopped_at_step_{i}"
        
        progress = (i + 1) / steps * 100
        tc.update_progress(progress, f"Step {i+1}/{steps}")
        time.sleep(0.01)  # Small delay
    
    tc.log("Long task completed", level="info")
    return f"completed_{steps}_steps"


class TestMockExecutionBackend:
    """Test MockExecutionBackend functionality."""
    
    def test_mock_backend_initialization(self, mock_backend):
        """Test mock backend initializes correctly."""
        assert isinstance(mock_backend, MockExecutionBackend)
        assert mock_backend.get_execution_count() == 0
        assert len(mock_backend.get_all_executions()) == 0
        assert len(mock_backend.get_pending_executions()) == 0
    
    def test_mock_backend_startup_shutdown(self):
        """Test backend startup and shutdown lifecycle."""
        backend = MockExecutionBackend()
        
        # Should start not started
        assert not backend._is_started
        
        # Startup
        backend.startup()
        assert backend._is_started
        
        # Shutdown
        backend.shutdown()
        assert not backend._is_started
    
    def test_mock_backend_immediate_execution(self, mock_backend):
        """Test immediate execution mode."""
        mock_backend.set_default_execution_mode(MockExecutionMode.IMMEDIATE)
        
        # Create a mock task instance and future
        future = LocalFuture()
        
        # Mock task instance
        mock_task_instance = Mock()
        mock_task_instance.id = "test_task_123"
        mock_task_instance.task_definition_name = "simple_mock_task"
        
        def mock_target_callable(*args, **kwargs):
            return 42
        
        mock_target_callable.__self__ = mock_task_instance
        
        # Execute
        mock_backend.execute(mock_target_callable, future, (1, 2), {})
        
        # Should complete immediately
        assert future.done
        assert future.result() == 42
        assert mock_backend.get_execution_count() == 1
        
        execution = mock_backend.get_all_executions()[0]
        assert execution.status == TaskStatus.COMPLETED
        assert execution.result == 42
    
    def test_mock_backend_predefined_results(self, mock_backend):
        """Test setting predefined results for tasks."""
        mock_backend.set_task_result("test_task", "predefined_result")
        
        future = LocalFuture()
        mock_task_instance = Mock()
        mock_task_instance.id = "test_123"
        mock_task_instance.task_definition_name = "test_task"
        
        def mock_target_callable(*args, **kwargs):
            return "should_not_be_called"
        
        mock_target_callable.__self__ = mock_task_instance
        
        mock_backend.execute(mock_target_callable, future, (), {})
        
        assert future.result() == "predefined_result"
    
    def test_mock_backend_predefined_exceptions(self, mock_backend):
        """Test setting predefined exceptions for tasks."""
        test_exception = ValueError("Predefined error")
        mock_backend.set_task_exception("failing_task", test_exception)
        
        future = LocalFuture()
        mock_task_instance = Mock()
        mock_task_instance.id = "test_123"
        mock_task_instance.task_definition_name = "failing_task"
        
        def mock_target_callable(*args, **kwargs):
            return "should_not_be_called"
        
        mock_target_callable.__self__ = mock_task_instance
        
        mock_backend.execute(mock_target_callable, future, (), {})
        
        assert future.done
        with pytest.raises(ValueError, match="Predefined error"):
            future.result()
    
    def test_mock_backend_manual_completion(self, mock_backend_manual):
        """Test manual completion mode."""
        future = LocalFuture()
        mock_task_instance = Mock()
        mock_task_instance.id = "manual_task_123"
        mock_task_instance.task_definition_name = "manual_task"
        
        def mock_target_callable(*args, **kwargs):
            return "should_be_manual"
        
        mock_target_callable.__self__ = mock_task_instance
        
        # Execute in manual mode
        mock_backend_manual.execute(mock_target_callable, future, (), {})
        
        # Should be running but not completed
        assert not future.done
        
        executions = mock_backend_manual.get_pending_executions()
        assert len(executions) == 1
        assert executions[0].status == TaskStatus.RUNNING
        
        # Manually complete
        success = mock_backend_manual.complete_task("manual_task_123", result="manual_result")
        assert success
        assert future.done
        assert future.result() == "manual_result"
    
    def test_mock_backend_task_cancellation(self, mock_backend):
        """Test task cancellation."""
        mock_backend.set_default_execution_mode(MockExecutionMode.MANUAL)
        
        future = LocalFuture()
        mock_task_instance = Mock()
        mock_task_instance.id = "cancel_task_123"
        mock_task_instance.task_definition_name = "cancel_task"
        
        def mock_target_callable(*args, **kwargs):
            return "should_be_cancelled"
        
        mock_target_callable.__self__ = mock_task_instance
        
        # Execute
        mock_backend.execute(mock_target_callable, future, (), {})
        assert not future.done
        
        # Cancel
        success = mock_backend.cancel_task_instance("cancel_task_123")
        assert success
        assert future.done
        
        with pytest.raises(BackendError, match="cancelled"):
            future.result()
        
        execution = mock_backend.get_execution("cancel_task_123")
        assert execution.status == TaskStatus.CANCELLED
    
    def test_mock_backend_state_inspection(self, mock_backend):
        """Test backend state inspection capabilities."""
        # Execute multiple tasks
        for i in range(3):
            future = LocalFuture()
            mock_task_instance = Mock()
            mock_task_instance.id = f"task_{i}"
            mock_task_instance.task_definition_name = f"test_task_{i}"
            
            def mock_target_callable(*args, **kwargs):
                return i * 10
            
            mock_target_callable.__self__ = mock_task_instance
            mock_backend.execute(mock_target_callable, future, (), {})
        
        # Check state
        assert mock_backend.get_execution_count() == 3
        
        all_executions = mock_backend.get_all_executions()
        assert len(all_executions) == 3
        
        # Check individual executions
        for i, execution in enumerate(all_executions):
            assert execution.instance_id == f"task_{i}"
            assert execution.result == i * 10


class TestMockTaskControlHandler:
    """Test MockTaskControlHandler functionality."""
    
    def test_mock_handler_initialization(self, mock_handler):
        """Test mock handler initializes correctly."""
        assert isinstance(mock_handler, MockTaskControlHandler)
        assert mock_handler.get_call_counts()['log'] == 0
        assert len(mock_handler.get_log_entries()) == 0
    
    def test_mock_handler_logging(self, mock_handler):
        """Test logging functionality."""
        tc = TaskControl(instance_id="test_123", task_definition_name="test_task")
        
        # Test various log levels
        mock_handler.log(tc, "Info message", "info", extra_field="extra_value")
        mock_handler.log(tc, "Warning message", "warning")
        mock_handler.log(tc, "Error message", "error")
        
        # Check call counts
        call_counts = mock_handler.get_call_counts()
        assert call_counts['log'] == 3
        
        # Check log entries
        entries = mock_handler.get_log_entries("test_123")
        assert len(entries) == 3
        
        info_entry = entries[0]
        assert info_entry.message == "Info message"
        assert info_entry.level == "info"
        assert info_entry.task_id == "test_123"
        assert info_entry.task_name == "test_task"
        assert info_entry.extra_data == {"extra_field": "extra_value"}
        
        # Test filtering by level
        error_entries = mock_handler.get_log_entries("test_123", "error")
        assert len(error_entries) == 1
        assert error_entries[0].message == "Error message"
    
    def test_mock_handler_progress_tracking(self, mock_handler):
        """Test progress update tracking."""
        tc = TaskControl(instance_id="progress_123", task_definition_name="progress_task")
        
        # Send progress updates
        mock_handler.update_progress(tc, 25.0, "started")
        mock_handler.update_progress(tc, 50.0, "halfway")
        mock_handler.update_progress(tc, 100.0, "completed")
        
        # Check call counts
        assert mock_handler.get_call_counts()['update_progress'] == 3
        
        # Check progress history
        progress_history = mock_handler.get_task_progress_history("progress_123")
        assert progress_history == [25.0, 50.0, 100.0]
        
        # Check latest progress
        latest = mock_handler.get_latest_progress("progress_123")
        assert latest.progress == 100.0
        assert latest.status == "completed"
    
    def test_mock_handler_status_tracking(self, mock_handler):
        """Test status update tracking."""
        tc = TaskControl(instance_id="status_123", task_definition_name="status_task")
        
        # Send status updates
        mock_handler.update_status(tc, "starting")
        mock_handler.update_status(tc, "running")
        mock_handler.update_status(tc, "finalizing")
        mock_handler.update_status(tc, "completed")
        
        # Check call counts
        assert mock_handler.get_call_counts()['update_status'] == 4
        
        # Check status history
        status_history = mock_handler.get_task_status_history("status_123")
        assert status_history == ["starting", "running", "finalizing", "completed"]
        
        # Check latest status
        latest = mock_handler.get_latest_status("status_123")
        assert latest.status == "completed"
    
    def test_mock_handler_stop_requests(self, mock_handler):
        """Test stop request handling."""
        tc = TaskControl(instance_id="stop_123", task_definition_name="stop_task")
        
        # Initially should not be stopped
        assert not tc.should_stop
        assert not mock_handler.was_stop_requested("stop_123")
        
        # Request stop
        mock_handler.request_stop(tc)
        
        # Should be stopped
        assert tc.should_stop
        assert mock_handler.was_stop_requested("stop_123")
        assert mock_handler.get_call_counts()['request_stop'] == 1
        
        stop_requests = mock_handler.get_stop_requests("stop_123")
        assert len(stop_requests) == 1
        assert stop_requests[0].task_id == "stop_123"
    
    def test_mock_handler_comprehensive_inspection(self, mock_handler):
        """Test comprehensive inspection capabilities."""
        tc1 = TaskControl(instance_id="task_1", task_definition_name="test_task")
        tc2 = TaskControl(instance_id="task_2", task_definition_name="test_task")
        
        # Generate mixed activity
        mock_handler.log(tc1, "Task 1 starting", "info")
        mock_handler.update_progress(tc1, 50.0, "working")
        mock_handler.log(tc2, "Task 2 starting", "info")
        mock_handler.update_status(tc1, "processing")
        mock_handler.request_stop(tc2)
        
        # Check statistics
        stats = mock_handler.get_statistics()
        assert stats['total_calls'] == 5
        assert stats['total_log_entries'] == 2
        assert stats['total_progress_updates'] == 1
        assert stats['total_status_updates'] == 1
        assert stats['total_stop_requests'] == 1
        assert stats['unique_tasks'] >= 2
        
        # Check message search
        assert mock_handler.has_logged_message("task_1", "starting")
        assert mock_handler.has_logged_message("task_2", "Task 2")
        assert not mock_handler.has_logged_message("task_1", "nonexistent")


class TestMockSessionManager:
    """Test MockSessionManager functionality."""
    
    def test_mock_session_creation(self, mock_session_manager):
        """Test session creation."""
        session_id = mock_session_manager.create_session("test_session", {"key": "value"})
        
        assert session_id is not None
        assert session_id.startswith("mock_session_")
        assert mock_session_manager.get_session_count() == 1
        
        session_info = mock_session_manager.get_session_info(session_id)
        assert session_info is not None
        assert session_info.session_name == "test_session"
        assert session_info.state == MockSessionState.CREATED
        assert session_info.metadata == {"key": "value"}
    
    def test_mock_session_lifecycle(self, mock_session_manager):
        """Test complete session lifecycle."""
        # Create session
        session_id = mock_session_manager.create_session("lifecycle_test")
        session_info = mock_session_manager.get_session_info(session_id)
        assert session_info.state == MockSessionState.CREATED
        
        # Start session
        success = mock_session_manager.start_session(session_id)
        assert success
        session_info = mock_session_manager.get_session_info(session_id)
        assert session_info.state == MockSessionState.ACTIVE
        assert session_info.started_at is not None
        
        # Complete session
        success = mock_session_manager.complete_session(session_id, success=True)
        assert success
        session_info = mock_session_manager.get_session_info(session_id)
        assert session_info.state == MockSessionState.COMPLETED
        assert session_info.completed_at is not None
    
    def test_mock_session_task_management(self, mock_session_manager):
        """Test task instance management in sessions."""
        session_id = mock_session_manager.create_session("task_test")
        mock_session_manager.start_session(session_id)
        
        # Add task instances
        success1 = mock_session_manager.add_task_instance(session_id, "task_1")
        success2 = mock_session_manager.add_task_instance(session_id, "task_2")
        assert success1 and success2
        
        # Check task instances
        task_instances = mock_session_manager.get_session_task_instances(session_id)
        assert len(task_instances) == 2
        assert "task_1" in task_instances
        assert "task_2" in task_instances
        
        # Remove task instance
        success = mock_session_manager.remove_task_instance(session_id, "task_1")
        assert success
        
        task_instances = mock_session_manager.get_session_task_instances(session_id)
        assert len(task_instances) == 1
        assert "task_2" in task_instances
    
    def test_mock_session_operations_tracking(self, mock_session_manager):
        """Test operation tracking."""
        session_id = mock_session_manager.create_session("ops_test")
        mock_session_manager.start_session(session_id)
        mock_session_manager.add_task_instance(session_id, "task_1")
        mock_session_manager.complete_session(session_id)
        
        # Check operations
        operations = mock_session_manager.get_session_operations(session_id)
        assert len(operations) == 4  # create, start, add_task, complete
        
        operation_types = [op.operation for op in operations]
        assert "create" in operation_types
        assert "start" in operation_types
        assert "add_task" in operation_types
        assert "complete" in operation_types
    
    def test_mock_session_statistics(self, mock_session_manager):
        """Test session statistics."""
        # Create multiple sessions in different states
        session1 = mock_session_manager.create_session("session1")
        session2 = mock_session_manager.create_session("session2")
        session3 = mock_session_manager.create_session("session3")
        
        mock_session_manager.start_session(session1)
        mock_session_manager.start_session(session2)
        mock_session_manager.complete_session(session1)
        mock_session_manager.cancel_session(session3)
        
        stats = mock_session_manager.get_statistics()
        assert stats['total_sessions'] == 3
        assert stats['active_session_count'] == 1
        assert stats['session_states']['completed'] == 1
        assert stats['session_states']['active'] == 1
        assert stats['session_states']['cancelled'] == 1


class TestMockEnvironmentIntegration:
    """Test integration of all mock components together."""
    
    def test_complete_mock_environment(self, mock_environment):
        """Test complete mock environment setup."""
        backend, handler, session_manager = mock_environment
        
        assert isinstance(backend, MockExecutionBackend)
        assert isinstance(handler, MockTaskControlHandler)
        assert isinstance(session_manager, MockSessionManager)
        
        # All should be clean
        assert backend.get_execution_count() == 0
        assert handler.get_call_counts()['log'] == 0
        assert session_manager.get_session_count() == 0
    
    def test_mock_task_execution_with_session(self, mock_environment):
        """Test task execution in mock environment with session tracking."""
        backend, handler, session_manager = mock_environment
        
        # Create session
        session_id = session_manager.create_session("integration_test")
        session_manager.start_session(session_id)
        
        # Mock a task instance that uses our mock backend
        with patch('numerous.tasks.backends.get_backend', return_value=backend):
            # This would normally be done by the task system
            # We're simulating the integration here
            
            future = LocalFuture()
            mock_task_instance = Mock()
            mock_task_instance.id = "integration_task_123"
            mock_task_instance.task_definition_name = "integration_task"
            mock_task_instance.task_control = TaskControl(
                instance_id="integration_task_123",
                task_definition_name="integration_task"
            )
            
            def mock_target_callable(*args, **kwargs):
                # Simulate task that uses TaskControl
                tc = mock_task_instance.task_control
                tc.log("Integration test started", "info")
                tc.update_progress(50.0, "working")
                tc.update_status("processing")
                tc.update_progress(100.0, "completed")
                return "integration_success"
            
            mock_target_callable.__self__ = mock_task_instance
            
            # Add task to session
            session_manager.add_task_instance(session_id, "integration_task_123")
            
            # Execute via mock backend
            backend.execute(mock_target_callable, future, (), {})
            
            # Verify execution
            assert future.done
            assert future.result() == "integration_success"
            
            # Verify backend tracking
            assert backend.get_execution_count() == 1
            execution = backend.get_execution("integration_task_123")
            assert execution.status == TaskStatus.COMPLETED
            
            # Verify handler tracking
            assert handler.get_call_counts()['log'] >= 1
            assert handler.get_call_counts()['update_progress'] >= 1
            assert handler.get_call_counts()['update_status'] >= 1
            
            # Verify session tracking
            task_instances = session_manager.get_session_task_instances(session_id)
            assert "integration_task_123" in task_instances


class TestBackendSwitching:
    """Test backend switching capabilities (Mock ↔ Local)."""
    
    def test_mock_to_local_backend_switch(self, mock_backend):
        """Test switching from mock to local backend."""
        from numerous.tasks.backends.local import LocalExecutionBackend
        
        # Start with mock backend
        assert isinstance(mock_backend, MockExecutionBackend)
        
        # Create local backend
        local_backend = LocalExecutionBackend(max_workers=2)
        local_backend.startup()
        
        try:
            # Both should be functional
            assert mock_backend._is_started
            assert local_backend._executor is not None
            
            # Different behavior expectations
            mock_backend.set_default_execution_mode(MockExecutionMode.IMMEDIATE)
            
            # Mock backend should track execution
            future_mock = LocalFuture()
            mock_task_instance = Mock()
            mock_task_instance.id = "switch_test_mock"
            mock_task_instance.task_definition_name = "switch_test"
            
            def mock_target(*args, **kwargs):
                return "mock_result"
            
            mock_target.__self__ = mock_task_instance
            
            mock_backend.execute(mock_target, future_mock, (), {})
            assert future_mock.result() == "mock_result"
            assert mock_backend.get_execution_count() == 1
            
        finally:
            local_backend.shutdown()
    
    def test_handler_switching(self):
        """Test switching between mock and local task control handlers."""
        from numerous.tasks.control import LocalTaskControlHandler
        
        # Get original handler
        original_handler = get_task_control_handler()
        
        # Switch to mock handler
        mock_handler = MockTaskControlHandler()
        set_task_control_handler(mock_handler)
        
        # Create TaskControl - should use mock handler
        tc = TaskControl(instance_id="switch_test", task_definition_name="test")
        
        # Test with mock handler
        tc.log("Mock test message", "info")
        assert mock_handler.get_call_counts()['log'] == 1
        
        # Switch back to local handler
        local_handler = LocalTaskControlHandler()
        set_task_control_handler(local_handler)
        
        # Create new TaskControl - should use local handler
        tc2 = TaskControl(instance_id="switch_test2", task_definition_name="test")
        
        # Test with local handler (should not affect mock handler counts)
        tc2.log("Local test message", "info")
        assert mock_handler.get_call_counts()['log'] == 1  # Unchanged
        
        # Restore original
        set_task_control_handler(original_handler) 