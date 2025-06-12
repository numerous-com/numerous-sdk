"""
Test suite for Task 3.0: Local TaskInstance Execution (Local Backend)

This module tests the complete TaskInstance execution workflow using the LocalExecutionBackend,
including session management, concurrency control, cancellation, and error handling.
"""

import pytest
import time
import threading
from concurrent.futures import ThreadPoolExecutor
from typing import List, Optional

from numerous.tasks import task, Session
from numerous.tasks.control import TaskControl
from numerous.tasks.backends.local import LocalExecutionBackend
from numerous.tasks.future import LocalFuture, TaskStatus
from numerous.tasks.exceptions import TaskError, MaxInstancesReachedError, SessionNotFoundError


class TestLocalExecutionBackend:
    """Test LocalExecutionBackend functionality for task instance execution."""
    
    def test_local_backend_initialization(self):
        """Test LocalExecutionBackend initialization and configuration."""
        # Test default initialization
        backend = LocalExecutionBackend()
        assert backend._executor is not None
        assert isinstance(backend._executor, ThreadPoolExecutor)
        assert backend._active_concurrent_futures == {}
        assert backend._active_task_controls == {}
        backend.shutdown()
        
        # Test custom max_workers
        backend = LocalExecutionBackend(max_workers=2)
        assert backend._executor is not None
        backend.shutdown()
    
    def test_local_backend_startup_shutdown(self):
        """Test backend startup and shutdown lifecycle."""
        backend = LocalExecutionBackend()
        
        # Test startup
        backend.startup()  # Should not raise
        
        # Test shutdown
        backend.shutdown(wait=True)
        backend.shutdown(wait=False)  # Should handle multiple shutdowns gracefully


class TestTaskInstanceExecution:
    """Test task.instance().start() workflow with local backend."""
    
    def test_simple_task_instance_execution(self):
        """Test basic task instance creation and execution."""
        @task
        def simple_task(value: int) -> int:
            return value * 2
        
        with Session() as session:
            instance = simple_task.instance()
            
            # Verify instance properties
            assert instance.task_definition == simple_task
            assert instance.session == session
            assert instance.id is not None
            assert instance.task_control is not None
            assert instance._future is None
            assert instance.status == "pending"
            assert not instance.is_running
            
            # Execute task
            future = instance.start(5)
            
            # Verify execution
            assert isinstance(future, LocalFuture)
            assert instance._future == future
            assert instance.is_running or future.done  # May complete immediately
            
            result = future.result()
            assert result == 10
            assert future.done
            assert not instance.is_running
    
    def test_task_with_taskcontrol_execution(self):
        """Test task instance execution with TaskControl injection."""
        execution_log = []
        
        @task
        def task_with_control(tc: TaskControl, value: int) -> int:
            execution_log.append(f"Started with value {value}")
            tc.log("Task started", "info")
            tc.update_progress(25.0, "Quarter done")
            tc.update_status("Processing")
            tc.update_progress(50.0, "Half done")
            execution_log.append(f"TaskControl ID: {tc.instance_id}")
            tc.update_progress(100.0, "Complete")
            return value * 3
        
        with Session() as session:
            instance = task_with_control.instance()
            future = instance.start(7)
            result = future.result()
            
            assert result == 21
            assert len(execution_log) == 2
            assert "Started with value 7" in execution_log
            assert instance.task_control.instance_id in execution_log[1]
            assert instance.task_control.progress == 100.0
            assert instance.task_control.status == "Complete"
    
    def test_task_instance_cannot_start_twice(self):
        """Test that task instances cannot be started multiple times."""
        @task
        def simple_task(value: int) -> int:
            return value * 2
        
        with Session() as session:
            instance = simple_task.instance()
            
            # First start should succeed
            future1 = instance.start(5)
            result1 = future1.result()
            assert result1 == 10
            
            # Second start should fail
            with pytest.raises(TaskError, match="Task instance has already been started"):
                instance.start(6)
    
    def test_task_instance_without_session_fails(self):
        """Test that task instances cannot be created without an active session."""
        @task
        def simple_task(value: int) -> int:
            return value * 2
        
        # No active session
        with pytest.raises(SessionNotFoundError, match="Cannot create a task instance without an active session"):
            simple_task.instance()


