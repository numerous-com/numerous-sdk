"""
Unit tests for local versioning functionality.

Tests the local version generation and task definition extraction without requiring API connection.
"""

import pytest
from unittest.mock import Mock

from numerous.tasks.versioning import generate_local_version, extract_task_definition, get_task_version


class TestLocalVersioning:
    """Test local versioning functionality."""
    
    def test_generate_local_version(self):
        """Test local version generation from task definition."""
        task_definition = {
            "function_name": "test_task",
            "module": "test_module",
            "doc": "Test documentation",
            "parameters": {"param1": {"annotation": "str"}}
        }
        
        # Generate version
        version = generate_local_version(task_definition)
        
        # Verify format
        assert version.startswith("local-")
        assert len(version) == 14  # "local-" + 8 hex chars
        
        # Verify deterministic (same input = same output)
        version2 = generate_local_version(task_definition)
        assert version == version2
        
        # Verify different input = different output
        different_definition = task_definition.copy()
        different_definition["function_name"] = "different_task"
        different_version = generate_local_version(different_definition)
        assert version != different_version
    
    def test_extract_task_definition(self):
        """Test task definition extraction from function."""
        def sample_task(param1: str, param2: int = 42):
            """Sample task for testing."""
            pass
        
        # Add mock task config
        sample_task._task_config = Mock()
        sample_task._task_config.max_parallel = 3
        sample_task._task_config.resource_sizing = "small"
        sample_task._task_config.timeout = 300
        
        # Extract definition
        definition = extract_task_definition(sample_task)
        
        # Verify basic metadata
        assert definition["function_name"] == "sample_task"
        assert definition["doc"] == "Sample task for testing."
        assert definition["max_parallel"] == 3
        assert definition["resource_sizing"] == "small"
        assert definition["timeout"] == 300
        
        # Verify parameters
        assert "param1" in definition["parameters"]
        assert "param2" in definition["parameters"]
        assert definition["parameters"]["param1"]["annotation"] == "<class 'str'>"
        assert definition["parameters"]["param2"]["default"] == "42"
    
    def test_get_task_version_explicit(self):
        """Test get_task_version with explicit version."""
        def sample_task():
            pass
        
        explicit_version = "2.1.0"
        version = get_task_version(sample_task, explicit_version)
        
        # Should return explicit version
        assert version == explicit_version
    
    def test_get_task_version_local(self):
        """Test get_task_version with local version generation."""
        def sample_task():
            """Sample task for testing."""
            pass
        
        version = get_task_version(sample_task)
        
        # Should generate local version
        assert version.startswith("local-")
        assert len(version) == 14
    
    def test_extract_task_definition_no_config(self):
        """Test task definition extraction from function without config."""
        def simple_task():
            """Simple task without configuration."""
            pass
        
        # Extract definition
        definition = extract_task_definition(simple_task)
        
        # Verify basic metadata
        assert definition["function_name"] == "simple_task"
        assert definition["doc"] == "Simple task without configuration."
        
        # Config fields should not be present if no config
        assert "max_parallel" not in definition
        assert "resource_sizing" not in definition
        assert "timeout" not in definition
        
        # Verify empty parameters
        assert definition["parameters"] == {}
    
    def test_extract_task_definition_complex_signature(self):
        """Test task definition extraction with complex function signature."""
        def complex_task(
            required_param: str,
            optional_param: int = 100,
            *args,
            keyword_only: bool = True,
            **kwargs
        ):
            """Complex task with various parameter types."""
            pass
        
        # Extract definition
        definition = extract_task_definition(complex_task)
        
        # Verify function name and doc
        assert definition["function_name"] == "complex_task"
        assert definition["doc"] == "Complex task with various parameter types."
        
        # Verify parameters
        params = definition["parameters"]
        assert "required_param" in params
        assert "optional_param" in params
        assert "args" in params
        assert "keyword_only" in params
        assert "kwargs" in params
        
        # Verify parameter details
        assert params["required_param"]["annotation"] == "<class 'str'>"
        assert params["required_param"]["default"] is None
        assert params["optional_param"]["default"] == "100"
        assert params["keyword_only"]["default"] == "True"
    
    def test_version_consistency(self):
        """Test that version generation is consistent for same task."""
        def test_task(param: str):
            """Test task for consistency check."""
            pass
        
        # Generate version multiple times
        version1 = get_task_version(test_task)
        version2 = get_task_version(test_task)
        version3 = get_task_version(test_task)
        
        # All should be the same
        assert version1 == version2 == version3
        assert version1.startswith("local-")
    
    def test_version_changes_with_function_changes(self):
        """Test that version changes when function definition changes."""
        def task_v1():
            """Version 1 of the task."""
            pass
        
        def task_v2():
            """Version 2 of the task."""
            pass
        
        # Same function name but different content should produce different versions
        version1 = get_task_version(task_v1)
        version2 = get_task_version(task_v2)
        
        assert version1 != version2
        assert version1.startswith("local-")
        assert version2.startswith("local-") 