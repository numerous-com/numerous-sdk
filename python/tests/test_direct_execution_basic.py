"""
Test Stage 1: Direct Function Execution (No Backend)

This module tests the @task decorator for direct function calls without any backend.
Stage 1 ensures that tasks can be executed as regular Python functions for initial
development and testing.
"""

import pytest
import os
from unittest.mock import patch

from numerous.tasks import task, TaskControl


class TestDirectExecutionBasic:
    """Test Stage 1: Direct function execution without any backend."""
    
    def test_simple_task_direct_call(self):
        """Test that @task decorator works for simple direct function calls."""
        @task
        def add_numbers(a: int, b: int) -> int:
            return a + b
        
        # Should work like a regular function call
        result = add_numbers(5, 3)
        assert result == 8
    
    def test_task_with_default_parameters(self):
        """Test direct execution with default parameters."""
        @task
        def greet(name: str, greeting: str = "Hello") -> str:
            return f"{greeting}, {name}!"
        
        result1 = greet("Alice")
        assert result1 == "Hello, Alice!"
        
        result2 = greet("Bob", "Hi")
        assert result2 == "Hi, Bob!"
    
    def test_task_with_keyword_arguments(self):
        """Test direct execution with keyword arguments."""
        @task
        def calculate(x: int, y: int, operation: str = "add") -> int:
            if operation == "add":
                return x + y
            elif operation == "multiply":
                return x * y
            else:
                raise ValueError(f"Unknown operation: {operation}")
        
        result1 = calculate(10, 5)
        assert result1 == 15
        
        result2 = calculate(x=10, y=5, operation="multiply")
        assert result2 == 50
        
        result3 = calculate(10, 5, operation="multiply")
        assert result3 == 50
    
    def test_task_without_taskcontrol_direct_call(self):
        """Test that tasks without TaskControl parameter work in direct execution."""
        @task
        def process_data(data: list) -> list:
            return [x * 2 for x in data if x > 0]
        
        input_data = [1, -2, 3, 0, 4]
        result = process_data(input_data)
        assert result == [2, 6, 8]
    
    def test_task_with_taskcontrol_direct_call(self):
        """Test that tasks with TaskControl parameter work in direct execution."""
        @task
        def task_with_control(tc: TaskControl, message: str) -> str:
            # TaskControl should be automatically injected and functional
            tc.log(f"Processing: {message}")
            tc.update_progress(50.0, "halfway")
            tc.update_status("processing")
            return f"Processed: {message}"
        
        # Should work in direct execution mode
        result = task_with_control("test message")
        assert result == "Processed: test message"
    
    def test_task_decorator_with_configuration(self):
        """Test that task configuration works in direct execution."""
        @task(name="custom_task", max_parallel=1, size="small")
        def configured_task(value: int) -> int:
            return value * 3
        
        result = configured_task(7)
        assert result == 21
        
        # Task object should have the correct configuration
        assert configured_task.name == "custom_task"
        assert configured_task.config.max_parallel == 1
        assert configured_task.config.size == "small"
    
    def test_direct_execution_with_exceptions(self):
        """Test that exceptions work correctly in direct execution."""
        @task
        def failing_task(should_fail: bool) -> str:
            if should_fail:
                raise ValueError("Task failed as requested")
            return "Success"
        
        # Should work normally
        result = failing_task(False)
        assert result == "Success"
        
        # Should raise exception normally
        with pytest.raises(ValueError, match="Task failed as requested"):
            failing_task(True)
    
    def test_task_with_complex_return_types(self):
        """Test direct execution with complex return types."""
        @task
        def complex_processing(data: dict) -> dict:
            result = {}
            for key, value in data.items():
                if isinstance(value, (int, float)):
                    result[f"processed_{key}"] = value * 2
                else:
                    result[f"processed_{key}"] = str(value).upper()
            return result
        
        input_data = {"a": 5, "b": "hello", "c": 3.14}
        result = complex_processing(input_data)
        
        expected = {
            "processed_a": 10,
            "processed_b": "HELLO", 
            "processed_c": 6.28
        }
        assert result == expected
    
    def test_no_api_environment_variables(self):
        """Test that direct execution works when no API environment variables are set."""
        # Ensure no API environment variables are set
        with patch.dict(os.environ, {}, clear=True):
            @task
            def clean_environment_task(x: int) -> int:
                return x ** 2
            
            result = clean_environment_task(4)
            assert result == 16
    
    def test_multiple_task_definitions(self):
        """Test that multiple task definitions work correctly."""
        @task
        def task_one(x: int) -> int:
            return x + 1
        
        @task
        def task_two(x: int) -> int:
            return x * 2
        
        @task
        def task_three(x: int, y: int) -> int:
            return x + y
        
        assert task_one(5) == 6
        assert task_two(5) == 10
        assert task_three(5, 3) == 8
    
    def test_task_with_variable_arguments(self):
        """Test direct execution with *args and **kwargs."""
        @task
        def flexible_task(*args, **kwargs) -> dict:
            return {
                "args": args,
                "kwargs": kwargs,
                "total_args": len(args),
                "total_kwargs": len(kwargs)
            }
        
        result = flexible_task(1, 2, 3, name="test", value=42)
        expected = {
            "args": (1, 2, 3),
            "kwargs": {"name": "test", "value": 42},
            "total_args": 3,
            "total_kwargs": 2
        }
        assert result == expected
    
    def test_task_introspection(self):
        """Test that task objects maintain proper introspection capabilities."""
        @task
        def documented_task(x: int, y: str = "default") -> str:
            """This is a documented task function."""
            return f"{y}: {x}"
        
        # Task should preserve function metadata
        assert documented_task.__name__ == "documented_task"
        assert documented_task.__doc__ == "This is a documented task function."
        
        # Should work normally
        result = documented_task(42)
        assert result == "default: 42"
        
        result2 = documented_task(42, "custom")
        assert result2 == "custom: 42"
    
    def test_task_max_parallel_gt_1_direct_call_error(self):
        """Test that tasks with max_parallel > 1 raise error in direct execution."""
        @task(max_parallel=2)
        def multi_parallel_task(x: int) -> int:
            return x * 2
        
        # Direct call should raise TypeError for max_parallel > 1
        with pytest.raises(TypeError, match="max_parallel > 1"):
            multi_parallel_task(5)
        
        # Task configuration should still be correct
        assert multi_parallel_task.config.max_parallel == 2
    
    def test_task_with_none_return(self):
        """Test direct execution with None return value."""
        @task
        def void_task(message: str) -> None:
            print(f"Processing: {message}")
        
        result = void_task("test")
        assert result is None
    
    def test_task_with_boolean_parameters_and_return(self):
        """Test direct execution with boolean types."""
        @task
        def boolean_logic(a: bool, b: bool, operation: str = "and") -> bool:
            if operation == "and":
                return a and b
            elif operation == "or":
                return a or b
            elif operation == "xor":
                return a != b
            else:
                raise ValueError(f"Unknown operation: {operation}")
        
        assert boolean_logic(True, True) == True
        assert boolean_logic(True, False) == False
        assert boolean_logic(True, False, "or") == True
        assert boolean_logic(True, False, "xor") == True
        assert boolean_logic(True, True, "xor") == False
    
    def test_task_with_nested_function_calls(self):
        """Test that tasks can call other tasks directly."""
        @task
        def helper_task(x: int) -> int:
            return x * 2
        
        @task
        def main_task(x: int) -> int:
            intermediate = helper_task(x)
            return intermediate + 1
        
        result = main_task(5)
        assert result == 11  # (5 * 2) + 1
    
    def test_task_type_annotations_preserved(self):
        """Test that type annotations are preserved in task objects."""
        @task
        def typed_task(x: int, y: str, z: float = 3.14) -> dict:
            return {"x": x, "y": y, "z": z}
        
        # Function should have proper type annotations
        import inspect
        sig = inspect.signature(typed_task._original_func)
        
        # Check parameter annotations
        params = sig.parameters
        assert params['x'].annotation == int
        assert params['y'].annotation == str
        assert params['z'].annotation == float
        assert params['z'].default == 3.14
        
        # Check return annotation
        assert sig.return_annotation == dict
        
        # Should work normally
        result = typed_task(42, "hello")
        expected = {"x": 42, "y": "hello", "z": 3.14}
        assert result == expected 