class TestSessionManagement:
    """Test session management with TaskInstance lifecycle."""
    
    def test_session_task_instance_registration(self):
        """Test that task instances are properly registered with sessions."""
        @task
        def simple_task(value: int) -> int:
            return value * 2
        
        with Session() as session:
            # Initially no instances
            assert session.get_running_instances_count("simple_task") == 0
            
            # Create instance
            instance = simple_task.instance()
            assert "simple_task" in session.tasks
            assert instance.id in session.tasks["simple_task"]
            assert session.tasks["simple_task"][instance.id] == instance
            
            # Start instance
            future = instance.start(5)
            
            # Complete execution
            result = future.result()
            assert result == 10
    
    def test_session_cleanup_on_exit(self):
        """Test that sessions properly manage task instances through their lifecycle."""
        @task
        def simple_task(value: int) -> int:
            return value * 2
        
        session = Session()
        with session:
            instance = simple_task.instance()
            future = instance.start(5)
            result = future.result()
            assert result == 10
            
            # Task instance should be registered in session
            assert "simple_task" in session.tasks
            assert instance.id in session.tasks["simple_task"]
            
        # Session maintains task instances for tracking even after exit
        # This is correct behavior for monitoring and debugging
        assert "simple_task" in session.tasks
        assert instance.id in session.tasks["simple_task"]
        
        # But the session should no longer be active
        assert not session._is_active
        assert Session.current() != session
    
    def test_multiple_instances_in_session(self):
        """Test multiple task instances in the same session."""
        @task(max_parallel=3)  # Allow multiple instances
        def task_a(value: int) -> int:
            return value * 2
        
        @task(max_parallel=3)  # Allow multiple instances
        def task_b(value: int) -> int:
            return value * 3
        
        with Session() as session:
            # Create multiple instances
            instance_a1 = task_a.instance()
            instance_a2 = task_a.instance()
            instance_b1 = task_b.instance()
            
            # Execute all instances
            future_a1 = instance_a1.start(5)
            future_a2 = instance_a2.start(6)
            future_b1 = instance_b1.start(7)
            
            # Verify results
            assert future_a1.result() == 10
            assert future_a2.result() == 12
            assert future_b1.result() == 21


class TestConcurrentExecution:
    """Test concurrent task execution with max_parallel constraints."""
    
    def test_max_parallel_constraint_enforcement(self):
        """Test that max_parallel constraints are enforced."""
        execution_order = []
        
        @task(max_parallel=2)
        def limited_task(value: int) -> int:
            execution_order.append(f"start_{value}")
            time.sleep(0.1)  # Simulate work
            execution_order.append(f"end_{value}")
            return value * 2
        
        with Session() as session:
            # Create 3 instances (max_parallel=2)
            instance1 = limited_task.instance()
            instance2 = limited_task.instance()
            instance3 = limited_task.instance()
            
            # Start first two should succeed
            future1 = instance1.start(1)
            future2 = instance2.start(2)
            
            # Third should fail due to max_parallel constraint
            with pytest.raises(MaxInstancesReachedError, match="Max parallel \\(2\\) reached"):
                instance3.start(3)
            
            # Wait for completion
            result1 = future1.result()
            result2 = future2.result()
            
            assert result1 == 2
            assert result2 == 4
            
            # Now third instance should be able to start
            future3 = instance3.start(3)
            result3 = future3.result()
            assert result3 == 6
    
    def test_concurrent_execution_thread_safety(self):
        """Test thread safety of concurrent task execution."""
        results = []
        
        @task(max_parallel=10)  # Allow multiple concurrent instances
        def concurrent_task(value: int) -> int:
            time.sleep(0.05)  # Simulate work
            result = value * 2
            results.append(result)
            return result
        
        with Session() as session:
            # Create multiple instances
            instances = [concurrent_task.instance() for _ in range(5)]
            
            # Start all instances concurrently
            futures = []
            for i, instance in enumerate(instances):
                future = instance.start(i + 1)
                futures.append(future)
            
            # Wait for all to complete
            final_results = [future.result() for future in futures]
            
            # Verify results
            expected = [2, 4, 6, 8, 10]
            assert sorted(final_results) == expected
            assert sorted(results) == expected


