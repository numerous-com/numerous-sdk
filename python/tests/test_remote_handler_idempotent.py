"""
Unit tests for RemoteTaskControlHandler with idempotent operations.

Tests the integration of idempotent operations with the remote handler using mocks.
"""

import pytest
from unittest.mock import Mock, AsyncMock, patch
import time

from numerous.tasks.control import TaskControl
from numerous.tasks.integration.remote_handler import RemoteTaskControlHandler


class TestRemoteTaskControlHandlerIdempotent:
    """Test RemoteTaskControlHandler with idempotent operations."""
    
    @pytest.fixture
    def mock_client(self):
        """Mock GraphQL client."""
        return Mock()
    
    @pytest.fixture
    def mock_task_control(self):
        """Mock TaskControl instance."""
        task_control = Mock(spec=TaskControl)
        task_control.task_definition_name = "test_task"
        task_control.instance_id = "test_instance_123"
        return task_control
    
    @pytest.fixture
    def handler(self, mock_client):
        """RemoteTaskControlHandler with mocked client."""
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            handler = RemoteTaskControlHandler(session_id="test_session_123")
            # Mock the idempotent operations
            handler._idempotent_ops.upsert_task_definition = AsyncMock()
            handler._idempotent_ops.upsert_task_instance = AsyncMock()
            return handler
    
    def test_initialization(self, handler):
        """Test handler initialization with session ID."""
        assert handler.session_id == "test_session_123"
        assert handler.is_connected
        assert handler._idempotent_ops is not None
        assert handler._registered_tasks == set()
    
    def test_initialization_auto_session(self):
        """Test handler initialization with auto-generated session ID."""
        with patch('numerous.collections._get_client.get_client', return_value=Mock()):
            handler = RemoteTaskControlHandler()
            assert handler.session_id.startswith("session_")
            assert handler.is_connected
    
    @pytest.mark.asyncio
    async def test_ensure_task_registered(self, handler, mock_task_control):
        """Test task registration with idempotent operations."""
        # Mock task function
        def sample_task():
            """Sample task for testing."""
            pass
        
        mock_task_control._task_func = sample_task
        mock_task_control._explicit_version = None
        
        # Mock the upsert result
        handler._idempotent_ops.upsert_task_definition.return_value = {
            "task_name": "test_task",
            "version": "local-abc12345",
            "status": "registered",
            "created": True,
            "timestamp": time.time()
        }
        
        # Test registration
        await handler.ensure_task_registered(mock_task_control)
        
        # Verify task was registered
        assert "test_task" in handler._registered_tasks
        handler._idempotent_ops.upsert_task_definition.assert_called_once()
        
        # Test that second call is skipped
        handler._idempotent_ops.upsert_task_definition.reset_mock()
        await handler.ensure_task_registered(mock_task_control)
        handler._idempotent_ops.upsert_task_definition.assert_not_called()
    
    @pytest.mark.asyncio
    async def test_ensure_instance_registered(self, handler, mock_task_control):
        """Test instance registration with idempotent operations."""
        from numerous.tasks.integration.idempotent_operations import TaskInstance
        
        # Mock task function
        def sample_task():
            """Sample task for testing."""
            pass
        
        mock_task_control._task_func = sample_task
        mock_task_control._explicit_version = None
        
        # Mock the upsert result
        mock_instance = TaskInstance(
            id="test_session_123:test_task:1234567890",
            session_id="test_session_123",
            task_name="test_task",
            task_version="local-abc12345",
            status="PENDING",
            created_at=time.time()
        )
        handler._idempotent_ops.upsert_task_instance.return_value = mock_instance
        
        # Test registration
        instance_id = await handler.ensure_instance_registered(mock_task_control)
        
        # Verify instance was registered
        assert instance_id == mock_instance.id
        assert instance_id.startswith("test_session_123:test_task:")
        handler._idempotent_ops.upsert_task_instance.assert_called_once()
    
    def test_log_with_registration(self, handler, mock_task_control):
        """Test log method with automatic task and instance registration."""
        # Mock task function
        def sample_task():
            """Sample task for testing."""
            pass
        
        mock_task_control._task_func = sample_task
        mock_task_control._explicit_version = None
        
        # Mock async operations
        handler._idempotent_ops.upsert_task_definition = AsyncMock(return_value={
            "task_name": "test_task",
            "version": "local-abc12345",
            "status": "registered"
        })
        
        from numerous.tasks.integration.idempotent_operations import TaskInstance
        mock_instance = TaskInstance(
            id="test_session_123:test_task:1234567890",
            session_id="test_session_123",
            task_name="test_task",
            task_version="local-abc12345",
            status="PENDING",
            created_at=time.time()
        )
        handler._idempotent_ops.upsert_task_instance = AsyncMock(return_value=mock_instance)
        
        # Test log method
        handler.log(mock_task_control, "Test message", "info")
        
        # Verify registration was called
        handler._idempotent_ops.upsert_task_definition.assert_called_once()
        handler._idempotent_ops.upsert_task_instance.assert_called_once()
        
        # Verify instance ID was cached
        assert hasattr(mock_task_control, '_remote_instance_id')
        assert mock_task_control._remote_instance_id == mock_instance.id
    
    def test_update_progress_with_registration(self, handler, mock_task_control):
        """Test update_progress method with automatic registration."""
        # Mock task function
        def sample_task():
            """Sample task for testing."""
            pass
        
        mock_task_control._task_func = sample_task
        mock_task_control._explicit_version = None
        
        # Mock async operations
        handler._idempotent_ops.upsert_task_definition = AsyncMock(return_value={
            "task_name": "test_task",
            "version": "local-abc12345",
            "status": "registered"
        })
        
        from numerous.tasks.integration.idempotent_operations import TaskInstance
        mock_instance = TaskInstance(
            id="test_session_123:test_task:1234567890",
            session_id="test_session_123",
            task_name="test_task",
            task_version="local-abc12345",
            status="PENDING",
            created_at=time.time()
        )
        handler._idempotent_ops.upsert_task_instance = AsyncMock(return_value=mock_instance)
        
        # Test update_progress method
        handler.update_progress(mock_task_control, 0.5, "Processing")
        
        # Verify registration was called
        handler._idempotent_ops.upsert_task_definition.assert_called_once()
        handler._idempotent_ops.upsert_task_instance.assert_called_once()
        
        # Verify instance ID was cached
        assert hasattr(mock_task_control, '_remote_instance_id')
        assert mock_task_control._remote_instance_id == mock_instance.id
    
    def test_cached_instance_id_reuse(self, handler, mock_task_control):
        """Test that cached instance ID is reused for subsequent calls."""
        # Pre-set cached instance ID
        mock_task_control._remote_instance_id = "cached_instance_id"
        
        # Mock async operations (should not be called)
        handler._idempotent_ops.upsert_task_definition = AsyncMock()
        handler._idempotent_ops.upsert_task_instance = AsyncMock()
        
        # Test log method
        handler.log(mock_task_control, "Test message", "info")
        
        # Verify registration was still called for task (but not instance)
        handler._idempotent_ops.upsert_task_definition.assert_called_once()
        handler._idempotent_ops.upsert_task_instance.assert_not_called()
    
    def test_fallback_version_generation(self, handler, mock_task_control):
        """Test fallback version generation when task function is not available."""
        # Don't set _task_func attribute to simulate fallback scenario
        
        # Mock async operations
        handler._idempotent_ops.upsert_task_definition = AsyncMock(return_value={
            "task_name": "test_task",
            "version": "local-12345678",
            "status": "registered"
        })
        
        from numerous.tasks.integration.idempotent_operations import TaskInstance
        mock_instance = TaskInstance(
            id="test_session_123:test_task:1234567890",
            session_id="test_session_123",
            task_name="test_task",
            task_version="local-12345678",
            status="PENDING",
            created_at=time.time()
        )
        handler._idempotent_ops.upsert_task_instance = AsyncMock(return_value=mock_instance)
        
        # Test log method with fallback
        handler.log(mock_task_control, "Test message", "info")
        
        # Verify registration was called with fallback version
        handler._idempotent_ops.upsert_task_definition.assert_called_once()
        call_args = handler._idempotent_ops.upsert_task_definition.call_args
        task_definition = call_args[0][1]  # Second argument is TaskDefinition
        assert task_definition.version.startswith("local-")
        assert task_definition.function_name == "test_task"
        assert task_definition.module == "unknown" 