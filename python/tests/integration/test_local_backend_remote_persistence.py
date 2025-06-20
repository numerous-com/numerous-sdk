"""
Integration tests for LocalExecutionBackend with RemoteTaskControlHandler.

These tests verify that TaskInstances can execute locally using LocalExecutionBackend
while persisting their state to the remote API via RemoteTaskControlHandler.

This tests the architecture where:
- Task execution happens in local Python interpreter via LocalExecutionBackend
- TaskControl operations are handled by RemoteTaskControlHandler for API persistence
"""

import pytest
import time
from concurrent.futures import Future

from numerous.tasks.control import set_task_control_handler, get_task_control_handler
from numerous.tasks.integration.remote_handler import RemoteTaskControlHandler
from numerous.tasks.task import task
from numerous.tasks.backends.local import LocalExecutionBackend
from numerous.tasks.session import Session


@pytest.mark.integration
class TestLocalBackendRemotePersistence:
    """Integration tests for LocalExecutionBackend with RemoteTaskControlHandler."""
    
    def setup_method(self):
        """Set up test environment before each test."""
        self.original_handler = get_task_control_handler()
        self.test_session_id = f"local_backend_remote_persist_{int(time.time())}"
        
        # Set up remote handler for persistence
        self.remote_handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        set_task_control_handler(self.remote_handler)
        
        # Set up local execution backend
        self.local_backend = LocalExecutionBackend()
    
    def teardown_method(self):
        """Clean up after each test."""
        # Reset to original handler
        set_task_control_handler(self.original_handler)
    
    def test_task_instance_local_execution_remote_persistence(self, integration_config):
        """Test TaskInstance execution with LocalExecutionBackend and RemoteTaskControlHandler."""
        
        @task
        def local_task_with_remote_persistence(tc):
            """Task that executes locally but persists remotely."""
            tc.log("TaskInstance starting execution locally", "info", 
                   backend="LocalExecutionBackend", persistence="RemoteTaskControlHandler")
            
            tc.update_status("Executing in local Python interpreter")
            tc.update_progress(25.0, "Local execution started")
            
            # Local computation
            computation_result = sum(i * i for i in range(50))
            tc.log(f"Local computation completed: {computation_result}", "info", 
                   computation_type="sum_of_squares", result=computation_result)
            
            tc.update_progress(75.0, "Local computation complete")
            
            # Simulate additional local work
            time.sleep(0.05)
            
            tc.update_status("Local execution finishing")
            tc.update_progress(100.0, "TaskInstance execution complete")
            tc.log("TaskInstance completed successfully", "info", 
                   execution_mode="local_backend_remote_persistence")
            
            return {
                "status": "success",
                "backend": "LocalExecutionBackend", 
                "persistence": "RemoteTaskControlHandler",
                "result": computation_result
            }
        
        # Create and execute TaskInstance locally
        task_instance = local_task_with_remote_persistence.instance()
        future = task_instance.start(backend=self.local_backend)
        
        # Wait for completion
        result = future.result(timeout=10.0)
        
        # Verify execution
        assert result["status"] == "success"
        assert result["backend"] == "LocalExecutionBackend"
        assert result["persistence"] == "RemoteTaskControlHandler"
        assert result["result"] == sum(i * i for i in range(50))
    
    def test_multiple_task_instances_local_execution_remote_persistence(self, integration_config):
        """Test multiple TaskInstances executing locally with remote persistence."""
        
        @task
        def multi_instance_task(tc, instance_id: int, work_amount: int):
            """Task that can be instantiated multiple times."""
            tc.log(f"Instance {instance_id} starting local execution", "info",
                   instance_id=instance_id, work_amount=work_amount)
            
            tc.update_status(f"Instance {instance_id} processing")
            
            results = []
            for i in range(work_amount):
                # Local computation
                item_result = instance_id * 100 + i
                results.append(item_result)
                
                # Remote persistence
                progress = ((i + 1) / work_amount) * 100.0
                tc.update_progress(progress, f"Instance {instance_id} item {i + 1}/{work_amount}")
                tc.log(f"Instance {instance_id} processed item {i + 1}", "debug",
                       instance_id=instance_id, item=i + 1, item_result=item_result)
                
                time.sleep(0.01)  # Simulate work
            
            tc.update_progress(100.0, f"Instance {instance_id} completed")
            tc.log(f"Instance {instance_id} completed local execution", "info",
                   instance_id=instance_id, total_results=len(results))
            
            return {
                "status": "success",
                "instance_id": instance_id,
                "results": results,
                "execution_mode": "local_backend_remote_persistence"
            }
        
        # Create multiple task instances
        instances = []
        futures = []
        
        for i in range(3):
            instance = multi_instance_task.instance(instance_id=i, work_amount=5)
            future = instance.start(backend=self.local_backend)
            instances.append(instance)
            futures.append(future)
        
        # Wait for all instances to complete
        results = []
        for future in futures:
            result = future.result(timeout=10.0)
            results.append(result)
        
        # Verify all instances completed successfully
        assert len(results) == 3
        for i, result in enumerate(results):
            assert result["status"] == "success"
            assert result["instance_id"] == i
            assert result["execution_mode"] == "local_backend_remote_persistence"
            assert len(result["results"]) == 5
    
    def test_session_managed_local_execution_remote_persistence(self, integration_config):
        """Test session-managed TaskInstance execution with local backend and remote persistence."""
        
        @task
        def session_managed_task(tc):
            """Task executed within a session context."""
            tc.log("Session-managed task starting", "info", 
                   session_context="managed", execution="local", persistence="remote")
            
            tc.update_status("Session task executing locally")
            tc.update_progress(50.0, "Session task in progress")
            
            # Simulate session-aware work
            session_data = {"session_id": "test_session", "managed": True}
            tc.log("Processing within session context", "info", session_data=session_data)
            
            # Local computation
            result = len(str(session_data)) * 10
            
            tc.update_progress(100.0, "Session task completed")
            tc.log("Session-managed task completed", "info", 
                   session_result=result, context="session_managed")
            
            return {
                "status": "success",
                "session_managed": True,
                "result": result,
                "execution_details": {
                    "backend": "LocalExecutionBackend",
                    "persistence": "RemoteTaskControlHandler",
                    "session_context": "managed"
                }
            }
        
        # Execute within session context
        with Session() as session:
            task_instance = session_managed_task.instance()
            future = task_instance.start(backend=self.local_backend)
            result = future.result(timeout=10.0)
        
        # Verify session-managed execution
        assert result["status"] == "success"
        assert result["session_managed"] is True
        assert result["execution_details"]["backend"] == "LocalExecutionBackend"
        assert result["execution_details"]["persistence"] == "RemoteTaskControlHandler"
        assert result["execution_details"]["session_context"] == "managed"
    
    def test_local_backend_task_cancellation_via_remote_api(self, integration_config):
        """Test cancelling locally executing TaskInstance via remote API."""
        
        @task
        def cancellable_local_task(tc):
            """Task that can be cancelled during local execution."""
            tc.log("Cancellable task starting local execution", "info",
                   cancellation_method="remote_api")
            
            for i in range(10):
                tc.update_progress(i * 10.0, f"Processing step {i + 1}/10")
                tc.log(f"Executing step {i + 1} locally", "info", step=i + 1)
                
                # Check for remote cancellation
                if tc.should_stop:
                    tc.log(f"Task cancelled via remote API at step {i + 1}", "warning",
                           cancelled_at_step=i + 1)
                    return {
                        "status": "cancelled",
                        "steps_completed": i + 1,
                        "cancellation_method": "remote_api"
                    }
                
                time.sleep(0.05)  # Simulate work
            
            tc.update_progress(100.0, "All steps completed")
            return {"status": "success", "steps_completed": 10}
        
        # Start task instance
        task_instance = cancellable_local_task.instance()
        future = task_instance.start(backend=self.local_backend)
        
        # Cancel after short delay (simulate remote API cancellation)
        time.sleep(0.15)  # Let a few steps execute
        task_instance.cancel()
        
        # Wait for completion
        result = future.result(timeout=10.0)
        
        # Verify cancellation
        assert result["status"] == "cancelled"
        assert result["steps_completed"] < 10  # Should be cancelled before completion
        assert result["cancellation_method"] == "remote_api"
    
    def test_local_backend_error_handling_with_remote_persistence(self, integration_config):
        """Test error handling in LocalExecutionBackend with RemoteTaskControlHandler."""
        
        @task
        def error_prone_local_task(tc):
            """Task that encounters an error during local execution."""
            tc.log("Error-prone task starting", "info", 
                   error_handling="local_backend_remote_persistence")
            
            try:
                tc.update_status("Performing potentially failing operation")
                tc.update_progress(25.0, "Starting risky operation")
                
                # This will actually fail
                result = 10 / 0  # Division by zero
                
                return {"status": "success", "result": result}
                
            except Exception as e:
                # Error is caught and logged remotely
                tc.log(f"Error occurred during local execution: {str(e)}", "error",
                       error_type=type(e).__name__, execution_location="local")
                tc.update_status(f"Task failed: {str(e)}")
                tc.update_progress(100.0, "Task failed")
                
                return {
                    "status": "error",
                    "error": str(e),
                    "error_type": type(e).__name__,
                    "execution_location": "local"
                }
        
        # Execute task instance
        task_instance = error_prone_local_task.instance()
        future = task_instance.start(backend=self.local_backend)
        result = future.result(timeout=10.0)
        
        # Verify error handling
        assert result["status"] == "error"
        assert result["error_type"] == "ZeroDivisionError"
        assert result["execution_location"] == "local"
        assert "division by zero" in result["error"].lower()
    
    def test_local_backend_performance_with_remote_logging(self, integration_config):
        """Test performance of LocalExecutionBackend with frequent remote logging."""
        
        @task
        def performance_test_task(tc):
            """Task that performs many operations with frequent remote logging."""
            tc.log("Performance test starting", "info", 
                   test_type="frequent_remote_logging", operations_count=100)
            
            start_time = time.time()
            results = []
            
            for i in range(100):  # Many operations
                # Local computation
                local_result = i ** 2
                results.append(local_result)
                
                # Frequent remote logging (every 10 operations)
                if i % 10 == 0:
                    progress = (i / 100) * 100.0
                    tc.update_progress(progress, f"Completed {i}/100 operations")
                    tc.log(f"Performance checkpoint: {i} operations completed", "debug",
                           operations_completed=i, current_result=local_result)
                
                # Brief pause to simulate work
                time.sleep(0.001)
            
            execution_time = time.time() - start_time
            
            tc.update_progress(100.0, "Performance test completed")
            tc.log("Performance test completed", "info",
                   total_operations=len(results), execution_time=execution_time,
                   operations_per_second=len(results) / execution_time)
            
            return {
                "status": "success",
                "operations_completed": len(results),
                "execution_time": execution_time,
                "operations_per_second": len(results) / execution_time,
                "backend_performance": "local_with_remote_logging"
            }
        
        # Execute performance test
        task_instance = performance_test_task.instance()
        future = task_instance.start(backend=self.local_backend)
        result = future.result(timeout=30.0)  # Longer timeout for performance test
        
        # Verify performance test results
        assert result["status"] == "success"
        assert result["operations_completed"] == 100
        assert result["execution_time"] > 0
        assert result["operations_per_second"] > 0
        assert result["backend_performance"] == "local_with_remote_logging" 