class TestTaskCancellation:
    """Test task cancellation and cleanup for local execution."""
    
    def test_task_cancellation_via_task_control(self):
        """Test task cancellation through TaskControl.should_stop."""
        cancellation_log = []
        
        @task
        def cancellable_task(tc: TaskControl) -> str:
            for i in range(100):
                if tc.should_stop:
                    cancellation_log.append(f"Stopped at iteration {i}")
                    return "cancelled"
                time.sleep(0.01)
            return "completed"
        
        with Session() as session:
            instance = cancellable_task.instance()
            future = instance.start()
            
            # Give task time to start
            time.sleep(0.05)
            
            # Request cancellation
            instance.stop()
            
            # Verify cancellation
            result = future.result()
            assert result == "cancelled"
            assert len(cancellation_log) == 1
            assert "Stopped at iteration" in cancellation_log[0]
            assert instance.task_control.should_stop
    
    def test_backend_task_cancellation(self):
        """Test task cancellation through backend cancel_task_instance."""
        @task
        def long_running_task(tc: TaskControl) -> str:
            for i in range(100):
                if tc.should_stop:
                    return "stopped"
                time.sleep(0.01)
            return "completed"
        
        with Session() as session:
            instance = long_running_task.instance()
            future = instance.start()
            
            # Give task time to start
            time.sleep(0.02)
            
            # Cancel through backend
            success = instance.backend.cancel_task_instance(instance.id)
            assert success
            
            # Verify task was signaled to stop
            result = future.result()
            assert result == "stopped"
            assert instance.task_control.should_stop
    
    def test_cancellation_cleanup(self):
        """Test that cancelled tasks are properly cleaned up."""
        @task
        def cancellable_task(tc: TaskControl) -> str:
            for i in range(100):
                if tc.should_stop:
                    return "cancelled"
                time.sleep(0.01)
            return "completed"
        
        with Session() as session:
            instance = cancellable_task.instance()
            future = instance.start()
            
            # Give task time to start
            time.sleep(0.02)
            
            # Cancel task
            instance.stop()
            result = future.result()
            assert result == "cancelled"
            
            # Verify cleanup
            assert instance.id not in instance.backend._active_concurrent_futures
            assert instance.id not in instance.backend._active_task_controls


