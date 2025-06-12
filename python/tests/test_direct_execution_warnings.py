"""
Test Stage 1: Session warning for direct execution

This module tests that appropriate warnings are raised when tasks execute
directly despite having an active session context.
"""

import pytest
import warnings
from numerous.tasks import task, TaskControl, Session


class TestDirectExecutionWarnings:
    """Test warning behavior when direct execution happens with active session."""
    
    def test_no_warning_without_session(self):
        """Test that no warning is raised when executing without session."""
        @task
        def simple_task(x: int) -> int:
            return x * 2
        
        # Should not raise any warnings
        with warnings.catch_warnings(record=True) as w:
            warnings.simplefilter("always")
            result = simple_task(5)
            assert result == 10
            assert len(w) == 0
    
    def test_warning_with_active_session(self):
        """Test that warning is raised when executing with active session."""
        @task
        def task_with_session(x: int) -> int:
            return x * 3
        
        # Should raise warning when session is active
        with warnings.catch_warnings(record=True) as w:
            warnings.simplefilter("always")
            
            with Session(name="test_session"):
                result = task_with_session(4)
                assert result == 12
                
                # Check that warning was raised
                assert len(w) == 1
                assert issubclass(w[0].category, UserWarning)
                warning_message = str(w[0].message)
                assert "task_with_session" in warning_message
                assert "executing directly despite active session" in warning_message
                assert "test_session" in warning_message
                assert "Use task.instance().start()" in warning_message
    
    def test_warning_with_taskcontrol_and_session(self):
        """Test that warning is raised for TaskControl tasks with active session."""
        @task
        def task_with_control(tc: TaskControl, message: str) -> str:
            tc.log("Processing message")
            return f"Processed: {message}"
        
        with warnings.catch_warnings(record=True) as w:
            warnings.simplefilter("always")
            
            with Session(name="control_session"):
                result = task_with_control("test")
                assert result == "Processed: test"
                
                # Should still warn about direct execution
                assert len(w) == 1
                assert "task_with_control" in str(w[0].message)
                assert "control_session" in str(w[0].message)
    
    def test_warning_with_nested_sessions(self):
        """Test warning behavior with nested sessions."""
        @task
        def nested_task(x: int) -> int:
            return x + 1
        
        with warnings.catch_warnings(record=True) as w:
            warnings.simplefilter("always")
            
            with Session(name="outer_session"):
                with Session(name="inner_session"):
                    result = nested_task(10)
                    assert result == 11
                    
                    # Should warn about the inner session (current session)
                    assert len(w) == 1
                    assert "inner_session" in str(w[0].message)
    
    def test_warning_with_multiple_tasks_in_session(self):
        """Test that warning is raised for each task execution in session."""
        @task
        def first_task(x: int) -> int:
            return x * 2
        
        @task
        def second_task(x: int) -> int:
            return x + 5
        
        with warnings.catch_warnings(record=True) as w:
            warnings.simplefilter("always")
            
            with Session(name="multi_task_session"):
                result1 = first_task(3)
                result2 = second_task(7)
                
                assert result1 == 6
                assert result2 == 12
                
                # Should have two warnings, one for each task
                assert len(w) == 2
                assert "first_task" in str(w[0].message)
                assert "second_task" in str(w[1].message)
                assert all("multi_task_session" in str(warning.message) for warning in w)
    
    def test_no_warning_with_instance_start(self):
        """Test that no warning is raised when using instance().start() with session."""
        @task
        def instance_task(x: int) -> int:
            return x * 4
        
        with warnings.catch_warnings(record=True) as w:
            warnings.simplefilter("always")
            
            with Session(name="instance_session"):
                # Using instance().start() should not trigger warning
                instance = instance_task.instance()
                future = instance.start(5)
                result = future.result()
                
                assert result == 20
                # No warnings should be raised for proper session usage
                assert len(w) == 0
    
    def test_warning_message_format(self):
        """Test that warning message contains all expected information."""
        @task(name="custom_task_name")
        def formatted_task() -> str:
            return "done"
        
        with warnings.catch_warnings(record=True) as w:
            warnings.simplefilter("always")
            
            with Session(name="format_session"):
                result = formatted_task()
                assert result == "done"
                
                assert len(w) == 1
                warning_message = str(w[0].message)
                
                # Check all expected components in warning message
                assert "custom_task_name" in warning_message
                assert "executing directly despite active session" in warning_message
                assert "format_session" in warning_message
                assert "Direct execution bypasses session tracking" in warning_message
                assert "Use task.instance().start()" in warning_message
    
    def test_warning_stacklevel(self):
        """Test that warning points to the correct location in user code."""
        @task
        def stacklevel_task() -> str:
            return "test"
        
        with warnings.catch_warnings(record=True) as w:
            warnings.simplefilter("always")
            
            with Session():
                stacklevel_task()  # This line should be identified in warning
                
            assert len(w) == 1
            # The warning should point to the line where the task was called
            # (stacklevel=2 should skip the task.__call__ frame)
            assert w[0].filename.endswith("test_direct_execution_warnings.py")
    
    def test_warning_can_be_suppressed(self):
        """Test that the warning can be suppressed if users choose to."""
        @task
        def suppressible_task() -> int:
            return 42
        
        # Test that warning can be filtered out
        with warnings.catch_warnings():
            warnings.filterwarnings("ignore", category=UserWarning, 
                                  message=".*executing directly despite active session.*")
            
            with Session():
                result = suppressible_task()
                assert result == 42
                # No exception should be raised, warning should be suppressed 