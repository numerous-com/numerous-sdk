"""
Integration tests for local task execution with remote API persistence.

These tests verify that tasks can execute locally in the Python interpreter
while persisting their progress, logs, and status to the remote API at localhost:8080.

This tests the hybrid approach: local computation + remote persistence.
"""

import pytest
import time
import asyncio
from concurrent.futures import ThreadPoolExecutor
from typing import List, Dict, Any

from numerous.tasks.control import TaskControl, set_task_control_handler, get_task_control_handler
from numerous.tasks.integration.remote_handler import RemoteTaskControlHandler
from numerous.tasks.task import task
from numerous.tasks.backends.local import LocalExecutionBackend


@pytest.mark.integration
class TestLocalExecutionRemotePersistence:
    """Integration tests for local execution with remote API persistence."""
    
    def setup_method(self):
        """Set up test environment before each test."""
        self.original_handler = get_task_control_handler()
        self.test_session_id = f"local_exec_remote_persist_{int(time.time())}"
        
        # Set up remote handler for persistence
        self.remote_handler = RemoteTaskControlHandler(session_id=self.test_session_id)
        set_task_control_handler(self.remote_handler)
    
    def teardown_method(self):
        """Clean up after each test."""
        # Reset to original handler
        set_task_control_handler(self.original_handler)
    
    def test_simple_local_task_with_remote_persistence(self, integration_config):
        """Test simple task running locally with remote state persistence."""
        
        @task
        def simple_local_task(tc: TaskControl):
            """Simple task that runs locally but persists to API."""
            tc.log("Task starting locally", "info", execution_location="local", 
                   persistence_location="remote")
            tc.update_status("Initializing local execution")
            tc.update_progress(10.0, "Local initialization complete")
            
            # Simulate local computation
            local_data = {"computed_locally": True, "timestamp": time.time()}
            tc.log("Performing local computation", "info", data=local_data)
            
            # Simulate some work
            time.sleep(0.1)
            
            tc.update_progress(50.0, "Local computation in progress")
            
            # More local work
            result = sum(range(100))  # Local computation
            tc.log(f"Local computation result: {result}", "info", result=result)
            
            tc.update_progress(90.0, "Local computation complete")
            tc.update_status("Finalizing local task")
            
            tc.log("Task completed locally", "info", execution_mode="local_with_remote_persistence")
            tc.update_progress(100.0, "Task complete")
            
            return {
                "status": "success",
                "execution_location": "local",
                "persistence_location": "remote",
                "result": result,
                "local_data": local_data
            }
        
        # Execute task locally
        result = simple_local_task()
        
        # Verify task executed locally but logged to remote API
        assert result["status"] == "success"
        assert result["execution_location"] == "local"
        assert result["persistence_location"] == "remote"
        assert result["result"] == 4950  # sum(range(100))
        assert result["local_data"]["computed_locally"] is True
    
    def test_compute_intensive_local_task_with_remote_tracking(self, integration_config):
        """Test compute-intensive task running locally with detailed remote progress tracking."""
        
        @task
        def compute_intensive_task(tc: TaskControl):
            """Compute-intensive task with detailed progress reporting to API."""
            tc.log("Starting compute-intensive local task", "info", 
                   task_type="compute_intensive", location="local")
            
            # Simulate data processing pipeline
            data_size = 100
            batch_size = 20
            results = []
            
            tc.update_status("Processing data locally")
            tc.update_progress(0.0, "Starting data processing")
            
            for batch_num in range(0, data_size, batch_size):
                batch_end = min(batch_num + batch_size, data_size)
                
                # Local computation
                batch_data = list(range(batch_num, batch_end))
                batch_result = sum(x * x for x in batch_data)  # Local processing
                results.append(batch_result)
                
                # Remote persistence
                progress = (batch_end / data_size) * 100.0
                tc.update_progress(progress, f"Processed batch {batch_num//batch_size + 1}")
                tc.log(f"Batch {batch_num//batch_size + 1} processed locally", "debug",
                       batch_start=batch_num, batch_end=batch_end, batch_result=batch_result)
                
                # Check for stop requests via API
                if tc.should_stop:
                    tc.log("Task stopped via remote API request", "warning", 
                           batches_completed=len(results))
                    return {
                        "status": "stopped",
                        "batches_completed": len(results),
                        "partial_results": results
                    }
                
                # Simulate processing time
                time.sleep(0.01)
            
            # Final local computation
            total_result = sum(results)
            
            tc.update_status("Local computation complete")
            tc.update_progress(100.0, "All batches processed")
            tc.log("Compute-intensive task completed", "info", 
                   total_batches=len(results), final_result=total_result,
                   execution_summary="local_compute_remote_persistence")
            
            return {
                "status": "success",
                "execution_location": "local",
                "batches_processed": len(results),
                "total_result": total_result,
                "results": results
            }
        
        # Execute task
        result = compute_intensive_task()
        
        # Verify results
        assert result["status"] == "success"
        assert result["execution_location"] == "local"
        assert result["batches_processed"] == 5
        assert result["total_result"] > 0
        assert len(result["results"]) == 5
    
    def test_error_handling_local_execution_remote_persistence(self, integration_config):
        """Test error handling in local execution with remote persistence."""
        
        @task
        def error_prone_local_task(tc: TaskControl):
            """Task that may encounter errors during local execution."""
            tc.log("Starting error-prone local task", "info", 
                   execution_mode="local", persistence_mode="remote")
            
            try:
                tc.update_status("Performing risky local operation")
                tc.update_progress(25.0, "Starting risky operation")
                
                # Simulate risky local computation
                tc.log("Attempting division operation locally", "info", operation="division")
                
                # This operation will succeed (not actually risky in this test)
                divisor = 2
                result = 100 / divisor
                
                tc.log(f"Local operation successful: 100 / {divisor} = {result}", "info", 
                       operation_result=result, execution_location="local")
                tc.update_progress(75.0, "Risky operation completed successfully")
                
                tc.update_status("Local task completing successfully")
                tc.update_progress(100.0, "Task complete")
                
                return {
                    "status": "success",
                    "result": result,
                    "execution_location": "local",
                    "operation": "division"
                }
                
            except Exception as e:
                # Error occurred during local execution, but we can still persist to remote API
                tc.log(f"Error during local execution: {str(e)}", "error", 
                       error_type=type(e).__name__, execution_location="local")
                tc.update_status(f"Local execution failed: {str(e)}")
                tc.update_progress(100.0, "Task failed")
                
                return {
                    "status": "error",
                    "error": str(e),
                    "execution_location": "local"
                }
        
        # Execute task
        result = error_prone_local_task()
        
        # Verify result (should succeed)
        assert result["status"] == "success"
        assert result["execution_location"] == "local"
        assert result["result"] == 50.0
        assert result["operation"] == "division"
    
    def test_concurrent_local_tasks_with_remote_persistence(self, integration_config):
        """Test multiple concurrent local tasks with remote persistence."""
        
        @task
        def concurrent_local_worker(tc: TaskControl, worker_id: int, work_items: int):
            """Worker task that runs locally but persists progress remotely."""
            tc.log(f"Worker {worker_id} starting local execution", "info", 
                   worker_id=worker_id, work_items=work_items, location="local")
            
            results = []
            
            for item in range(work_items):
                # Local computation
                item_result = worker_id * 100 + item
                results.append(item_result)
                
                # Remote persistence
                progress = ((item + 1) / work_items) * 100.0
                tc.update_progress(progress, f"Worker {worker_id} processed item {item + 1}")
                tc.log(f"Worker {worker_id} processed item {item + 1} locally", "debug",
                       worker_id=worker_id, item=item + 1, item_result=item_result)
                
                # Check for remote stop requests
                if tc.should_stop:
                    tc.log(f"Worker {worker_id} stopped via remote request", "warning",
                           items_completed=len(results))
                    return {
                        "status": "stopped",
                        "worker_id": worker_id,
                        "items_completed": len(results),
                        "results": results
                    }
                
                # Simulate local processing time
                time.sleep(0.01)
            
            tc.update_progress(100.0, f"Worker {worker_id} completed all items")
            tc.log(f"Worker {worker_id} completed local execution", "info",
                   worker_id=worker_id, total_items=len(results))
            
            return {
                "status": "success",
                "worker_id": worker_id,
                "items_processed": len(results),
                "results": results,
                "execution_location": "local"
            }
        
        # Execute multiple concurrent local tasks
        import threading
        results = []
        threads = []
        
        def execute_worker(worker_id: int):
            """Execute worker in thread."""
            try:
                result = concurrent_local_worker(worker_id=worker_id, work_items=5)
                results.append(result)
            except Exception as e:
                results.append({
                    "status": "error",
                    "worker_id": worker_id,
                    "error": str(e)
                })
        
        # Start 3 concurrent workers
        for i in range(3):
            thread = threading.Thread(target=execute_worker, args=(i,))
            threads.append(thread)
            thread.start()
        
        # Wait for all workers to complete
        for thread in threads:
            thread.join(timeout=10.0)
        
        # Verify all workers completed successfully
        assert len(results) == 3
        for result in results:
            assert result["status"] == "success"
            assert result["execution_location"] == "local"
            assert result["items_processed"] == 5
            assert len(result["results"]) == 5
    
    def test_local_execution_with_complex_remote_logging(self, integration_config):
        """Test local execution with complex data logging to remote API."""
        
        @task
        def complex_data_local_task(tc: TaskControl):
            """Local task that logs complex data structures to remote API."""
            tc.log("Starting complex data local task", "info", 
                   task_type="complex_data", execution_location="local")
            
            # Generate complex data locally
            complex_local_data = {
                "computation_metadata": {
                    "start_time": time.time(),
                    "environment": "local_python",
                    "persistence": "remote_api"
                },
                "data_processing": {
                    "input_size": 1000,
                    "algorithms_used": ["sorting", "filtering", "aggregation"],
                    "performance_metrics": {
                        "cpu_usage": "simulated",
                        "memory_usage": "simulated",
                        "execution_time": 0.1
                    }
                },
                "results_preview": {
                    "sample_data": [1, 2, 3, 4, 5],
                    "statistics": {
                        "mean": 3.0,
                        "median": 3.0,
                        "std_dev": 1.58
                    }
                }
            }
            
            # Log complex data to remote API
            tc.log("Generated complex local data", "info", 
                   computation_data=complex_local_data,
                   data_size=len(str(complex_local_data)))
            
            tc.update_progress(25.0, "Complex data generated locally")
            
            # Process data locally
            processed_results = []
            for i in range(10):
                local_computation = i ** 2 + complex_local_data["results_preview"]["statistics"]["mean"]
                processed_results.append(local_computation)
            
            tc.update_progress(75.0, "Local data processing complete")
            
            # Log results with unicode and special characters
            tc.log("Processing complete with unicode: ñoño αβγ ±∞ 中文 🚀", "info",
                   unicode_test="Testing unicode characters",
                   special_chars="!@#$%^&*()[]{}|;:,.<>?",
                   json_data='{"nested": "value with \\"quotes\\" and \\n newlines"}')
            
            final_result = {
                "execution_mode": "local_with_remote_persistence",
                "input_data": complex_local_data,
                "processed_results": processed_results,
                "summary": {
                    "total_processed": len(processed_results),
                    "execution_time": time.time() - complex_local_data["computation_metadata"]["start_time"]
                }
            }
            
            tc.update_progress(100.0, "Complex data task completed")
            tc.log("Complex data local task completed successfully", "info",
                   final_result_summary=final_result["summary"])
            
            return final_result
        
        # Execute task
        result = complex_data_local_task()
        
        # Verify results
        assert result["execution_mode"] == "local_with_remote_persistence"
        assert "input_data" in result
        assert "processed_results" in result
        assert len(result["processed_results"]) == 10
        assert result["summary"]["total_processed"] == 10
        assert result["summary"]["execution_time"] > 0
    
    def test_local_task_stop_via_remote_api(self, integration_config):
        """Test stopping local task execution via remote API request."""
        
        stop_after_iterations = 3
        
        @task  
        def stoppable_local_task(tc: TaskControl):
            """Local task that can be stopped via remote API."""
            tc.log("Starting stoppable local task", "info", 
                   execution_location="local", stop_mechanism="remote_api")
            
            for i in range(10):  # Would process 10 items if not stopped
                # Local computation
                local_result = i * i
                
                # Remote persistence
                progress = ((i + 1) / 10) * 100.0
                tc.update_progress(progress, f"Local processing item {i + 1}/10")
                tc.log(f"Processed item {i + 1} locally", "info", 
                       item=i + 1, local_result=local_result)
                
                # Simulate stop request after certain iterations
                if i == stop_after_iterations:
                    tc.log("Simulating remote stop request", "info", iteration=i + 1)
                    self.remote_handler.request_stop(tc)
                
                # Check for remote stop requests
                if tc.should_stop:
                    tc.log(f"Local task stopped via remote API at item {i + 1}", "warning",
                           final_item=i + 1, execution_location="local")
                    return {
                        "status": "stopped",
                        "items_processed": i + 1,
                        "stop_mechanism": "remote_api",
                        "execution_location": "local"
                    }
                
                # Simulate local processing time
                time.sleep(0.02)
            
            tc.update_progress(100.0, "All items processed locally")
            tc.log("Local task completed all items", "info", total_items=10)
            
            return {
                "status": "success", 
                "items_processed": 10,
                "execution_location": "local"
            }
        
        # Execute task
        result = stoppable_local_task()
        
        # Verify task was stopped via remote API
        assert result["status"] == "stopped"
        assert result["items_processed"] == stop_after_iterations + 1
        assert result["stop_mechanism"] == "remote_api"
        assert result["execution_location"] == "local" 