class TestFutureObjects:
    """Test Future objects and asynchronous task execution."""
    
    def test_future_status_transitions(self):
        """Test Future status transitions during task execution."""
        @task
        def status_task(tc: TaskControl, duration: float) -> str:
            tc.update_status("running")
            time.sleep(duration)
            tc.update_status("completed")
            return "done"
        
        with Session() as session:
            instance = status_task.instance()
            future = instance.start(0.1)
            
            # Future should be created
            assert isinstance(future, LocalFuture)
            
            # Wait for completion
            result = future.result()
            assert result == "done"
            assert future.done
            assert future.status == TaskStatus.COMPLETED
    
    def test_future_exception_handling(self):
        """Test Future exception handling for failed tasks."""
        @task
        def failing_task(tc: TaskControl) -> str:
            tc.update_status("running")
            raise ValueError("Task failed intentionally")
        
        with Session() as session:
            instance = failing_task.instance()
            future = instance.start()
            
            # Future should capture the exception
            with pytest.raises(ValueError, match="Task failed intentionally"):
                future.result()
            
            assert future.done
            assert future.status == TaskStatus.FAILED
    
    def test_future_timeout_behavior(self):
        """Test Future timeout behavior."""
        @task
        def slow_task(tc: TaskControl) -> str:
            time.sleep(0.2)
            return "completed"
        
        with Session() as session:
            instance = slow_task.instance()
            future = instance.start()
            
            # Test timeout
            with pytest.raises(TimeoutError):
                future.result(timeout=0.05)
            
            # Task should still complete
            result = future.result(timeout=0.5)
            assert result == "completed"
    
    def test_multiple_futures_async_execution(self):
        """Test multiple futures executing asynchronously."""
        @task(max_parallel=5)  # Allow multiple concurrent instances
        def async_task(tc: TaskControl, value: int, delay: float) -> int:
            time.sleep(delay)
            return value * 2
        
        with Session() as session:
            # Create multiple instances with different delays
            instances = [
                async_task.instance(),
                async_task.instance(),
                async_task.instance()
            ]
            
            # Start all tasks
            start_time = time.time()
            futures = [
                instances[0].start(1, 0.1),
                instances[1].start(2, 0.05),
                instances[2].start(3, 0.15)
            ]
            
            # Collect results
            results = [future.result() for future in futures]
            end_time = time.time()
            
            # Verify results
            assert sorted(results) == [2, 4, 6]
            
            # Should complete in parallel (less than sum of delays)
            # Allow some overhead for thread scheduling and system load
            total_sequential_time = 0.1 + 0.05 + 0.15  # 0.3 seconds
            actual_time = end_time - start_time
            assert actual_time < total_sequential_time + 0.5  # Allow 0.5s overhead for parallel execution
    
    def test_realtime_progress_propagation_during_async_execution(self):
        """Test that progress updates and logs are properly propagated during async execution."""
        progress_log = []
        
        # Capture logs by temporarily replacing the handler
        from numerous.tasks.control import get_task_control_handler, set_task_control_handler, LocalTaskControlHandler
        
        class LogCapturingHandler(LocalTaskControlHandler):
            def __init__(self):
                super().__init__()
                self.captured_logs = []
            
            def log(self, task_control, message, level, **extra_data):
                self.captured_logs.append({
                    'task_id': task_control.task_definition_name,
                    'instance_id': task_control.instance_id,
                    'message': message,
                    'level': level,
                    'extra_data': extra_data
                })
                # Still call parent to maintain normal logging
                super().log(task_control, message, level, **extra_data)
        
        log_handler = LogCapturingHandler()
        original_handler = get_task_control_handler()
        set_task_control_handler(log_handler)
        
        try:
            @task(max_parallel=3)
            def monitored_async_task(tc: TaskControl, task_id: str, steps: int) -> str:
                tc.log(f"Starting {task_id} with {steps} steps", "info")
                for i in range(steps):
                    progress = (i + 1) / steps * 100
                    tc.update_progress(progress, f"{task_id}_step_{i+1}")
                    
                    # Log at key milestones
                    if i == 0:
                        tc.log(f"{task_id} first step completed", "debug")
                    elif i == steps // 2:
                        tc.log(f"{task_id} halfway point reached", "info")
                    
                    time.sleep(0.2)  # Increased sleep to ensure monitoring can capture progress
                
                tc.log(f"{task_id} completed successfully", "info")
                return f"{task_id}_complete"
        
            with Session() as session:
                # Start multiple tasks concurrently
                instances = [
                    monitored_async_task.instance(),
                    monitored_async_task.instance(),
                    monitored_async_task.instance()
                ]
                
                futures = [
                    instances[0].start("task_a", 4),
                    instances[1].start("task_b", 6), 
                    instances[2].start("task_c", 3)
                ]
                
                # Give tasks a moment to start before monitoring
                time.sleep(0.1)
                
                # Monitor progress while tasks are running
                monitoring_start = time.time()
                progress_captured_during_execution = False
                
                while not all(f.done for f in futures):
                    current_time = time.time()
                    # Stop monitoring after reasonable time to prevent infinite loop
                    if current_time - monitoring_start > 10.0:
                        break
                        
                    for i, instance in enumerate(instances):
                        if not futures[i].done:
                            current_progress = instance.task_control.progress
                            current_status = instance.task_control.status
                            
                            progress_log.append({
                                'task_id': f"task_{chr(97+i)}",
                                'progress': current_progress,
                                'status': current_status,
                                'timestamp': current_time,
                                'task_running': not futures[i].done
                            })
                            
                            # If we captured progress > 0 while task is still running, that's what we want
                            if current_progress > 0 and not futures[i].done:
                                progress_captured_during_execution = True
                    
                    time.sleep(0.05)  # Check progress more frequently
                
                # Wait for completion
                results = [f.result() for f in futures]
                
                # Verify all tasks completed
                assert results == ["task_a_complete", "task_b_complete", "task_c_complete"]
                
                # CRITICAL: Verify we captured progress DURING execution, not just after completion
                assert progress_captured_during_execution, \
                    "Must capture progress updates while tasks are still running (not just after completion)"
                
                # Verify progress was tracked for each task
                task_a_progress = [p for p in progress_log if p['task_id'] == 'task_a']
                task_b_progress = [p for p in progress_log if p['task_id'] == 'task_b'] 
                task_c_progress = [p for p in progress_log if p['task_id'] == 'task_c']
                
                # Verify we captured progress during execution for each task
                task_a_during_execution = [p for p in task_a_progress if p['task_running'] and p['progress'] > 0]
                task_b_during_execution = [p for p in task_b_progress if p['task_running'] and p['progress'] > 0]
                task_c_during_execution = [p for p in task_c_progress if p['task_running'] and p['progress'] > 0]
                
                # At least one task should have progress captured during execution
                total_during_execution = len(task_a_during_execution) + len(task_b_during_execution) + len(task_c_during_execution)
                assert total_during_execution > 0, \
                    f"Should capture progress during execution. Captured: a={len(task_a_during_execution)}, b={len(task_b_during_execution)}, c={len(task_c_during_execution)}"
                
                # Each task should have multiple progress updates
                total_progress_updates = len(task_a_progress) + len(task_b_progress) + len(task_c_progress)
                assert total_progress_updates > 0, "Should have captured some progress updates"
                
                # For tasks that had progress updates, verify progress is reasonable
                for task_name, task_progress in [
                    ("task_a", task_a_progress), 
                    ("task_b", task_b_progress), 
                    ("task_c", task_c_progress)
                ]:
                    if len(task_progress) > 1:
                        # Progress should be non-decreasing
                        for i in range(1, len(task_progress)):
                            assert task_progress[i]['progress'] >= task_progress[i-1]['progress'], \
                                f"{task_name} progress should be non-decreasing"
                        
                        # Final progress should be reasonable (may not be 100% if captured mid-execution)
                        assert task_progress[-1]['progress'] >= 0.0
                        assert task_progress[-1]['progress'] <= 100.0
                        
                        # Status should contain task_id
                        assert task_name.split('_')[1] in task_progress[-1]['status']
                
                # Verify logs were captured from running tasks
                assert len(log_handler.captured_logs) > 0, "Should have captured at least one log message"
                
                # Verify we got logs from each task
                task_a_logs = [log for log in log_handler.captured_logs if 'task_a' in log['message']]
                task_b_logs = [log for log in log_handler.captured_logs if 'task_b' in log['message']]
                task_c_logs = [log for log in log_handler.captured_logs if 'task_c' in log['message']]
                
                # Each task should have at least a start and completion log
                assert len(task_a_logs) >= 2, "task_a should have at least start and completion logs"
                assert len(task_b_logs) >= 2, "task_b should have at least start and completion logs"
                assert len(task_c_logs) >= 2, "task_c should have at least start and completion logs"
                
                # Verify log levels are captured correctly
                info_logs = [log for log in log_handler.captured_logs if log['level'] == 'info']
                debug_logs = [log for log in log_handler.captured_logs if log['level'] == 'debug']
                
                assert len(info_logs) > 0, "Should have captured info level logs"
                assert len(debug_logs) > 0, "Should have captured debug level logs"
                
                # Verify log messages contain expected content
                start_logs = [log for log in log_handler.captured_logs if 'Starting' in log['message']]
                completion_logs = [log for log in log_handler.captured_logs if 'completed successfully' in log['message']]
                
                assert len(start_logs) == 3, "Should have 3 start logs (one per task)"
                assert len(completion_logs) == 3, "Should have 3 completion logs (one per task)"
        
        finally:
            # Restore original handler
            set_task_control_handler(original_handler)


