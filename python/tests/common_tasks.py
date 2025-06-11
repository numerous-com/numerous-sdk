"""
Common task definitions for testing different execution modes.

These tasks are designed to test various aspects of the task system:
- Simple tasks (no TaskControl)
- TaskControl-enabled tasks  
- Error handling
- Progress tracking
- Logging
- Cancellation
- Different parameter types
"""

from numerous.tasks import task, TaskControl
import time
import json
from typing import List, Dict, Any, Optional


# --- Simple Tasks (No TaskControl) ---

@task
def add_numbers(a: int, b: int) -> int:
    """Simple arithmetic task for basic functionality testing."""
    return a + b

@task
def multiply_numbers(a: float, b: float) -> float:
    """Simple multiplication with float inputs."""
    return a * b

@task
def process_list(items: List[str], prefix: str = "processed") -> List[str]:
    """Process a list of items with optional prefix."""
    return [f"{prefix}_{item}" for item in items]

@task
def json_task(data: Dict[str, Any]) -> Dict[str, Any]:
    """Task that processes JSON-like data structures."""
    result = data.copy()
    result["processed"] = True
    result["item_count"] = len(data.get("items", []))
    return result

@task
def failing_task(should_fail: bool = True, error_message: str = "Task failed intentionally") -> str:
    """Task that can be configured to fail for error testing."""
    if should_fail:
        raise ValueError(error_message)
    return "success"


# --- TaskControl-Enabled Tasks ---

@task
def progress_task(tc: TaskControl, steps: int, delay: float = 0.01) -> str:
    """Task that reports progress through multiple steps."""
    tc.log(f"Starting progress task with {steps} steps", level="info")
    
    for i in range(steps):
        if tc.should_stop:
            tc.log(f"Task stopped at step {i}", level="warning")
            return f"stopped_at_step_{i}"
        
        progress = (i + 1) / steps * 100
        tc.update_progress(progress, f"Completed step {i+1}/{steps}")
        tc.log(f"Finished step {i+1}", level="debug")
        
        if delay > 0:
            time.sleep(delay)
    
    tc.log("Progress task completed successfully", level="info")
    return f"completed_{steps}_steps"

@task
def logging_task(tc: TaskControl, messages: List[str]) -> int:
    """Task that logs multiple messages at different levels."""
    tc.log("Starting logging task", level="info")
    
    log_levels = ["debug", "info", "warning", "error"]
    
    for i, message in enumerate(messages):
        level = log_levels[i % len(log_levels)]
        tc.log(f"Message {i+1}: {message}", level=level, step=i+1, total=len(messages))
        
        # Update progress
        progress = (i + 1) / len(messages) * 100
        tc.update_progress(progress, f"Logged message {i+1}")
    
    tc.log("Logging task completed", level="info")
    return len(messages)

@task
def data_processing_task(tc: TaskControl, data: List[Dict[str, Any]], batch_size: int = 5) -> Dict[str, Any]:
    """Task that processes data in batches with detailed progress reporting."""
    total_items = len(data)
    processed_items = []
    errors = []
    
    tc.log(f"Processing {total_items} items in batches of {batch_size}", level="info")
    
    for batch_start in range(0, total_items, batch_size):
        if tc.should_stop:
            tc.log("Processing stopped by user request", level="warning")
            break
            
        batch_end = min(batch_start + batch_size, total_items)
        batch = data[batch_start:batch_end]
        
        tc.update_status(f"Processing batch {batch_start//batch_size + 1}")
        tc.log(f"Processing batch {batch_start}-{batch_end-1}", level="debug")
        
        for i, item in enumerate(batch):
            try:
                # Simulate processing
                processed_item = {
                    **item,
                    "processed": True,
                    "batch_id": batch_start // batch_size + 1,
                    "item_index": batch_start + i
                }
                processed_items.append(processed_item)
                
            except Exception as e:
                error_info = {
                    "item_index": batch_start + i,
                    "error": str(e),
                    "item": item
                }
                errors.append(error_info)
                tc.log(f"Error processing item {batch_start + i}: {e}", level="error")
        
        # Update progress after each batch
        progress = min(batch_end / total_items * 100, 100)
        tc.update_progress(progress, f"Processed {batch_end}/{total_items} items")
        
        # Small delay to simulate work
        time.sleep(0.001)
    
    result = {
        "total_input": total_items,
        "processed_count": len(processed_items),
        "error_count": len(errors),
        "processed_items": processed_items,
        "errors": errors
    }
    
    tc.log(f"Data processing completed: {len(processed_items)} processed, {len(errors)} errors", level="info")
    return result

