"""
End-to-end integration tests for task execution with API persistence.

These tests require the --integration flag and assume API availability at localhost:8080.
They test complete task execution flows with remote logging and progress tracking.
"""

import pytest
import time
import json
from typing import List, Dict, Any, Optional
from unittest.mock import patch

from numerous.tasks.control import TaskControl, set_task_control_handler, get_task_control_handler
from numerous.tasks.integration.remote_handler import RemoteTaskControlHandler
from numerous.tasks.task import task


@pytest.mark.integration
class TestTaskExecutionIntegration:
    """Integration tests for complete task execution with API persistence."""
    
    def setup_method(self):
        """Set up test environment before each test."""
        self.original_handler = get_task_control_handler()
        self.test_session_id = f"integration_session_{int(time.time())}"
    
    def teardown_method(self):
        """Clean up after each test."""
        # Reset to original handler
        set_task_control_handler(self.original_handler)
    
    def test_simple_task_execution_with_remote_handler(self, integration_config):
        """Test simple task execution with RemoteTaskControlHandler."""
        # Set up remote handler - will fail at initialization if API not available
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        set_task_control_handler(handler)
        
        @task
        def simple_integration_task(tc: TaskControl):
            """Simple task for integration testing."""
            tc.log("Task started", "info", task_type="simple")
            tc.update_status("Processing")
            tc.update_progress(50.0, "Halfway done")
            
            # Simulate some work
            time.sleep(0.1)
            
            tc.update_progress(100.0, "Completed")
            tc.log("Task completed successfully", "info")
            
            return {"status": "success", "result": "simple_task_complete"}
        
        # Execute task
        result = simple_integration_task()
        
        # Verify result
        assert result["status"] == "success"
        assert result["result"] == "simple_task_complete"
    
    def test_data_processing_task_with_remote_handler(self, integration_config):
        """Test data processing task with detailed logging and progress updates."""
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        set_task_control_handler(handler)
        
        @task
        def data_processing_task(tc: TaskControl):
            """Data processing task for integration testing."""
            tc.log("Starting data processing task", "info", 
                   task_id="data_proc_001", version="1.0")
            
            # Simulate data loading
            tc.update_status("Loading data")
            tc.update_progress(10.0, "Loading input data")
            tc.log("Loading data from source", "info", phase="data_loading")
            
            # Simulate processing stages
            data_size = 100
            batch_size = 20
            
            for i in range(0, data_size, batch_size):
                batch_end = min(i + batch_size, data_size)
                progress = ((batch_end) / data_size) * 90.0 + 10.0  # 10% to 100%
                
                tc.update_progress(progress, f"Processing batch {i//batch_size + 1}")
                tc.log(f"Processing records {i} to {batch_end}", "debug", 
                       batch_id=i//batch_size + 1, records_processed=batch_end)
                
                # Check for stop requests
                if tc.should_stop:
                    tc.log("Task stopped by user request", "warning")
                    return {"status": "stopped", "processed": batch_end}
                
                time.sleep(0.01)  # Simulate processing time
            
            tc.update_status("Finalizing results")
            tc.update_progress(100.0, "Processing complete")
            tc.log("Data processing completed successfully", "info", 
                   total_records=data_size, processing_time=0.1)
            
            return {
                "status": "success",
                "records_processed": data_size,
                "batches_completed": data_size // batch_size
            }
        
        # Execute task
        result = data_processing_task()
        
        # Verify result
        assert result["status"] == "success"
        assert result["records_processed"] == 100
        assert result["batches_completed"] == 5
    
    def test_error_handling_task_with_remote_handler(self, integration_config):
        """Test error handling in tasks with RemoteTaskControlHandler."""
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        set_task_control_handler(handler)
        
        @task
        def error_prone_task(tc: TaskControl):
            """Task that may encounter errors for testing error handling."""
            tc.log("Starting error-prone task", "info", task_type="error_test")
            
            try:
                tc.update_status("Performing risky operation")
                tc.update_progress(25.0, "Starting risky operation")
                
                # Simulate work that might fail
                tc.log("Attempting risky operation", "info", operation="division")
                
                # This will succeed (not actually risky in this test)
                result = 100 / 2
                
                tc.log(f"Risky operation succeeded: {result}", "info", result=result)
                tc.update_progress(75.0, "Risky operation completed")
                
                tc.update_status("Finalizing")
                tc.update_progress(100.0, "Completed successfully")
                
                return {"status": "success", "result": result}
                
            except Exception as e:
                tc.log(f"Error in task execution: {str(e)}", "error", 
                       error_type=type(e).__name__, operation="division")
                tc.update_status(f"Failed: {str(e)}")
                return {"status": "error", "error": str(e)}
        
        # Execute task
        result = error_prone_task()
        
        # Verify result (should succeed)
        assert result["status"] == "success"
        assert result["result"] == 50.0
    
    def test_multi_step_task_with_remote_handler(self, integration_config):
        """Test multi-step task with detailed progress tracking."""
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        set_task_control_handler(handler)
        
        @task
        def multi_step_task(tc: TaskControl):
            """Multi-step task for comprehensive integration testing."""
            steps = [
                ("initialize", "Initializing task environment"),
                ("validate", "Validating input parameters"),
                ("prepare", "Preparing data structures"),
                ("execute", "Executing main logic"),
                ("verify", "Verifying results"),
                ("cleanup", "Cleaning up resources"),
                ("finalize", "Finalizing task")
            ]
            
            tc.log("Starting multi-step integration task", "info", 
                   total_steps=len(steps), task_id="multi_step_001")
            
            results = {}
            
            for i, (step_name, step_description) in enumerate(steps):
                progress = (i / len(steps)) * 100.0
                
                tc.update_status(f"Step {i+1}/{len(steps)}: {step_name}")
                tc.update_progress(progress, step_description)
                
                tc.log(f"Starting step: {step_name}", "info", 
                       step_id=i+1, step_name=step_name)
                
                # Simulate step execution
                time.sleep(0.02)
                
                # Simulate step-specific work
                if step_name == "validate":
                    tc.log("Validation passed", "info", validation_result="passed")
                    results[step_name] = "validated"
                elif step_name == "execute":
                    tc.log("Main logic executed", "info", execution_result="success")
                    results[step_name] = "executed"
                elif step_name == "verify":
                    tc.log("Results verified", "info", verification_result="passed")
                    results[step_name] = "verified"
                else:
                    results[step_name] = "completed"
                
                # Check for stop requests
                if tc.should_stop:
                    tc.log(f"Task stopped during step: {step_name}", "warning", 
                           completed_steps=i+1)
                    return {"status": "stopped", "completed_steps": i+1, "results": results}
                
                tc.log(f"Completed step: {step_name}", "info", 
                       step_id=i+1, step_result=results[step_name])
            
            # Final progress update
            tc.update_progress(100.0, "All steps completed")
            tc.log("Multi-step task completed successfully", "info", 
                   total_steps_completed=len(steps))
            
            return {
                "status": "success",
                "steps_completed": len(steps),
                "results": results
            }
        
        # Execute task
        result = multi_step_task()
        
        # Verify result
        assert result["status"] == "success"
        assert result["steps_completed"] == 7
        assert "validate" in result["results"]
        assert "execute" in result["results"]
        assert "verify" in result["results"]
    
    def test_concurrent_task_execution_with_remote_handler(self, integration_config):
        """Test concurrent task execution with RemoteTaskControlHandler."""
        import threading
        
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        set_task_control_handler(handler)
        
        @task
        def concurrent_worker_task(tc: TaskControl, worker_id: int, work_units: int):
            """Worker task for concurrent execution testing."""
            tc.log(f"Worker {worker_id} starting", "info", 
                   worker_id=worker_id, work_units=work_units)
            
            results = []
            
            for unit in range(work_units):
                progress = (unit / work_units) * 100.0
                tc.update_progress(progress, f"Processing unit {unit+1}/{work_units}")
                
                tc.log(f"Worker {worker_id} processing unit {unit+1}", "debug", 
                       worker_id=worker_id, unit=unit+1)
                
                # Simulate work
                time.sleep(0.01)
                results.append(f"unit_{unit+1}_result")
                
                # Check for stop requests
                if tc.should_stop:
                    tc.log(f"Worker {worker_id} stopped", "warning", 
                           units_completed=unit+1)
                    return {"status": "stopped", "results": results}
            
            tc.update_progress(100.0, "Worker completed")
            tc.log(f"Worker {worker_id} completed successfully", "info", 
                   units_processed=work_units)
            
            return {"status": "success", "worker_id": worker_id, "results": results}
        
        # Execute multiple concurrent tasks
        results = []
        threads = []
        
        def execute_worker(worker_id: int):
            """Execute worker task in thread."""
            try:
                result = concurrent_worker_task(worker_id=worker_id, work_units=5)
                results.append(result)
            except Exception as e:
                results.append({"status": "error", "worker_id": worker_id, "error": str(e)})
        
        # Start 3 concurrent workers
        for i in range(3):
            thread = threading.Thread(target=execute_worker, args=(i,))
            threads.append(thread)
            thread.start()
        
        # Wait for all workers to complete
        for thread in threads:
            thread.join(timeout=10.0)
        
        # Verify results
        assert len(results) == 3
        for result in results:
            assert result["status"] == "success"
            assert len(result["results"]) == 5
    
    def test_task_stop_request_handling(self, integration_config):
        """Test stop request handling during task execution."""
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        set_task_control_handler(handler)
        
        stop_after_iterations = 3
        
        @task
        def stoppable_task(tc: TaskControl):
            """Task that can be stopped during execution."""
            tc.log("Starting stoppable task", "info", task_type="stoppable")
            
            for i in range(10):  # Would run 10 iterations if not stopped
                progress = (i / 10) * 100.0
                tc.update_progress(progress, f"Iteration {i+1}/10")
                tc.log(f"Processing iteration {i+1}", "info", iteration=i+1)
                
                # Simulate work
                time.sleep(0.01)
                
                # Simulate stop request after certain iterations
                if i == stop_after_iterations:
                    tc.log("Simulating stop request", "info", iteration=i+1)
                    handler.request_stop(tc)
                
                # Check for stop requests
                if tc.should_stop:
                    tc.log(f"Task stopped at iteration {i+1}", "warning", 
                           final_iteration=i+1)
                    return {"status": "stopped", "iterations_completed": i+1}
            
            tc.update_progress(100.0, "All iterations completed")
            tc.log("Task completed all iterations", "info", total_iterations=10)
            
            return {"status": "success", "iterations_completed": 10}
        
        # Execute task
        result = stoppable_task()
        
        # Verify result (should be stopped)
        assert result["status"] == "stopped"
        assert result["iterations_completed"] == stop_after_iterations + 1  # +1 because we stop after the iteration
    
    def test_task_with_complex_data_logging(self, integration_config):
        """Test task execution with complex data structures in logging."""
        handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        set_task_control_handler(handler)
        
        @task
        def complex_data_task(tc: TaskControl):
            """Task that logs complex data structures."""
            tc.log("Starting complex data task", "info", task_type="complex_data")
            
            # Log various complex data types
            complex_data = {
                "nested_dict": {
                    "numbers": [1, 2, 3, 4, 5],
                    "strings": ["hello", "world", "integration", "test"],
                    "boolean": True,
                    "null_value": None
                },
                "metadata": {
                    "timestamp": time.time(),
                    "version": "1.0.0",
                    "environment": "integration_test"
                }
            }
            
            tc.log("Processing complex data structure", "info", 
                   data_structure=complex_data,
                   data_size=len(str(complex_data)))
            
            tc.update_progress(25.0, "Complex data logged")
            
            # Log with unicode characters
            tc.log("Processing unicode data", "info", 
                   unicode_text="Testing ñoño αβγ ±∞ 中文 🚀",
                   special_chars="!@#$%^&*()[]{}|;:,.<>?")
            
            tc.update_progress(50.0, "Unicode data processed")
            
            # Log with escaped characters
            tc.log("Processing escaped characters", "info", 
                   json_string='{"key": "value with \\"quotes\\" and \\n newlines"}',
                   regex_pattern=r"\\d+\\.\\d+")
            
            tc.update_progress(75.0, "Escaped characters processed")
            
            # Final result
            result = {
                "processed_data": complex_data,
                "stats": {
                    "total_numbers": len(complex_data["nested_dict"]["numbers"]),
                    "total_strings": len(complex_data["nested_dict"]["strings"])
                }
            }
            
            tc.update_progress(100.0, "Complex data task completed")
            tc.log("Complex data task completed successfully", "info", 
                   final_result=result)
            
            return result
        
        # Execute task
        result = complex_data_task()
        
        # Verify result
        assert "processed_data" in result
        assert "stats" in result
        assert result["stats"]["total_numbers"] == 5
        assert result["stats"]["total_strings"] == 4
    
 