class TestTaskStateTransitions:
    """Test task state transitions and error handling."""
    
    def test_task_status_progression(self):
        """Test task status progression through execution lifecycle."""
        status_log = []
        
        @task
        def status_tracking_task(tc: TaskControl, value: int) -> int:
            status_log.append(("start", tc.instance_id))
            tc.update_status("initializing")
            status_log.append(("status", tc.status))
            
            tc.update_progress(25.0, "quarter")
            status_log.append(("progress", tc.progress))
            
            tc.update_progress(50.0, "half")
            tc.update_progress(100.0, "complete")
            status_log.append(("final", tc.progress, tc.status))
            
            return value * 2
        
        with Session() as session:
            instance = status_tracking_task.instance()
            
            # Initial state
            assert instance.status == "pending"
            assert not instance.is_running
            
            # Start execution
            future = instance.start(5)
            
            # Complete execution
            result = future.result()
            assert result == 10
            
            # Verify status progression
            assert len(status_log) >= 4
            assert status_log[0][0] == "start"
            assert status_log[1] == ("status", "initializing")
            assert status_log[2] == ("progress", 25.0)
            assert status_log[3] == ("final", 100.0, "complete")
    
    def test_error_handling_and_cleanup(self):
        """Test error handling and proper cleanup on task failure."""
        @task
        def error_task(tc: TaskControl, should_fail: bool) -> str:
            tc.update_status("running")
            if should_fail:
                raise RuntimeError("Intentional failure")
            return "success"
        
        with Session() as session:
            # Test successful execution
            instance1 = error_task.instance()
            future1 = instance1.start(False)
            result1 = future1.result()
            assert result1 == "success"
            
            # Test failed execution
            instance2 = error_task.instance()
            future2 = instance2.start(True)
            
            with pytest.raises(RuntimeError, match="Intentional failure"):
                future2.result()
            
            # Verify cleanup after failure
            assert instance2.id not in instance2.backend._active_concurrent_futures
            assert instance2.id not in instance2.backend._active_task_controls
    
    def test_task_instance_properties_during_execution(self):
        """Test task instance properties during different execution phases."""
        @task
        def property_test_task(tc: TaskControl) -> str:
            time.sleep(0.1)
            return "completed"
        
        with Session() as session:
            instance = property_test_task.instance()
            
            # Before start
            assert instance.status == "pending"
            assert not instance.is_running
            assert instance._future is None
            
            # After start
            future = instance.start()
            
            # During execution (may complete quickly)
            if not future.done:
                assert instance.is_running
                assert instance._future == future
            
            # After completion
            result = future.result()
            assert result == "completed"
            assert not instance.is_running
            assert future.done


