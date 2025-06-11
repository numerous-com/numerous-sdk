"""
Test TaskControl state accessibility in Stage 1: Direct Function Execution

This module tests whether TaskControl progress, status, and other information
can be accessed after task execution in direct mode.
"""

import pytest
import io
import sys
from unittest.mock import patch
from contextlib import redirect_stdout

from numerous.tasks import task, TaskControl


class TestDirectExecutionTaskControl:
    """Test TaskControl state accessibility in direct execution mode."""
    
    def test_taskcontrol_state_after_direct_execution(self):
        """Test that TaskControl state is accessible after direct execution."""
        task_control_instance = None
        
        @task
        def task_with_progress(tc: TaskControl, steps: int) -> str:
            nonlocal task_control_instance
            task_control_instance = tc  # Capture the TaskControl instance
            
            for i in range(steps):
                progress = (i + 1) / steps * 100
                tc.update_progress(progress, f"Step {i+1} of {steps}")
                tc.log(f"Completed step {i+1}")
            
            tc.update_status("completed")
            tc.log("Task finished successfully")
            return "Task completed"
        
        # Execute the task
        result = task_with_progress(5)
        assert result == "Task completed"
        
        # Verify TaskControl instance was captured
        assert task_control_instance is not None
        
        # Verify we can access the final state
        assert task_control_instance.progress == 100.0
        assert task_control_instance.status == "completed"
        assert task_control_instance.instance_id is not None
        assert task_control_instance.task_definition_name == "task_with_progress"
    
    def test_taskcontrol_progressive_state_updates(self):
        """Test that TaskControl state updates are properly maintained during execution."""
        captured_states = []
        
        @task  
        def progressive_task(tc: TaskControl, iterations: int) -> list:
            for i in range(iterations):
                progress = (i + 1) / iterations * 100
                status = f"iteration_{i+1}"
                
                tc.update_progress(progress, status)
                
                # Capture current state
                captured_states.append({
                    'progress': tc.progress,
                    'status': tc.status,
                    'iteration': i + 1
                })
            
            return captured_states
        
        # Execute task
        result = progressive_task(3)
        
        # Verify the captured states show proper progression
        assert len(result) == 3
        
        # Check first iteration
        assert result[0]['progress'] == pytest.approx(33.33, abs=0.01)
        assert result[0]['status'] == "iteration_1"
        assert result[0]['iteration'] == 1
        
        # Check final iteration
        assert result[2]['progress'] == 100.0
        assert result[2]['status'] == "iteration_3"
        assert result[2]['iteration'] == 3
    
    def test_taskcontrol_should_stop_functionality(self):
        """Test that should_stop functionality works in direct execution."""
        task_control_instance = None
        
        @task
        def stoppable_task(tc: TaskControl, max_iterations: int) -> dict:
            nonlocal task_control_instance
            task_control_instance = tc
            
            completed_iterations = 0
            for i in range(max_iterations):
                if tc.should_stop:
                    tc.log(f"Task stopped at iteration {i}")
                    break
                    
                tc.update_progress((i + 1) / max_iterations * 100, f"iteration_{i+1}")
                completed_iterations += 1
                
                # Simulate stop request after 2 iterations
                if i == 1:
                    tc.request_stop()
            
            return {
                'completed_iterations': completed_iterations,
                'was_stopped': tc.should_stop,
                'final_progress': tc.progress
            }
        
        # Execute task
        result = stoppable_task(10)
        
        # Verify task was stopped properly
        assert result['completed_iterations'] == 2  # Should stop after 2 iterations
        assert result['was_stopped'] == True
        assert result['final_progress'] == 20.0  # 2/10 * 100
        
        # Verify TaskControl instance state
        assert task_control_instance.should_stop == True
    
    def test_taskcontrol_multiple_tasks_isolation(self):  
        """Test that TaskControl instances are properly isolated between task executions."""
        first_tc = None
        second_tc = None
        
        @task
        def first_task(tc: TaskControl) -> str:
            nonlocal first_tc
            first_tc = tc
            tc.update_progress(50.0, "first_task_status")
            tc.log("First task log")
            return "first"
        
        @task
        def second_task(tc: TaskControl) -> str:
            nonlocal second_tc
            second_tc = tc
            tc.update_progress(75.0, "second_task_status")
            tc.log("Second task log")
            return "second"
        
        # Execute both tasks
        result1 = first_task()
        result2 = second_task()
        
        assert result1 == "first"
        assert result2 == "second"
        
        # Verify TaskControl instances are different
        assert first_tc is not second_tc
        assert first_tc.instance_id != second_tc.instance_id
        
        # Verify each maintains its own state
        assert first_tc.progress == 50.0
        assert first_tc.status == "first_task_status"
        assert second_tc.progress == 75.0
        assert second_tc.status == "second_task_status"
    
    def test_taskcontrol_logging_captured(self):
        """Test that TaskControl logging works in direct execution."""
        logged_messages = []
        
        # Custom handler to capture log messages
        from numerous.tasks.control import TaskControlHandler, set_task_control_handler
        
        class CapturingHandler(TaskControlHandler):
            def log(self, task_control, message: str, level: str, **extra_data):
                logged_messages.append({
                    'message': message,
                    'level': level,
                    'task_id': task_control.instance_id,
                    'task_name': task_control.task_definition_name,
                    'extra_data': extra_data
                })
            
            def update_progress(self, task_control, progress: float, status):
                pass  # No-op for this test
                
            def update_status(self, task_control, status: str):
                pass  # No-op for this test
                
            def request_stop(self, task_control):
                task_control._should_stop_internal = True
        
        # Set custom handler
        original_handler = None
        try:
            capturing_handler = CapturingHandler()
            set_task_control_handler(capturing_handler)
            
            @task
            def logging_task(tc: TaskControl, message: str) -> str:
                tc.log(f"Processing: {message}", level="info", extra_field="test_value")
                tc.log("Debug message", level="debug")
                tc.log("Warning message", level="warning", error_code=123)
                return f"Processed: {message}"
            
            # Execute task
            result = logging_task("test input")
            assert result == "Processed: test input"
            
            # Verify logs were captured
            assert len(logged_messages) == 3
            
            # Check first log message
            assert logged_messages[0]['message'] == "Processing: test input"
            assert logged_messages[0]['level'] == "info"
            assert logged_messages[0]['extra_data'] == {'extra_field': 'test_value'}
            assert logged_messages[0]['task_name'] == "logging_task"
            
            # Check other log messages
            assert logged_messages[1]['message'] == "Debug message"
            assert logged_messages[1]['level'] == "debug"
            assert logged_messages[2]['message'] == "Warning message"
            assert logged_messages[2]['level'] == "warning"
            assert logged_messages[2]['extra_data'] == {'error_code': 123}
            
        finally:
            # Reset to default handler
            set_task_control_handler(None)
    
    def test_taskcontrol_persistence_across_function_calls(self):
        """Test that TaskControl state persists across multiple function calls within a task."""
        @task
        def multi_step_task(tc: TaskControl, data: list) -> dict:
            def process_step_1(tc_ref):
                tc_ref.update_progress(25.0, "step_1_complete")
                tc_ref.log("Step 1 completed")
                return len(data)
            
            def process_step_2(tc_ref, count):
                tc_ref.update_progress(50.0, "step_2_complete")
                tc_ref.log("Step 2 completed")
                return count * 2
            
            def process_step_3(tc_ref, value):
                tc_ref.update_progress(75.0, "step_3_complete")
                tc_ref.log("Step 3 completed")
                return value + 10
            
            def finalize(tc_ref, final_value):
                tc_ref.update_progress(100.0, "all_complete")
                tc_ref.log("All steps completed")
                return final_value
            
            # Execute steps, passing TaskControl through
            count = process_step_1(tc)
            doubled = process_step_2(tc, count)
            added = process_step_3(tc, doubled)
            final = finalize(tc, added)
            
            return {
                'final_value': final,
                'final_progress': tc.progress,
                'final_status': tc.status
            }
        
        # Execute task
        result = multi_step_task([1, 2, 3])
        
        # Verify the result and final state
        assert result['final_value'] == 16  # (3 * 2) + 10
        assert result['final_progress'] == 100.0
        assert result['final_status'] == "all_complete" 