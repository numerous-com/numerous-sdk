"""
Test Stage 1: Session context manager for direct function calls

This module tests whether Session context management works properly for
direct task execution without backend dependencies.
"""

import pytest
from numerous.tasks import task, TaskControl, Session


class TestDirectExecutionSession:
    """Test Session context manager in Stage 1 direct execution."""
    
    def test_direct_execution_without_session_context(self):
        """Test that direct execution works without requiring Session context."""
        @task
        def simple_task(x: int) -> int:
            return x * 2
        
        # Should work without any Session context
        result = simple_task(5)
        assert result == 10
    
    def test_direct_execution_with_taskcontrol_without_session(self):
        """Test that TaskControl works in direct execution without Session context."""
        @task
        def task_with_control(tc: TaskControl, message: str) -> str:
            tc.log(f"Processing: {message}")
            tc.update_progress(100.0, "completed")
            return f"Processed: {message}"
        
        # Should work without any Session context
        result = task_with_control("test")
        assert result == "Processed: test"
    
    def test_session_context_manager_basic_functionality(self):
        """Test that Session context manager works correctly."""
        assert Session.current() is None
        
        session = Session(name="test_session")
        assert session.name == "test_session"
        assert session.id is not None
        assert not session._is_active
        
        with session:
            assert Session.current() == session
            assert session._is_active
        
        assert Session.current() is None
        assert not session._is_active
    
    def test_nested_session_contexts(self):
        """Test nested session context management."""
        session1 = Session(name="outer")
        session2 = Session(name="inner")
        
        assert Session.current() is None
        
        with session1:
            assert Session.current() == session1
            
            with session2:
                assert Session.current() == session2
            
            assert Session.current() == session1
        
        assert Session.current() is None
    
    def test_session_re_entry_error(self):
        """Test that re-entering an active session raises error."""
        session = Session()
        
        with session:
            with pytest.raises(Exception):  # Should be SessionError
                with session:
                    pass
    
    def test_direct_execution_optionally_uses_session_context(self):
        """Test that direct execution can optionally use Session context."""
        captured_sessions = []
        
        @task
        def session_aware_task(tc: TaskControl, message: str) -> dict:
            current_session = Session.current()
            captured_sessions.append(current_session)
            return {
                "message": message,
                "has_session": current_session is not None,
                "session_name": current_session.name if current_session else None
            }
        
        # Test without session
        result1 = session_aware_task("no session")
        assert result1["has_session"] == False
        assert result1["session_name"] is None
        
        # Test with session
        with Session(name="test_session") as session:
            result2 = session_aware_task("with session")
            assert result2["has_session"] == True
            assert result2["session_name"] == "test_session"
        
        # Back to no session
        result3 = session_aware_task("no session again")
        assert result3["has_session"] == False
        assert result3["session_name"] is None
    
    def test_multiple_tasks_in_session_context(self):
        """Test multiple tasks can be executed within the same session context."""
        results = []
        
        @task
        def task_a(tc: TaskControl, value: int) -> str:
            session = Session.current()
            results.append(f"task_a: {session.name if session else 'no_session'}: {value}")
            return f"A{value}"
        
        @task  
        def task_b(tc: TaskControl, value: int) -> str:
            session = Session.current()
            results.append(f"task_b: {session.name if session else 'no_session'}: {value}")
            return f"B{value}"
        
        with Session(name="shared_session"):
            result_a = task_a(1)
            result_b = task_b(2)
            
        assert result_a == "A1"
        assert result_b == "B2"
        assert len(results) == 2
        assert "task_a: shared_session: 1" in results
        assert "task_b: shared_session: 2" in results
    
    def test_session_isolation_between_calls(self):
        """Test that session contexts are properly isolated."""
        captured_context = []
        
        @task
        def context_capture_task(tc: TaskControl) -> str:
            session = Session.current()
            captured_context.append({
                "session_id": session.id if session else None,
                "session_name": session.name if session else None
            })
            return "captured"
        
        # First session
        with Session(name="session_1") as s1:
            context_capture_task()
        
        # Second session
        with Session(name="session_2") as s2:
            context_capture_task()
        
        # No session
        context_capture_task()
        
        assert len(captured_context) == 3
        assert captured_context[0]["session_name"] == "session_1"
        assert captured_context[1]["session_name"] == "session_2"
        assert captured_context[2]["session_name"] is None
        
        # Session IDs should be different
        assert captured_context[0]["session_id"] != captured_context[1]["session_id"]
    
    def test_session_task_tracking_in_direct_mode(self):
        """Test that session can track tasks even in direct execution mode."""
        # This test explores if we should track direct execution tasks in sessions
        
        @task
        def trackable_task(tc: TaskControl, value: int) -> int:
            return value * 2
        
        session = Session(name="tracking_session")
        
        with session:
            # Direct execution - should it be tracked?
            result = trackable_task(5)
            assert result == 10
            
            # Session should be available but no automatic task tracking in direct mode
            # This is expected behavior since direct execution bypasses TaskInstance
            assert len(session.tasks) == 0
    
    def test_session_with_max_parallel_tasks_direct_execution(self):
        """Test that max_parallel constraints work differently in direct vs instance execution."""
        
        @task(max_parallel=1)
        def constrained_task(value: int) -> int:
            return value * 2
        
        # Direct execution should work regardless of session context
        # because it doesn't go through TaskInstance.start()
        with Session(name="test_session"):
            result1 = constrained_task(1)
            result2 = constrained_task(2)  # Should work in direct mode
            assert result1 == 2
            assert result2 == 4
    
    def test_session_factory_methods(self):
        """Test Session creation with different options."""
        # Default session
        session1 = Session()
        assert session1.name.startswith("session_")
        assert session1.id is not None
        
        # Named session
        session2 = Session(name="custom_name")
        assert session2.name == "custom_name"
        assert session2.id is not None
        
        # Session with specific ID
        custom_id = "custom-session-id"
        session3 = Session(name="test", session_id=custom_id)
        assert session3.name == "test"
        assert session3.id == custom_id 