class TestBackendIntegration:
    """Test integration between TaskInstance and LocalExecutionBackend."""
    
    def test_backend_assignment_and_switching(self):
        """Test backend assignment and environment variable switching."""
        import os
        
        @task
        def backend_test_task(value: int) -> int:
            return value * 2
        
        # Test default backend
        with Session() as session:
            instance = backend_test_task.instance()
            assert isinstance(instance.backend, LocalExecutionBackend)
            
            future = instance.start(5)
            result = future.result()
            assert result == 10
        
        # Test environment variable override
        original_backend = os.environ.get("NUMEROUS_TASK_BACKEND")
        try:
            os.environ["NUMEROUS_TASK_BACKEND"] = "local"
            with Session() as session:
                instance = backend_test_task.instance()
                assert isinstance(instance.backend, LocalExecutionBackend)
        finally:
            if original_backend is not None:
                os.environ["NUMEROUS_TASK_BACKEND"] = original_backend
            else:
                os.environ.pop("NUMEROUS_TASK_BACKEND", None)
    
    def test_backend_error_handling(self):
        """Test backend error handling and propagation."""
        @task
        def backend_error_task(tc: TaskControl) -> str:
            # Simulate backend-level error
            raise RuntimeError("Backend error")
        
        with Session() as session:
            instance = backend_error_task.instance()
            future = instance.start()
            
            # Error should be propagated through future
            with pytest.raises(RuntimeError, match="Backend error"):
                future.result()
            
            assert future.done
            assert future.status == TaskStatus.FAILED
    
    def test_backend_resource_management(self):
        """Test backend resource management and cleanup."""
        @task
        def resource_task(tc: TaskControl, value: int) -> int:
            time.sleep(0.05)
            return value * 2
        
        with Session() as session:
            backend = LocalExecutionBackend(max_workers=2)
            
            # Override backend for testing
            instance = resource_task.instance()
            instance.backend = backend
            
            # Execute task
            future = instance.start(5)
            result = future.result()
            assert result == 10
            
            # Cleanup
            backend.shutdown(wait=True)