@task
def cancellable_long_task(tc: TaskControl, duration_seconds: float = 1.0, check_interval: float = 0.1) -> str:
    """Long-running task that can be cancelled and checks stop signal regularly."""
    tc.log(f"Starting long task (duration: {duration_seconds}s, check interval: {check_interval}s)", level="info")
    
    start_time = time.time()
    iterations = 0
    
    while True:
        current_time = time.time()
        elapsed = current_time - start_time
        
        if elapsed >= duration_seconds:
            break
            
        if tc.should_stop:
            tc.log(f"Task cancelled after {elapsed:.2f}s ({iterations} iterations)", level="warning")
            return f"cancelled_after_{elapsed:.2f}s_{iterations}_iterations"
        
        # Update progress based on time elapsed
        progress = min(elapsed / duration_seconds * 100, 100)
        tc.update_progress(progress, f"Running for {elapsed:.2f}s")
        
        iterations += 1
        time.sleep(check_interval)
    
    tc.log(f"Long task completed after {duration_seconds}s ({iterations} iterations)", level="info")
    return f"completed_{duration_seconds}s_{iterations}_iterations"

@task
def error_handling_task(tc: TaskControl, error_mode: str = "none") -> str:
    """Task that can simulate different types of errors for testing error handling."""
    tc.log(f"Starting error handling task with mode: {error_mode}", level="info")
    
    tc.update_progress(25.0, "Initializing")
    
    if error_mode == "early":
        tc.log("Simulating early error", level="error")
        raise ValueError("Early error in task execution")
    
    tc.update_progress(50.0, "Midpoint processing")
    
    if error_mode == "mid":
        tc.log("Simulating mid-execution error", level="error") 
        raise RuntimeError("Error during processing")
    
    tc.update_progress(75.0, "Near completion")
    
    if error_mode == "late":
        tc.log("Simulating late error", level="error")
        raise Exception("Error near task completion")
    
    tc.update_progress(100.0, "Completed successfully")
    tc.log("Error handling task completed without errors", level="info")
    return f"success_mode_{error_mode}"


# --- Parallel Task Definitions ---

@task
def concurrent_worker(worker_id: int, work_duration: float = 0.1, items_to_process: int = 3) -> Dict[str, Any]:
    """Worker task designed for concurrent execution testing."""
    start_time = time.time()
    processed_items = []
    
    for i in range(items_to_process):
        # Simulate work
        time.sleep(work_duration / items_to_process)
        processed_items.append(f"worker_{worker_id}_item_{i}")
    
    end_time = time.time()
    
    return {
        "worker_id": worker_id,
        "items_processed": len(processed_items),
        "items": processed_items,
        "duration": end_time - start_time,
        "start_time": start_time,
        "end_time": end_time
    }

@task
def resource_intensive_task(tc: TaskControl, iterations: int = 100, memory_size: int = 1000) -> Dict[str, Any]:
    """Task that uses CPU and memory resources for performance testing."""
    tc.log(f"Starting resource intensive task: {iterations} iterations, {memory_size} memory units", level="info")
    
    # Allocate some memory
    memory_data = list(range(memory_size))
    
    # CPU intensive work
    result_sum = 0
    for i in range(iterations):
        if tc.should_stop:
            tc.log(f"Resource intensive task stopped at iteration {i}", level="warning")
            break
            
        # Some CPU work
        for j in range(100):
            result_sum += (i * j) % 1000
        
        # Update progress every 10 iterations
        if i % 10 == 0:
            progress = i / iterations * 100
            tc.update_progress(progress, f"Iteration {i}/{iterations}")
    
    tc.log("Resource intensive task completed", level="info")
    
    return {
        "iterations_completed": i + 1 if 'i' in locals() else iterations,
        "result_sum": result_sum,
        "memory_size": len(memory_data),
        "completed": not tc.should_stop
    }


# --- Configuration Tasks ---

@task(max_parallel=3, size="large")
def parallel_task(task_id: int, delay: float = 0.1) -> str:
    """Task configured for parallel execution."""
    time.sleep(delay)
    return f"parallel_task_{task_id}_completed"

@task(max_parallel=1, size="small")
def sequential_task(task_id: int, shared_resource: str = "default") -> str:
    """Task that must run sequentially (max_parallel=1)."""
    time.sleep(0.05)  # Simulate exclusive resource access
    return f"sequential_{task_id}_{shared_resource}"


# --- Helper Functions for Test Data ---

def create_test_data(count: int = 10) -> List[Dict[str, Any]]:
    """Create test data for data processing tasks."""
    return [
        {
            "id": i,
            "name": f"item_{i}",
            "value": i * 10,
            "category": "test",
            "metadata": {"created": f"2023-01-{i+1:02d}"}
        }
        for i in range(count)
    ]

def create_large_test_data(count: int = 1000) -> List[Dict[str, Any]]:
    """Create larger test dataset for performance testing."""
    return [
        {
            "id": i,
            "data": f"data_string_{i}" * 10,  # Larger strings
            "numbers": list(range(i, i + 10)),
            "computed": i ** 2 + i * 3
        }
        for i in range(count)
    ] 