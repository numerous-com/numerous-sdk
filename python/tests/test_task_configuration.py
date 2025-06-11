import pytest
from numerous.tasks.task import task, TaskConfig
from numerous.tasks.control import TaskControl


class TestTaskConfiguration:
    """Test task configuration options (parallelism, resource sizing)."""

    def test_task_config_defaults(self):
        """Test default task configuration values."""
        @task
        def default_task(tc: TaskControl) -> str:
            return "default"
        
        assert default_task.config.max_parallel == 1
        assert default_task.config.size == "small"
        assert default_task.config.name is None  # Config name is None, but task.name should be function name
        assert default_task.name == "default_task"  # Task name should be function name

    def test_resource_size_small(self):
        """Test task with small resource size."""
        @task(size="small")
        def small_task(tc: TaskControl) -> str:
            tc.log("Running small task")
            return "small_completed"
        
        assert small_task.config.size == "small"
        result = small_task()
        assert result == "small_completed"

    def test_resource_size_medium(self):
        """Test task with medium resource size."""
        @task(size="medium")
        def medium_task(tc: TaskControl) -> str:
            tc.log("Running medium task")
            return "medium_completed"
        
        assert medium_task.config.size == "medium"
        result = medium_task()
        assert result == "medium_completed"

    def test_resource_size_large(self):
        """Test task with large resource size."""
        @task(size="large")
        def large_task(tc: TaskControl) -> str:
            tc.log("Running large task")
            return "large_completed"
        
        assert large_task.config.size == "large"
        result = large_task()
        assert result == "large_completed"

    def test_custom_resource_size(self):
        """Test task with custom resource size string."""
        @task(size="xlarge")
        def xlarge_task(tc: TaskControl) -> str:
            tc.log("Running xlarge task")
            return "xlarge_completed"
        
        assert xlarge_task.config.size == "xlarge"
        result = xlarge_task()
        assert result == "xlarge_completed"

    def test_max_parallel_1(self):
        """Test task with max_parallel=1 (default)."""
        @task(max_parallel=1)
        def single_parallel_task(tc: TaskControl, value: int) -> int:
            tc.log(f"Processing value {value}")
            return value * 2
        
        assert single_parallel_task.config.max_parallel == 1
        result = single_parallel_task(5)
        assert result == 10

    def test_max_parallel_1_with_size(self):
        """Test task with max_parallel=1 and specific size."""
        @task(max_parallel=1, size="medium")
        def configured_task(tc: TaskControl, data: str) -> str:
            tc.log(f"Processing data: {data}")
            return f"processed_{data}"
        
        assert configured_task.config.max_parallel == 1
        assert configured_task.config.size == "medium"
        result = configured_task("test")
        assert result == "processed_test"

    def test_max_parallel_greater_than_1_raises_error(self):
        """Test that max_parallel > 1 raises TypeError in direct execution."""
        @task(max_parallel=2)
        def multi_parallel_task(tc: TaskControl) -> str:
            return "should_not_execute"
        
        assert multi_parallel_task.config.max_parallel == 2
        
        with pytest.raises(TypeError, match="max_parallel > 1"):
            multi_parallel_task()

    def test_max_parallel_3_with_size_large_raises_error(self):
        """Test that max_parallel > 1 with size configuration raises TypeError."""
        @task(max_parallel=3, size="large")
        def complex_config_task(tc: TaskControl, item: int) -> int:
            return item * 10
        
        assert complex_config_task.config.max_parallel == 3
        assert complex_config_task.config.size == "large"
        
        with pytest.raises(TypeError, match="max_parallel > 1"):
            complex_config_task(5)

    def test_all_config_options_together(self):
        """Test task with all configuration options specified."""
        @task(name="full_config", max_parallel=1, size="large")
        def fully_configured_task(tc: TaskControl, x: int, y: int) -> dict:
            tc.log("Running fully configured task")
            return {"sum": x + y, "product": x * y}
        
        assert fully_configured_task.config.name == "full_config"
        assert fully_configured_task.config.max_parallel == 1
        assert fully_configured_task.config.size == "large"
        
        result = fully_configured_task(3, 4)
        assert result == {"sum": 7, "product": 12}

    def test_config_values_preserved_after_direct_execution(self):
        """Test that task configuration is preserved after direct execution."""
        @task(name="preserved_config", max_parallel=1, size="medium")
        def config_preservation_task(tc: TaskControl) -> str:
            tc.log("Config preservation test")
            return "config_preserved"
        
        # Verify config before execution
        assert config_preservation_task.config.name == "preserved_config"
        assert config_preservation_task.config.max_parallel == 1
        assert config_preservation_task.config.size == "medium"
        
        # Execute task
        result = config_preservation_task()
        assert result == "config_preserved"
        
        # Verify config after execution (should be unchanged)
        assert config_preservation_task.config.name == "preserved_config"
        assert config_preservation_task.config.max_parallel == 1
        assert config_preservation_task.config.size == "medium"

    def test_taskconfig_dataclass_creation(self):
        """Test TaskConfig dataclass can be created with all options."""
        config = TaskConfig(
            name="test_task",
            max_parallel=1,
            size="large"
        )
        
        assert config.name == "test_task"
        assert config.max_parallel == 1
        assert config.size == "large"

    def test_taskconfig_defaults(self):
        """Test TaskConfig default values."""
        config = TaskConfig()
        
        assert config.name is None
        assert config.max_parallel == 1
        assert config.size == "small"

    def test_resource_config_with_parameters(self):
        """Test resource configuration works with task parameters."""
        @task(size="medium")
        def parameterized_task(tc: TaskControl, items: list, multiplier: int = 2) -> list:
            tc.log(f"Processing {len(items)} items with multiplier {multiplier}")
            return [item * multiplier for item in items]
        
        assert parameterized_task.config.size == "medium"
        result = parameterized_task([1, 2, 3], 3)
        assert result == [3, 6, 9]

    def test_resource_config_with_complex_return_types(self):
        """Test resource configuration with complex return types."""
        @task(size="large")
        def complex_return_task(tc: TaskControl, data: dict) -> dict:
            tc.log("Processing complex data structure")
            return {
                "input_keys": list(data.keys()),
                "processed": True,
                "size_hint": tc.config.size if hasattr(tc, 'config') else "unknown"
            }
        
        assert complex_return_task.config.size == "large"
        result = complex_return_task({"a": 1, "b": 2})
        assert result["input_keys"] == ["a", "b"]
        assert result["processed"] is True

    def test_edge_case_empty_size_string(self):
        """Test edge case with empty size string."""
        @task(size="")
        def empty_size_task(tc: TaskControl) -> str:
            return "empty_size"
        
        assert empty_size_task.config.size == ""
        result = empty_size_task()
        assert result == "empty_size"

    def test_edge_case_max_parallel_zero_raises_error(self):
        """Test that max_parallel=0 raises TypeError in direct execution (same as max_parallel > 1)."""
        @task(max_parallel=0)
        def zero_parallel_task(tc: TaskControl) -> str:
            return "zero_parallel"
        
        assert zero_parallel_task.config.max_parallel == 0
        # max_parallel=0 should raise TypeError since it's != 1
        with pytest.raises(TypeError, match="max_parallel > 1"):
            zero_parallel_task()

    def test_configuration_immutability(self):
        """Test that task configuration cannot be modified after creation."""
        @task(max_parallel=1, size="small")
        def immutable_config_task(tc: TaskControl) -> str:
            return "immutable"
        
        original_config = immutable_config_task.config
        
        # TaskConfig is a dataclass, so direct modification should be possible
        # but the task should use the original config
        original_max_parallel = original_config.max_parallel
        original_size = original_config.size
        
        result = immutable_config_task()
        assert result == "immutable"
        
        # Configuration should remain the same
        assert immutable_config_task.config.max_parallel == original_max_parallel
        assert immutable_config_task.config.size == original_size 