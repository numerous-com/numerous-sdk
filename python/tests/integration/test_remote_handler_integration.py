"""
Integration tests for RemoteTaskControlHandler.

These tests require the --integration flag and assume API availability at localhost:8080.
They test actual API communication without managing Docker Compose lifecycle.
"""

import pytest
import time
import logging
from unittest.mock import Mock, patch
from typing import Optional, Dict, Any

from numerous.tasks.control import TaskControl, set_task_control_handler
from numerous.tasks.integration.remote_handler import RemoteTaskControlHandler
from numerous.tasks.task import task


@pytest.mark.integration
class TestRemoteHandlerIntegration:
    """Integration tests for RemoteTaskControlHandler with real API communication."""
    
    def setup_method(self):
        """Set up test environment before each test."""
        self.original_handler = None
        self.test_session_id = f"test_session_{int(time.time())}"
        self.test_task_instance_id = f"test_task_{int(time.time())}"
    
    def teardown_method(self):
        """Clean up after each test."""
        # Reset to original handler
        set_task_control_handler(self.original_handler)
    
    def test_remote_handler_initialization_with_api(self, integration_config):
        """Test RemoteTaskControlHandler initialization with API available."""
        # RemoteTaskControlHandler will fail at initialization if API is not available
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        
        # If we reach here, API connection was successful
        assert handler is not None
        assert handler.session_id == self.test_session_id
        assert handler.is_connected
        assert handler._client is not None
    
    def test_remote_handler_initialization_fallback(self):
        """Test RemoteTaskControlHandler falls back gracefully when API unavailable."""
        # Mock the get_client to raise an exception
        with patch('numerous.organization.get_client') as mock_get_client:
            mock_get_client.side_effect = Exception("API not available")
            
            handler = RemoteTaskControlHandler(session_id=self.test_session_id)
            
            # Should initialize in fallback mode
            assert handler is not None
            assert not handler.is_connected
            assert handler._client is None
    
    def test_remote_handler_log_with_api_connection(self, integration_config):
        """Test logging through RemoteTaskControlHandler with API connection."""
        # RemoteTaskControlHandler will fail at initialization if API is not available
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        
        tc = TaskControl(
            instance_id=self.test_task_instance_id,
            task_definition_name="test_task"
        )
        
        # Test logging with actual API communication
        handler.log(tc, "Test log message", "info", test_data="integration_test")
        handler.log(tc, "Debug message", "debug", component="test_handler")
        handler.log(tc, "Warning message", "warning", retry_count=1)
        handler.log(tc, "Error message", "error", error_code="TEST001")
        
        # All log calls should complete without exceptions
        assert True  # If we reach here, logging worked
    

    
    def test_remote_handler_progress_updates(self, integration_config):
        """Test progress updates through RemoteTaskControlHandler."""
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        
        tc = TaskControl(
            instance_id=self.test_task_instance_id,
            task_definition_name="test_task"
        )
        
        # Test progress updates with actual API communication
        handler.update_progress(tc, 0.0, "Starting")
        handler.update_progress(tc, 25.0, "Quarter complete")
        handler.update_progress(tc, 50.0, "Halfway done")
        handler.update_progress(tc, 75.0, "Almost finished")
        handler.update_progress(tc, 100.0, "Complete")
        
        # All progress updates should complete without exceptions
        assert True
    
    def test_remote_handler_status_updates(self, integration_config):
        """Test status updates through RemoteTaskControlHandler."""
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        
        tc = TaskControl(
            instance_id=self.test_task_instance_id,
            task_definition_name="test_task"
        )
        
        # Test status updates with actual API communication
        handler.update_status(tc, "Initializing")
        handler.update_status(tc, "Processing data")
        handler.update_status(tc, "Generating results")
        handler.update_status(tc, "Completed successfully")
        
        # All status updates should complete without exceptions
        assert True
    
    def test_remote_handler_stop_requests(self, integration_config):
        """Test stop request handling through RemoteTaskControlHandler."""
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        tc = TaskControl(
            instance_id=self.test_task_instance_id,
            task_definition_name="test_task"
        )
        
        # Initially should not be stopped
        assert not tc.should_stop
        
        # Request stop
        handler.request_stop(tc)
        
        # Should be marked as stopped
        assert tc.should_stop
    
    def test_remote_handler_with_task_execution(self, integration_config):
        """Test RemoteTaskControlHandler with actual task execution."""
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        
        # Set as global handler
        set_task_control_handler(handler)
        
        # Define test task
        @task
        def integration_test_task(tc: TaskControl):
            """Test task that uses TaskControl with remote handler."""
            tc.log("Starting integration test task", "info")
            tc.update_progress(10.0, "Initialized")
            
            tc.log("Processing data", "info", step=1)
            tc.update_progress(50.0, "Processing")
            
            if tc.should_stop:
                tc.log("Task was stopped", "warning")
                return "stopped"
            
            tc.log("Completing task", "info", step=2)
            tc.update_progress(100.0, "Completed")
            
            return "integration_test_complete"
        
        # Execute task
        result = integration_test_task()
        
        # Verify execution
        assert result == "integration_test_complete"
    
    def test_remote_handler_error_handling(self, integration_config):
        """Test error handling in RemoteTaskControlHandler."""
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        tc = TaskControl(
            instance_id=self.test_task_instance_id,
            task_definition_name="test_task"
        )
        
        # Test that handler gracefully handles various scenarios
        try:
            # These should not raise exceptions even if API calls fail
            handler.log(tc, "Test with complex data", "info", 
                       complex_data={"nested": {"dict": [1, 2, 3]}, "unicode": "ñoño"})
            handler.update_progress(tc, 42.5, "Complex status with symbols: ±∞αβγ")
            handler.update_status(tc, "Status with newlines\nand\ttabs")
            
        except Exception as e:
            pytest.fail(f"RemoteTaskControlHandler should handle errors gracefully: {e}")
    
    def test_remote_handler_concurrent_operations(self, integration_config):
        """Test RemoteTaskControlHandler with concurrent operations."""
        import threading
        
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        
        # Create multiple TaskControls
        task_controls = [
            TaskControl(
                instance_id=f"{self.test_task_instance_id}_{i}",
                task_definition_name=f"concurrent_task_{i}"
            )
            for i in range(3)
        ]
        
        results = []
        
        def worker(tc: TaskControl, worker_id: int):
            """Worker function for concurrent testing."""
            try:
                for step in range(5):
                    handler.log(tc, f"Worker {worker_id} step {step}", "info", 
                               worker_id=worker_id, step=step)
                    handler.update_progress(tc, step * 20.0, f"Step {step}")
                    time.sleep(0.01)  # Small delay to simulate work
                
                handler.update_status(tc, f"Worker {worker_id} completed")
                results.append(f"worker_{worker_id}_success")
                
            except Exception as e:
                results.append(f"worker_{worker_id}_error_{e}")
        
        # Start concurrent workers
        threads = []
        for i, tc in enumerate(task_controls):
            thread = threading.Thread(target=worker, args=(tc, i))
            threads.append(thread)
            thread.start()
        
        # Wait for all workers to complete
        for thread in threads:
            thread.join(timeout=10.0)  # 10 second timeout
        
        # Verify all workers completed successfully
        assert len(results) == 3
        for result in results:
            assert "success" in result
    
    def test_remote_handler_check_stop_requested(self, integration_config):
        """Test check_stop_requested functionality."""
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        tc = TaskControl(
            instance_id=self.test_task_instance_id,
            task_definition_name="test_task"
        )
        
        # Initially should return False
        assert not handler.check_stop_requested(tc)
        
        # Set stop flag locally
        tc._should_stop_internal = True
        
        # Should now return True
        assert handler.check_stop_requested(tc)
    
    def test_remote_handler_session_persistence(self, integration_config):
        """Test that session ID is maintained throughout handler lifecycle."""
        session_id = f"persistent_session_{int(time.time())}"
        
        handler = RemoteTaskControlHandler(session_id=session_id)
        
        # Create multiple task controls with the same handler
        task_controls = [
            TaskControl(
                instance_id=f"task_{i}",
                task_definition_name=f"persistent_task_{i}"
            )
            for i in range(3)
        ]
        
        # Perform operations with all task controls
        for i, tc in enumerate(task_controls):
            handler.log(tc, f"Task {i} operation", "info", task_index=i)
            handler.update_progress(tc, i * 33.33, f"Progress for task {i}")
        
        # Session ID should remain consistent
        assert handler.session_id == session_id 