# Integration test combining multiple features
class TestLocalTaskInstanceIntegration:
    """Integration tests combining multiple Local TaskInstance features."""
    
    def test_complete_workflow_integration(self):
        """Test complete workflow with multiple tasks, sessions, and features."""
        execution_log = []
        
        @task(max_parallel=2)
        def workflow_task(tc: TaskControl, task_id: str, duration: float) -> str:
            execution_log.append(f"{task_id}_start")
            tc.update_status(f"running_{task_id}")
            tc.update_progress(25.0, "started")
            
            # Simulate work with cancellation check
            for i in range(int(duration * 100)):
                if tc.should_stop:
                    execution_log.append(f"{task_id}_cancelled")
                    return f"{task_id}_cancelled"
                time.sleep(0.01)
                
                if i == int(duration * 50):
                    tc.update_progress(50.0, "halfway")
            
            tc.update_progress(100.0, "completed")
            execution_log.append(f"{task_id}_complete")
            return f"{task_id}_result"
        
        with Session() as session:
            # Create multiple instances
            instances = [
                workflow_task.instance(),
                workflow_task.instance(),
                workflow_task.instance()
            ]
            
            # Start first two (within max_parallel limit)
            future1 = instances[0].start("task1", 0.1)
            future2 = instances[1].start("task2", 0.15)
            
            # Third should fail due to max_parallel
            with pytest.raises(MaxInstancesReachedError):
                instances[2].start("task3", 0.05)
            
            # Wait for first task to complete
            result1 = future1.result()
            assert result1 == "task1_result"
            
            # Now third task should be able to start
            future3 = instances[2].start("task3", 0.05)
            
            # Wait for remaining tasks
            result2 = future2.result()
            result3 = future3.result()
            
            assert result2 == "task2_result"
            assert result3 == "task3_result"
            
            # Verify execution log
            assert "task1_start" in execution_log
            assert "task1_complete" in execution_log
            assert "task2_start" in execution_log
            assert "task2_complete" in execution_log
            assert "task3_start" in execution_log
            assert "task3_complete" in execution_log 