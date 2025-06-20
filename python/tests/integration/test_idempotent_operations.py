"""
Integration tests for idempotent task and instance registration operations.

Tests the upsert operations for task definitions and instances with local versioning.
Requires API to be running at localhost:8080.
"""

import pytest
import time
from unittest.mock import Mock, AsyncMock

from numerous.tasks.integration.idempotent_operations import (
    IdempotentOperations, 
    TaskDefinition, 
    TaskInstance
)
from numerous.tasks.versioning import generate_local_version, extract_task_definition


class TestIdempotentOperations:
    """Test idempotent operations for task and instance registration."""
    
    @pytest.fixture
    def mock_client(self):
        """Mock GraphQL client for testing."""
        return Mock()
    
    @pytest.fixture
    def idempotent_ops(self, mock_client):
        """IdempotentOperations instance with mock client."""
        return IdempotentOperations(mock_client)
    
    def test_generate_instance_id(self, idempotent_ops):
        """Test instance ID generation with session scoping."""
        session_id = "test_session_123"
        task_name = "data_processor"
        
        # Generate instance ID
        instance_id = idempotent_ops.generate_instance_id(session_id, task_name)
        
        # Verify format
        assert instance_id.startswith(f"{session_id}:{task_name}:")
        assert len(instance_id.split(':')) == 3
        
        # Verify timestamp component
        timestamp_part = instance_id.split(':')[2]
        assert timestamp_part.isdigit()
        assert int(timestamp_part) > 0
    
    def test_validate_session_ownership(self, idempotent_ops):
        """Test session ownership validation."""
        session_id = "test_session_123"
        task_name = "data_processor"
        
        # Generate valid instance ID
        valid_instance_id = idempotent_ops.generate_instance_id(session_id, task_name)
        
        # Test valid ownership
        assert idempotent_ops.validate_session_ownership(valid_instance_id, session_id)
        
        # Test invalid ownership
        invalid_instance_id = "other_session:data_processor:1234567890"
        assert not idempotent_ops.validate_session_ownership(invalid_instance_id, session_id)
    
    @pytest.mark.asyncio
    async def test_upsert_task_definition(self, idempotent_ops):
        """Test task definition upsert operation."""
        task_name = "test_task"
        task_definition = TaskDefinition(
            name=task_name,
            version="local-abc12345",
            function_name="test_function",
            module="test_module",
            parameters={"param1": {"annotation": "str", "default": None}},
            metadata={"doc": "Test task", "max_parallel": 2}
        )
        
        # Test upsert operation
        result = await idempotent_ops.upsert_task_definition(task_name, task_definition)
        
        # Verify result structure
        assert result["task_name"] == task_name
        assert result["version"] == task_definition.version
        assert result["status"] == "registered"
        assert "timestamp" in result
        assert isinstance(result["created"], bool)
    
    @pytest.mark.asyncio
    async def test_upsert_task_instance(self, idempotent_ops):
        """Test task instance upsert operation."""
        session_id = "test_session_123"
        task_name = "test_task"
        task_version = "local-abc12345"
        
        # Generate instance ID
        instance_id = idempotent_ops.generate_instance_id(session_id, task_name)
        
        # Test upsert operation
        instance = await idempotent_ops.upsert_task_instance(
            instance_id=instance_id,
            session_id=session_id,
            task_name=task_name,
            task_version=task_version
        )
        
        # Verify instance structure
        assert instance.id == instance_id
        assert instance.session_id == session_id
        assert instance.task_name == task_name
        assert instance.task_version == task_version
        assert instance.status == "PENDING"
        assert instance.created_at > 0


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


@pytest.mark.integration
class TestIdempotentOperationsIntegration:
    """Integration tests requiring actual API connection."""
    
    @pytest.fixture
    def real_client(self):
        """Real GraphQL client for integration testing."""
        # This would use the actual organization client
        # For now, skip if no real API available
        pytest.skip("Integration test requires API at localhost:8080")
    
    @pytest.fixture
    def real_idempotent_ops(self, real_client):
        """IdempotentOperations with real client."""
        return IdempotentOperations(real_client)
    
    @pytest.mark.asyncio
    async def test_real_task_registration(self, real_idempotent_ops):
        """Test actual task registration with API."""
        # This test would run against real API
        # Implementation depends on actual GraphQL mutations being available
        pass
    
    @pytest.mark.asyncio
    async def test_real_instance_registration(self, real_idempotent_ops):
        """Test actual instance registration with API."""
        # This test would run against real API
        # Implementation depends on actual GraphQL mutations being available
        pass


class TestSessionScoping:
    """Test session scoping functionality."""
    
    def test_session_scoped_instance_ids(self):
        """Test that different sessions generate different instance IDs."""
        ops = IdempotentOperations(Mock())
        
        session1 = "session_1"
        session2 = "session_2"
        task_name = "same_task"
        
        # Generate IDs for different sessions
        id1 = ops.generate_instance_id(session1, task_name)
        id2 = ops.generate_instance_id(session2, task_name)
        
        # Verify they're different
        assert id1 != id2
        assert id1.startswith(f"{session1}:")
        assert id2.startswith(f"{session2}:")
        
        # Verify session ownership
        assert ops.validate_session_ownership(id1, session1)
        assert not ops.validate_session_ownership(id1, session2)
        assert ops.validate_session_ownership(id2, session2)
        assert not ops.validate_session_ownership(id2, session1)
    
    def test_predictable_instance_ids_within_session(self):
        """Test that instance IDs are predictable within a session."""
        ops = IdempotentOperations(Mock())
        
        session_id = "test_session"
        task_name = "test_task"
        
        # Generate multiple IDs
        id1 = ops.generate_instance_id(session_id, task_name)
        time.sleep(0.001)  # Small delay to ensure different timestamps
        id2 = ops.generate_instance_id(session_id, task_name)
        
        # Verify they follow the expected pattern
        assert id1.startswith(f"{session_id}:{task_name}:")
        assert id2.startswith(f"{session_id}:{task_name}:")
        
        # Verify they're different (due to timestamp)
        assert id1 != id2
        
        # Verify both belong to the same session
        assert ops.validate_session_ownership(id1, session_id)
        assert ops.validate_session_ownership(id2, session_id) 