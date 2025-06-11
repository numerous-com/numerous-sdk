"""
Tests for local backend task execution (developer mode).
"""

import pytest
import time
import concurrent.futures
from typing import List, Dict, Any

from numerous.tasks import task, TaskControl, Session


# Test task definitions
@task
def add_numbers(a: int, b: int) -> int:
    """Simple arithmetic task."""
    return a + b

@task
def process_list(items: List[str], prefix: str = "processed") -> List[str]:
    """Process a list of items."""
    return [f"{prefix}_{item}" for item in items]

@task
def failing_task(should_fail: bool = True) -> str:
    """Task that can be configured to fail."""
    if should_fail:
        raise ValueError("Task failed intentionally")
    return "success"

@task
def progress_task(tc: TaskControl, steps: int, delay: float = 0.01) -> str:
    """Task that reports progress."""
    tc.log(f"Starting progress task with {steps} steps", level="info")
    
    for i in range(steps):
        if tc.should_stop:
            tc.log(f"Task stopped at step {i}", level="warning")
            return f"stopped_at_step_{i}"
        
        progress = (i + 1) / steps * 100
        tc.update_progress(progress, f"Step {i+1}/{steps}")
        
        if delay > 0:
            time.sleep(delay)
    
    tc.log("Task completed successfully", level="info")
    return f"completed_{steps}_steps"

@task
def concurrent_worker(worker_id: int, duration: float = 0.05) -> Dict[str, Any]:
    """Worker task for concurrent testing."""
    start_time = time.time()
    time.sleep(duration)
    end_time = time.time()
    
    return {
        "worker_id": worker_id,
        "duration": end_time - start_time,
        "start_time": start_time,
        "end_time": end_time
    }


class TestLocalBackendBasics:
    """Test basic local backend functionality."""
    
    def test_simple_task_execution(self):
        """Test simple task execution."""
        with Session() as session:
            result = add_numbers(10, 5)
            assert result == 15
    
    def test_list_processing(self):
        """Test list processing task."""
        with Session() as session:
            items = ["apple", "banana", "cherry"]
            result = process_list(items)
            expected = ["processed_apple", "processed_banana", "processed_cherry"]
            assert result == expected
            
            # Test with custom prefix
            result = process_list(items, "custom")
            expected = ["custom_apple", "custom_banana", "custom_cherry"]
            assert result == expected
    
    def test_error_handling(self):
        """Test error handling."""
        with Session() as session:
            # Test successful execution
            result = failing_task(False)
            assert result == "success"
            
            # Test error case
            with pytest.raises(ValueError, match="Task failed intentionally"):
                failing_task(True)
    
    def test_session_management(self):
        """Test session context management."""
        with Session(name="test_session") as session:
            assert session.name == "test_session"
            assert session._is_active
            assert Session.current() == session
            
            # Execute task in session
            result = add_numbers(1, 2)
            assert result == 3
        
        # Session should be inactive after context
        assert not session._is_active
        assert Session.current() is None


class TestTaskControlFeatures:
    """Test TaskControl functionality."""
    
    def test_progress_reporting(self):
        """Test progress reporting."""
        with Session() as session:
            result = progress_task(5, delay=0.001)
            assert result == "completed_5_steps"
    
    def test_task_cancellation(self):
        """Test task cancellation."""
        with Session() as session:
            # Start a longer task
            instance = progress_task.instance()
            future = instance.start(20, delay=0.01)
            
            # Let it run briefly then cancel
            time.sleep(0.05)
            instance.stop()
            
            result = future.result()
            assert result.startswith("stopped_at_step_")
            
            # Verify it stopped early
            step_num = int(result.split("_")[-1])
            assert step_num < 20
    
    def test_multiple_tasks_in_session(self):
        """Test running multiple tasks in same session."""
        with Session() as session:
            # Execute multiple tasks
            result1 = add_numbers(10, 20)
            result2 = process_list(["x", "y"], "test")
            result3 = progress_task(3, delay=0.001)
            
            assert result1 == 30
            assert result2 == ["test_x", "test_y"]
            assert result3 == "completed_3_steps"


class TestConcurrentExecution:
    """Test concurrent task execution."""
    
    def test_concurrent_simple_tasks(self):
        """Test running simple tasks concurrently."""
        with concurrent.futures.ThreadPoolExecutor(max_workers=4) as executor:
            futures = []
            for i in range(5):
                def run_task(task_id):
                    with Session() as session:
                        return concurrent_worker(task_id, 0.02)
                
                future = executor.submit(run_task, i)
                futures.append(future)
            
            results = [f.result() for f in futures]
            
            # Verify all tasks completed
            assert len(results) == 5
            for i, result in enumerate(results):
                assert result["worker_id"] == i
                assert result["duration"] >= 0.02
    
    def test_mixed_concurrent_tasks(self):
        """Test running different task types concurrently."""
        with concurrent.futures.ThreadPoolExecutor(max_workers=3) as executor:
            # Submit different types of tasks
            def run_simple():
                with Session() as session:
                    return add_numbers(100, 200)
            future1 = executor.submit(run_simple)
            
            def run_progress():
                with Session() as session:
                    return progress_task(3, delay=0.01)
            future2 = executor.submit(run_progress)
            
            def run_list_processing():
                with Session() as session:
                    return process_list(["a", "b", "c"])
            future3 = executor.submit(run_list_processing)
            
            # Wait for all to complete
            result1 = future1.result()
            result2 = future2.result()
            result3 = future3.result()
            
            # Verify results
            assert result1 == 300
            assert result2 == "completed_3_steps"
            assert result3 == ["processed_a", "processed_b", "processed_c"]


class TestPerformance:
    """Test performance characteristics."""
    
    def test_task_startup_speed(self):
        """Test that tasks start up quickly."""
        with Session() as session:
            start_time = time.time()
            
            for i in range(10):
                result = add_numbers(i, i + 1)
                assert result == i + i + 1
            
            end_time = time.time()
            total_time = end_time - start_time
            
            # Should complete 10 simple tasks quickly
            assert total_time < 0.5  # Half a second should be plenty 