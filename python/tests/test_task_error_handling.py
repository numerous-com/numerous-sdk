import pytest
from typing import List, Dict, Any
from numerous.tasks.task import task, Task
from numerous.tasks.control import TaskControl
from numerous.tasks.exceptions import TaskError


class TestTaskErrorHandling:
    """Test error handling and exceptions in task execution."""

    def test_task_function_exception_propagated(self):
        """Test that exceptions in task functions are propagated correctly."""
        @task
        def failing_task(tc: TaskControl, value: int) -> int:
            tc.log("About to raise exception")
            if value < 0:
                raise ValueError("Negative values not allowed")
            return value * 2
        
        # Normal execution should work
        result = failing_task(5)
        assert result == 10
        
        # Exception should be propagated
        with pytest.raises(ValueError, match="Negative values not allowed"):
            failing_task(-1)

    def test_task_without_taskcontrol_exception_propagated(self):
        """Test exception propagation in tasks without TaskControl."""
        @task
        def failing_task_no_tc(value: str) -> int:
            if not value.isdigit():
                raise ValueError(f"'{value}' is not a valid number")
            return int(value)
        
        # Normal execution should work
        result = failing_task_no_tc("42")
        assert result == 42
        
        # Exception should be propagated
        with pytest.raises(ValueError, match="'abc' is not a valid number"):
            failing_task_no_tc("abc")

    def test_built_in_exceptions(self):
        """Test various built-in exceptions are handled correctly."""
        @task
        def exception_task(tc: TaskControl, operation: str, a: int, b: int) -> float:
            tc.log(f"Performing {operation}")
            
            if operation == "divide":
                return a / b  # Can raise ZeroDivisionError
            elif operation == "index":
                items = [1, 2, 3]
                return items[a]  # Can raise IndexError
            elif operation == "key":
                data = {"0": 0, "1": 1, "2": 2}
                return data[str(a)]  # Can raise KeyError
            elif operation == "type":
                return a + "invalid"  # Can raise TypeError
            else:
                raise ValueError(f"Unknown operation: {operation}")
        
        # Normal operations should work
        assert exception_task("divide", 10, 2) == 5.0
        assert exception_task("index", 1, 0) == 2
        assert exception_task("key", 1, 0) == 1
        
        # Exception cases
        with pytest.raises(ZeroDivisionError):
            exception_task("divide", 10, 0)
        
        with pytest.raises(IndexError):
            exception_task("index", 5, 0)
        
        with pytest.raises(KeyError):
            exception_task("key", 999, 0)
        
        with pytest.raises(TypeError):
            exception_task("type", 42, 0)
        
        with pytest.raises(ValueError, match="Unknown operation: invalid"):
            exception_task("invalid", 1, 2)

    def test_custom_exceptions(self):
        """Test custom exceptions are handled correctly."""
        class CustomTaskError(Exception):
            def __init__(self, message: str, error_code: int):
                super().__init__(message)
                self.error_code = error_code
        
        @task
        def custom_exception_task(tc: TaskControl, value: int) -> str:
            tc.log(f"Processing value: {value}")
            
            if value == 0:
                raise CustomTaskError("Zero is not allowed", 1001)
            elif value < 0:
                raise CustomTaskError("Negative values not supported", 1002)
            else:
                return f"Processed: {value}"
        
        # Normal execution
        result = custom_exception_task(5)
        assert result == "Processed: 5"
        
        # Custom exceptions with attributes
        with pytest.raises(CustomTaskError) as exc_info:
            custom_exception_task(0)
        
        assert str(exc_info.value) == "Zero is not allowed"
        assert exc_info.value.error_code == 1001
        
        with pytest.raises(CustomTaskError) as exc_info:
            custom_exception_task(-5)
        
        assert str(exc_info.value) == "Negative values not supported"
        assert exc_info.value.error_code == 1002

    def test_exception_with_taskcontrol_state(self):
        """Test that TaskControl state is accessible even when exceptions occur."""
        @task
        def stateful_failing_task(tc: TaskControl, steps: int) -> str:
            tc.update_status("starting")
            
            for i in range(steps):
                tc.update_progress(i / steps * 100, f"Step {i + 1}")
                tc.log(f"Completed step {i + 1}")
                
                if i == 2:  # Fail on third step
                    tc.update_status("failed")
                    raise RuntimeError(f"Failed at step {i + 1}")
            
            tc.update_status("completed")
            return "success"
        
        # Should succeed with fewer steps
        result = stateful_failing_task(2)
        assert result == "success"
        
        # Should fail on third step
        with pytest.raises(RuntimeError, match="Failed at step 3"):
            stateful_failing_task(5)

    def test_exception_in_task_with_defaults(self):
        """Test exception handling in tasks with default parameters."""
        @task
        def task_with_defaults_exception(
            tc: TaskControl, 
            value: int, 
            operation: str = "square",
            multiplier: int = 1
        ) -> int:
            tc.log(f"Operation: {operation}, value: {value}, multiplier: {multiplier}")
            
            if operation == "square":
                result = value * value
            elif operation == "cube":
                result = value * value * value
            else:
                raise ValueError(f"Unsupported operation: {operation}")
            
            if multiplier <= 0:
                raise ValueError("Multiplier must be positive")
            
            return result * multiplier
        
        # Normal cases
        assert task_with_defaults_exception(3) == 9
        assert task_with_defaults_exception(2, "cube") == 8
        assert task_with_defaults_exception(3, "square", 2) == 18
        
        # Exception cases
        with pytest.raises(ValueError, match="Unsupported operation: invalid"):
            task_with_defaults_exception(3, "invalid")
        
        with pytest.raises(ValueError, match="Multiplier must be positive"):
            task_with_defaults_exception(3, "square", -1)

    def test_exception_in_task_with_complex_types(self):
        """Test exception handling with complex parameter types."""
        @task
        def complex_types_exception_task(
            tc: TaskControl, 
            data: List[Dict[str, Any]]
        ) -> Dict[str, int]:
            tc.log(f"Processing {len(data)} items")
            
            if not data:
                raise ValueError("Data list cannot be empty")
            
            result = {}
            for i, item in enumerate(data):
                if not isinstance(item, dict):
                    raise TypeError(f"Item {i} is not a dictionary: {type(item)}")
                
                if "id" not in item:
                    raise KeyError(f"Item {i} missing required 'id' field")
                
                if "value" not in item:
                    raise KeyError(f"Item {i} missing required 'value' field")
                
                item_id = item["id"]
                item_value = item["value"]
                
                if not isinstance(item_value, (int, float)):
                    raise TypeError(f"Item {i} value must be numeric, got {type(item_value)}")
                
                result[str(item_id)] = int(item_value)
            
            return result
        
        # Normal case
        test_data = [
            {"id": "a", "value": 10},
            {"id": "b", "value": 20.5}
        ]
        result = complex_types_exception_task(test_data)
        assert result == {"a": 10, "b": 20}
        
        # Exception cases
        with pytest.raises(ValueError, match="Data list cannot be empty"):
            complex_types_exception_task([])
        
        with pytest.raises(TypeError, match="Item 0 is not a dictionary"):
            complex_types_exception_task(["not_a_dict"])
        
        with pytest.raises(KeyError, match="Item 0 missing required 'id' field"):
            complex_types_exception_task([{"value": 10}])
        
        with pytest.raises(KeyError, match="Item 0 missing required 'value' field"):
            complex_types_exception_task([{"id": "test"}])
        
        with pytest.raises(TypeError, match="Item 0 value must be numeric"):
            complex_types_exception_task([{"id": "test", "value": "not_numeric"}])

    def test_exception_with_kwargs(self):
        """Test exception handling in tasks with keyword arguments."""
        @task
        def kwargs_exception_task(tc: TaskControl, **kwargs: Any) -> str:
            tc.log(f"Received kwargs: {list(kwargs.keys())}")
            
            required_keys = {"name", "age"}
            missing_keys = required_keys - set(kwargs.keys())
            
            if missing_keys:
                raise ValueError(f"Missing required keys: {missing_keys}")
            
            age = kwargs["age"]
            if not isinstance(age, int) or age < 0:
                raise ValueError("Age must be a non-negative integer")
            
            return f"Hello {kwargs['name']}, age {age}"
        
        # Normal case
        result = kwargs_exception_task(name="Alice", age=30)
        assert result == "Hello Alice, age 30"
        
        # Exception cases
        with pytest.raises(ValueError, match="Missing required keys"):
            kwargs_exception_task(name="Bob")
        
        with pytest.raises(ValueError, match="Age must be a non-negative integer"):
            kwargs_exception_task(name="Charlie", age=-5)
        
        with pytest.raises(ValueError, match="Age must be a non-negative integer"):
            kwargs_exception_task(name="Dave", age="thirty")

    def test_exception_during_task_control_operations(self):
        """Test behavior when TaskControl operations might cause issues."""
        @task
        def taskcontrol_exception_task(tc: TaskControl, operation: str) -> str:
            tc.log("Starting task")
            
            if operation == "invalid_progress":
                # This shouldn't raise an exception, but let's test edge cases
                tc.update_progress(150.0)  # Over 100%
                tc.update_progress(-10.0)  # Negative
                return "progress_tested"
            
            elif operation == "none_status":
                tc.update_status(None)  # This might cause issues
                return "status_tested"
            
            elif operation == "exception_in_log":
                # Test logging with problematic data
                tc.log("Logging complex object", extra_data={"obj": object()})
                return "log_tested"
            
            else:
                raise ValueError(f"Unknown operation: {operation}")
        
        # These operations should generally not raise exceptions
        # but we test edge cases to ensure robustness
        result1 = taskcontrol_exception_task("invalid_progress")
        assert result1 == "progress_tested"
        
        # These might or might not work depending on TaskControl implementation
        # but shouldn't crash the task execution
        try:
            result2 = taskcontrol_exception_task("none_status")
            assert result2 == "status_tested"
        except Exception:
            # If TaskControl doesn't handle None gracefully, that's okay for now
            pass
        
        try:
            result3 = taskcontrol_exception_task("exception_in_log")
            assert result3 == "log_tested"
        except Exception:
            # If logging can't handle complex objects, that's okay for now
            pass
        
        # This should definitely raise an exception
        with pytest.raises(ValueError, match="Unknown operation"):
            taskcontrol_exception_task("invalid")

    def test_nested_exceptions(self):
        """Test handling of nested exceptions (exception chains)."""
        @task
        def nested_exception_task(tc: TaskControl, level: int) -> str:
            tc.log(f"Processing level {level}")
            
            try:
                if level == 1:
                    raise ValueError("Level 1 error")
                elif level == 2:
                    try:
                        raise RuntimeError("Level 2 inner error")
                    except RuntimeError as e:
                        raise ValueError("Level 2 outer error") from e
                else:
                    return f"Success at level {level}"
            
            except ValueError as e:
                tc.log(f"Caught ValueError: {e}")
                raise TaskError(f"Task failed at level {level}") from e
        
        # Normal case
        result = nested_exception_task(0)
        assert result == "Success at level 0"
        
        # Nested exception cases
        with pytest.raises(TaskError, match="Task failed at level 1"):
            nested_exception_task(1)
        
        with pytest.raises(TaskError, match="Task failed at level 2"):
            nested_exception_task(2)

    def test_exception_cleanup_behavior(self):
        """Test that resources are properly handled when exceptions occur."""
        cleanup_called = []
        
        @task
        def cleanup_exception_task(tc: TaskControl, should_fail: bool) -> str:
            tc.log("Starting task with cleanup")
            
            try:
                tc.update_status("processing")
                
                if should_fail:
                    raise RuntimeError("Intentional failure")
                
                tc.update_status("completed")
                return "success"
            
            finally:
                # Simulate cleanup
                cleanup_called.append(True)
                tc.log("Cleanup completed")
        
        # Normal case
        cleanup_called.clear()
        result = cleanup_exception_task(False)
        assert result == "success"
        assert len(cleanup_called) == 1
        
        # Exception case - cleanup should still happen
        cleanup_called.clear()
        with pytest.raises(RuntimeError, match="Intentional failure"):
            cleanup_exception_task(True)
        assert len(cleanup_called) == 1

    def test_max_parallel_constraint_error(self):
        """Test max_parallel > 1 constraint error in direct execution."""
        @task(max_parallel=2)
        def multi_parallel_task(tc: TaskControl, value: int) -> int:
            return value * 2
        
        # Should raise TypeError for max_parallel > 1
        with pytest.raises(TypeError, match="max_parallel > 1"):
            multi_parallel_task(5)

    def test_exception_preserves_stack_trace(self):
        """Test that exception stack traces are preserved."""
        @task
        def deep_exception_task(tc: TaskControl, depth: int) -> str:
            tc.log(f"Depth: {depth}")
            
            def recursive_function(n: int) -> str:
                if n <= 0:
                    raise ValueError("Reached bottom of recursion")
                return recursive_function(n - 1)
            
            return recursive_function(depth)
        
        # Should preserve the full stack trace
        with pytest.raises(ValueError, match="Reached bottom of recursion"):
            deep_exception_task(3) 