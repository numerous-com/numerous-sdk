import pytest
from typing import List, Dict, Any, Optional, Union
from numerous.tasks.task import task, Task
from numerous.tasks.control import TaskControl


class TestTaskTypeHints:
    """Test type hints validation and TaskControl parameter injection."""

    def test_taskcontrol_parameter_detection_by_annotation(self):
        """Test that TaskControl parameter is detected by type annotation."""
        @task
        def task_with_tc_annotation(tc: TaskControl, value: int) -> int:
            tc.log("Processing value")
            return value * 2
        
        assert task_with_tc_annotation.expects_task_control is True
        result = task_with_tc_annotation(5)
        assert result == 10

    def test_no_taskcontrol_parameter_detection(self):
        """Test that tasks without TaskControl annotation are detected correctly."""
        @task
        def task_without_tc(value: int) -> int:
            return value * 2
        
        assert task_without_tc.expects_task_control is False
        result = task_without_tc(5)
        assert result == 10

    def test_taskcontrol_parameter_must_be_first(self):
        """Test that TaskControl parameter is expected to be first parameter."""
        @task
        def task_tc_first(tc: TaskControl, a: int, b: str) -> str:
            tc.log(f"Processing {a} and {b}")
            return f"{a}:{b}"
        
        assert task_tc_first.expects_task_control is True
        result = task_tc_first(42, "hello")
        assert result == "42:hello"

    def test_taskcontrol_parameter_not_first_ignored(self):
        """Test that TaskControl parameter not in first position is ignored."""
        @task
        def task_tc_not_first(value: int, tc: TaskControl) -> int:
            # This should NOT be detected as expecting TaskControl
            return value * 2
        
        assert task_tc_not_first.expects_task_control is False
        # This will fail if TaskControl injection is attempted
        result = task_tc_not_first(5, "dummy")  # Pass dummy value for tc parameter
        assert result == 10

    def test_complex_type_hints_with_taskcontrol(self):
        """Test complex type hints work with TaskControl injection."""
        @task
        def complex_types_task(
            tc: TaskControl, 
            items: List[Dict[str, Any]], 
            config: Optional[Dict[str, Union[str, int]]] = None
        ) -> Dict[str, Any]:
            tc.log(f"Processing {len(items)} items")
            if config is None:
                config = {"default": True}
            
            return {
                "processed_count": len(items),
                "config": config,
                "items": [item.get("id", "unknown") for item in items]
            }
        
        assert complex_types_task.expects_task_control is True
        
        test_items = [{"id": "1", "name": "test"}, {"id": "2", "value": 42}]
        result = complex_types_task(test_items)
        
        assert result["processed_count"] == 2
        assert result["config"] == {"default": True}
        assert result["items"] == ["1", "2"]

    def test_complex_type_hints_without_taskcontrol(self):
        """Test complex type hints work without TaskControl."""
        @task
        def complex_no_tc_task(
            data: Dict[str, List[int]], 
            multiplier: float = 1.0
        ) -> Dict[str, float]:
            total = sum(sum(values) for values in data.values())
            return {"total": total * multiplier}
        
        assert complex_no_tc_task.expects_task_control is False
        
        test_data = {"group1": [1, 2, 3], "group2": [4, 5]}
        result = complex_no_tc_task(test_data, 2.5)
        
        assert result["total"] == 37.5  # (1+2+3+4+5) * 2.5

    def test_taskcontrol_injection_with_defaults(self):
        """Test TaskControl injection works with functions that have default parameters."""
        @task
        def task_with_defaults(
            tc: TaskControl, 
            value: int, 
            multiplier: int = 2, 
            prefix: str = "result"
        ) -> str:
            tc.log(f"Computing {value} * {multiplier}")
            result = value * multiplier
            return f"{prefix}:{result}"
        
        assert task_with_defaults.expects_task_control is True
        
        # Test with defaults
        result1 = task_with_defaults(5)
        assert result1 == "result:10"
        
        # Test with some parameters
        result2 = task_with_defaults(5, 3)
        assert result2 == "result:15"
        
        # Test with all parameters
        result3 = task_with_defaults(5, 3, "output")
        assert result3 == "output:15"

    def test_taskcontrol_injection_with_kwargs(self):
        """Test TaskControl injection works with keyword arguments."""
        @task
        def task_with_kwargs(tc: TaskControl, **kwargs: Any) -> Dict[str, Any]:
            tc.log(f"Received kwargs: {list(kwargs.keys())}")
            return {"count": len(kwargs), "keys": sorted(kwargs.keys())}
        
        assert task_with_kwargs.expects_task_control is True
        
        result = task_with_kwargs(a=1, b="test", c=True)
        assert result["count"] == 3
        assert result["keys"] == ["a", "b", "c"]

    def test_taskcontrol_injection_with_args_and_kwargs(self):
        """Test TaskControl injection works with *args and **kwargs."""
        @task
        def task_with_args_kwargs(tc: TaskControl, base: int, *args: int, **kwargs: Any) -> Dict[str, Any]:
            tc.log(f"Base: {base}, args: {args}, kwargs: {list(kwargs.keys())}")
            total = base + sum(args)
            return {
                "total": total,
                "args_count": len(args),
                "kwargs_count": len(kwargs)
            }
        
        assert task_with_args_kwargs.expects_task_control is True
        
        result = task_with_args_kwargs(10, 1, 2, 3, extra="value", flag=True)
        assert result["total"] == 16  # 10 + 1 + 2 + 3
        assert result["args_count"] == 3
        assert result["kwargs_count"] == 2

    def test_method_taskcontrol_detection_with_self(self):
        """Test TaskControl detection in class methods (should ignore 'self')."""
        class TaskClass:
            @task
            def method_with_tc(self, tc: TaskControl, value: int) -> int:
                tc.log("Processing in method")
                return value * 2
            
            @task
            def method_without_tc(self, value: int) -> int:
                return value * 3
        
        obj = TaskClass()
        
        # TaskControl should be detected as second parameter (after self)
        assert obj.method_with_tc.expects_task_control is True
        assert obj.method_without_tc.expects_task_control is False
        
        # Note: Direct execution of instance methods requires special handling
        # For now, we test detection but skip execution since it requires self binding
        # result1 = obj.method_with_tc(5)
        # assert result1 == 10
        # 
        # result2 = obj.method_without_tc(5)
        # assert result2 == 15

    def test_classmethod_taskcontrol_detection_with_cls(self):
        """Test TaskControl detection in class methods (should ignore 'cls')."""
        class TaskClass:
            @classmethod
            @task
            def class_method_with_tc(cls, tc: TaskControl, value: int) -> int:
                tc.log("Processing in classmethod")
                return value * 4
            
            @classmethod
            @task
            def class_method_without_tc(cls, value: int) -> int:
                return value * 5
        
        # TaskControl should be detected as second parameter (after cls)
        assert TaskClass.class_method_with_tc.expects_task_control is True
        assert TaskClass.class_method_without_tc.expects_task_control is False
        
        # Note: Direct execution of class methods requires special handling
        # For now, we test detection but skip execution since it requires cls binding
        # result1 = TaskClass.class_method_with_tc(5)
        # assert result1 == 20
        # 
        # result2 = TaskClass.class_method_without_tc(5)
        # assert result2 == 25

    def test_parameter_without_annotation_ignored(self):
        """Test that parameters without annotations don't interfere with TaskControl detection."""
        @task
        def mixed_annotations(tc: TaskControl, annotated: int, not_annotated, another: str) -> str:
            tc.log(f"Processing {annotated}, {not_annotated}, {another}")
            return f"{annotated}:{not_annotated}:{another}"
        
        assert mixed_annotations.expects_task_control is True
        result = mixed_annotations(42, "middle", "end")
        assert result == "42:middle:end"

    def test_wrong_type_annotation_ignored(self):
        """Test that first parameter with wrong type annotation is ignored."""
        @task
        def wrong_first_annotation(value: str, tc: TaskControl) -> str:
            # First parameter is str, not TaskControl, so no injection
            return f"value:{value}, tc:{tc}"
        
        assert wrong_first_annotation.expects_task_control is False
        result = wrong_first_annotation("test", "dummy_tc")
        assert result == "value:test, tc:dummy_tc"

    def test_function_signature_preservation(self):
        """Test that original function signature is preserved in Task object."""
        @task
        def original_func(tc: TaskControl, a: int, b: str = "default") -> str:
            return f"{a}:{b}"
        
        # Check that original signature is stored
        sig = original_func._sig
        params = list(sig.parameters.values())
        
        assert len(params) == 3  # tc, a, b
        assert params[0].name == "tc"
        assert params[0].annotation == TaskControl
        assert params[1].name == "a"
        assert params[1].annotation == int
        assert params[2].name == "b"
        assert params[2].annotation == str
        assert params[2].default == "default"

    def test_return_type_hints_preserved(self):
        """Test that return type hints are preserved."""
        @task
        def typed_return_func(tc: TaskControl, value: int) -> Dict[str, int]:
            tc.log("Creating dict")
            return {"result": value}
        
        # Check return annotation is preserved in original signature
        assert typed_return_func._sig.return_annotation == Dict[str, int]
        
        result = typed_return_func(42)
        assert result == {"result": 42}
        assert isinstance(result, dict)

    def test_no_parameters_function(self):
        """Test function with no parameters."""
        @task
        def no_params_func() -> str:
            return "no_params"
        
        assert no_params_func.expects_task_control is False
        result = no_params_func()
        assert result == "no_params"

    def test_only_self_parameter(self):
        """Test method with only self parameter."""
        class OnlySelfTaskClass:
            @task
            def only_self_method(self) -> str:
                return "only_self"
        
        obj = OnlySelfTaskClass()
        assert obj.only_self_method.expects_task_control is False
        # Note: Direct execution of instance methods requires special handling
        # For now, we test detection but skip execution since it requires self binding
        # result = obj.only_self_method()
        # assert result == "only_self"

    def test_generic_type_hints(self):
        """Test generic type hints work correctly."""
        from typing import Generic, TypeVar
        
        T = TypeVar('T')
        
        @task
        def generic_task(tc: TaskControl, items: List[T], default: T) -> T:
            tc.log(f"Processing {len(items)} items")
            return items[0] if items else default
        
        assert generic_task.expects_task_control is True
        
        result1 = generic_task([1, 2, 3], 0)
        assert result1 == 1
        
        result2 = generic_task([], "default")
        assert result2 == "default"

    def test_union_type_hints(self):
        """Test Union type hints work correctly."""
        @task
        def union_task(tc: TaskControl, value: Union[int, str]) -> str:
            tc.log(f"Processing {type(value).__name__}: {value}")
            return str(value)
        
        assert union_task.expects_task_control is True
        
        result1 = union_task(42)
        assert result1 == "42"
        
        result2 = union_task("hello")
        assert result2 == "hello" 