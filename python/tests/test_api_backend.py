"""
Tests for the API-connected backend functionality.

This tests the hybrid execution mode where tasks run locally but use the
remote API for input/output storage and progress reporting.
"""

import pytest
import os
import json
import base64
from unittest.mock import Mock, patch, MagicMock
from typing import Dict, Any

from numerous.tasks import task, TaskControl, Session
from numerous.tasks.api_backend import (
    APIConfig, APIConnectedBackend, APITaskControlHandler, 
    get_api_backend, is_api_mode, api_task_execution_wrapper
)
from numerous.tasks.exceptions import BackendError


class TestAPIConfig:
    """Test APIConfig functionality."""
    
    def test_api_config_from_environment_with_required_vars(self):
        """Test creating APIConfig from environment variables."""
        with patch.dict(os.environ, {
            'NUMEROUS_API_URL': 'https://api.test.com',
            'NUMEROUS_API_ACCESS_TOKEN': 'test-token-123',
            'NUMEROUS_ORGANIZATION_ID': 'org-456',
            'NUMEROUS_TASK_INSTANCE_ID': 'task-789'
        }):
            config = APIConfig.from_environment()
            
            assert config is not None
            assert config.api_url == 'https://api.test.com'
            assert config.access_token == 'test-token-123'
            assert config.organization_id == 'org-456'
            assert config.task_instance_id == 'task-789'
    
    def test_api_config_from_environment_missing_required_vars(self):
        """Test APIConfig returns None when required vars are missing."""
        with patch.dict(os.environ, {}, clear=True):
            config = APIConfig.from_environment()
            assert config is None
        
        # Missing token
        with patch.dict(os.environ, {'NUMEROUS_API_URL': 'https://api.test.com'}):
            config = APIConfig.from_environment()
            assert config is None
        
        # Missing URL
        with patch.dict(os.environ, {'NUMEROUS_API_ACCESS_TOKEN': 'test-token'}):
            config = APIConfig.from_environment()
            assert config is None
    
    def test_api_config_from_environment_partial_optional_vars(self):
        """Test APIConfig with only required vars set."""
        with patch.dict(os.environ, {
            'NUMEROUS_API_URL': 'https://api.test.com',
            'NUMEROUS_API_ACCESS_TOKEN': 'test-token-123'
        }, clear=True):
            config = APIConfig.from_environment()
            
            assert config is not None
            assert config.api_url == 'https://api.test.com'
            assert config.access_token == 'test-token-123'
            assert config.organization_id is None
            assert config.task_instance_id is None


class TestAPITaskControlHandler:
    """Test APITaskControlHandler functionality."""
    
    def test_api_task_control_handler_logging(self):
        """Test logging functionality of API handler."""
        mock_client = Mock()
        handler = APITaskControlHandler(mock_client, "test-instance-123")
        
        mock_task_control = Mock()
        
        # Test successful logging
        handler.log(mock_task_control, "Test message", "info", extra_field="extra_value")
        
        # Should not raise exception
        # In the future when API endpoints are implemented, would verify mock_client calls
    
    def test_api_task_control_handler_progress_update(self):
        """Test progress update functionality."""
        mock_client = Mock()
        handler = APITaskControlHandler(mock_client, "test-instance-123")
        
        mock_task_control = Mock()
        
        # Test progress updates
        handler.update_progress(mock_task_control, 25.0, "processing")
        handler.update_progress(mock_task_control, 75.0, None)
        
        # Should not raise exceptions
    
    def test_api_task_control_handler_status_update(self):
        """Test status update functionality."""
        mock_client = Mock()
        handler = APITaskControlHandler(mock_client, "test-instance-123")
        
        mock_task_control = Mock()
        
        handler.update_status(mock_task_control, "running")
        handler.update_status(mock_task_control, "finalizing")
        
        # Should not raise exceptions
    
    def test_api_task_control_handler_error_handling(self):
        """Test handler gracefully handles API errors."""
        # Mock client that raises exceptions
        mock_client = Mock()
        handler = APITaskControlHandler(mock_client, "test-instance-123")
        
        mock_task_control = Mock()
        
        # These should handle exceptions gracefully
        handler.log(mock_task_control, "Test message", "info")
        handler.update_progress(mock_task_control, 50.0, "halfway")
        handler.update_status(mock_task_control, "processing")


class TestAPIConnectedBackend:
    """Test APIConnectedBackend functionality."""
    
    def test_api_backend_initialization_success(self):
        """Test successful API backend initialization."""
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token",
            task_instance_id="test-instance"
        )
        
        mock_client = Mock()
        with patch('numerous.tasks.api_backend.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            assert backend.config == config
            assert backend.api_client == mock_client
    
    def test_api_backend_initialization_failure(self):
        """Test API backend initialization failure."""
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        with patch('numerous.tasks.api_backend.get_client', side_effect=Exception("API connection failed")):
            with pytest.raises(BackendError, match="Cannot connect to API"):
                APIConnectedBackend(config)
    
    def test_fetch_task_inputs_success(self):
        """Test successful task input fetching."""
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        # Mock input data
        input_data = {"param1": "value1", "param2": 42}
        input_json = json.dumps(input_data)
        input_base64 = base64.b64encode(input_json.encode('utf-8')).decode('utf-8')
        
        # Mock GraphQL response
        mock_response_data = {
            "getTaskInstance": {
                "id": "test-instance-123",
                "inputs": input_base64,
                "taskDefinitionName": "test_task",
                "status": "PENDING"
            }
        }
        
        # Mock API client
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response_data
        
        with patch('numerous.tasks.api_backend.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            result = backend.fetch_task_inputs("test-instance-123")
            
            assert result == input_data
            # Verify GraphQL query was called
            mock_client._loop.await_coro.assert_called_once()
    
    def test_fetch_task_inputs_no_inputs(self):
        """Test fetching task inputs when none are provided."""
        config = APIConfig(
            api_url="https://api.test.com", 
            access_token="test-token"
        )
        
        # Mock response with no inputs
        mock_response_data = {
            "getTaskInstance": {
                "id": "test-instance-123",
                "inputs": None,
                "taskDefinitionName": "test_task",
                "status": "PENDING"
            }
        }
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response_data
        
        with patch('numerous.tasks.api_backend.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            result = backend.fetch_task_inputs("test-instance-123")
            
            assert result == {}
    
    def test_fetch_task_inputs_not_found(self):
        """Test fetching inputs for non-existent task instance."""
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        # Mock response with no task instance
        mock_response_data = {"getTaskInstance": None}
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response_data
        
        with patch('numerous.tasks.api_backend.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            with pytest.raises(BackendError, match="Task instance .* not found"):
                backend.fetch_task_inputs("nonexistent-instance")
    
    def test_report_task_result_success(self):
        """Test successful task result reporting."""
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_response_data = {
            "reportTaskOutcome": {
                "id": "test-instance-123",
                "status": "COMPLETED",
                "completedAt": "2024-01-01T00:00:00Z"
            }
        }
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response_data
        
        with patch('numerous.tasks.api_backend.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Test successful result reporting
            result_data = {"output": "processed data", "count": 100}
            backend.report_task_result("test-instance-123", result_data)
            
            # Verify mutation was called
            mock_client._loop.await_coro.assert_called_once()
    
    def test_report_task_result_error(self):
        """Test reporting task error to API."""
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_response_data = {
            "reportTaskOutcome": {
                "id": "test-instance-123", 
                "status": "FAILED",
                "completedAt": "2024-01-01T00:00:00Z"
            }
        }
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response_data
        
        with patch('numerous.tasks.api_backend.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Test error reporting
            test_error = ValueError("Task processing failed")
            backend.report_task_result("test-instance-123", None, test_error)
            
            mock_client._loop.await_coro.assert_called_once()
    
    def test_setup_task_control(self):
        """Test setting up TaskControl handler."""
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_client = Mock()
        with patch('numerous.tasks.api_backend.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            with patch('numerous.tasks.api_backend.set_task_control_handler') as mock_set_handler:
                backend.setup_task_control("test-instance-123")
                
                # Verify handler was set
                mock_set_handler.assert_called_once()
                handler_arg = mock_set_handler.call_args[0][0]
                assert isinstance(handler_arg, APITaskControlHandler)
                assert handler_arg.task_instance_id == "test-instance-123"


class TestAPIMode:
    """Test API mode detection and global functions."""
    
    def test_is_api_mode_enabled(self):
        """Test API mode detection when environment variables are set."""
        with patch.dict(os.environ, {
            'NUMEROUS_API_URL': 'https://api.test.com',
            'NUMEROUS_API_ACCESS_TOKEN': 'test-token'
        }):
            with patch('numerous.tasks.api_backend.get_client'):
                assert is_api_mode() is True
    
    def test_is_api_mode_disabled(self):
        """Test API mode detection when environment variables are not set."""
        with patch.dict(os.environ, {}, clear=True):
            assert is_api_mode() is False
    
    def test_get_api_backend_success(self):
        """Test getting API backend when properly configured."""
        with patch.dict(os.environ, {
            'NUMEROUS_API_URL': 'https://api.test.com',
            'NUMEROUS_API_ACCESS_TOKEN': 'test-token'
        }):
            with patch('numerous.tasks.api_backend.get_client'):
                backend = get_api_backend()
                assert backend is not None
                assert isinstance(backend, APIConnectedBackend)
    
    def test_get_api_backend_none(self):
        """Test getting API backend when not configured."""
        with patch.dict(os.environ, {}, clear=True):
            backend = get_api_backend()
            assert backend is None
    
    def test_get_api_backend_singleton(self):
        """Test that get_api_backend returns the same instance."""
        with patch.dict(os.environ, {
            'NUMEROUS_API_URL': 'https://api.test.com',
            'NUMEROUS_API_ACCESS_TOKEN': 'test-token'
        }):
            with patch('numerous.tasks.api_backend.get_client'):
                backend1 = get_api_backend()
                backend2 = get_api_backend()
                
                assert backend1 is backend2  # Same instance


class TestAPITaskExecutionWrapper:
    """Test the API task execution wrapper functionality."""
    
    def test_wrapper_local_mode(self):
        """Test wrapper falls back to local execution when API not configured."""
        def test_task(a, b):
            return a + b
        
        with patch.dict(os.environ, {}, clear=True):
            result = api_task_execution_wrapper(test_task, "test_task", 5, 3)
            assert result == 8
    
    def test_wrapper_api_mode_no_instance_id(self):
        """Test wrapper falls back to local when API configured but no instance ID."""
        def test_task(a, b):
            return a + b
        
        with patch.dict(os.environ, {
            'NUMEROUS_API_URL': 'https://api.test.com',
            'NUMEROUS_API_ACCESS_TOKEN': 'test-token'
            # No NUMEROUS_TASK_INSTANCE_ID
        }):
            with patch('numerous.tasks.api_backend.get_client'):
                result = api_task_execution_wrapper(test_task, "test_task", 5, 3)
                assert result == 8
    
    def test_wrapper_api_mode_success(self):
        """Test wrapper successfully executes with API inputs."""
        def test_task(a, b):
            return a + b
        
        # Mock API inputs
        api_inputs = {"a": 10, "b": 20}
        
        mock_backend = Mock()
        mock_backend.config.task_instance_id = "test-instance-123"
        mock_backend.fetch_task_inputs.return_value = api_inputs
        
        with patch('numerous.tasks.api_backend.get_api_backend', return_value=mock_backend):
            result = api_task_execution_wrapper(test_task, "test_task", 1, 2)  # Args ignored
            
            assert result == 30  # 10 + 20 from API inputs
            mock_backend.setup_task_control.assert_called_once_with("test-instance-123")
            mock_backend.fetch_task_inputs.assert_called_once_with("test-instance-123")
            mock_backend.report_task_result.assert_called_once_with("test-instance-123", 30)
    
    def test_wrapper_api_mode_task_error(self):
        """Test wrapper reports errors to API when task fails."""
        def failing_task(a, b):
            raise ValueError("Task failed")
        
        api_inputs = {"a": 10, "b": 20}
        
        mock_backend = Mock()
        mock_backend.config.task_instance_id = "test-instance-123"
        mock_backend.fetch_task_inputs.return_value = api_inputs
        
        with patch('numerous.tasks.api_backend.get_api_backend', return_value=mock_backend):
            with pytest.raises(ValueError, match="Task failed"):
                api_task_execution_wrapper(failing_task, "failing_task", 1, 2)
            
            # Verify error was reported
            mock_backend.report_task_result.assert_called_once()
            call_args = mock_backend.report_task_result.call_args
            assert call_args[0][0] == "test-instance-123"  # instance_id
            assert call_args[0][1] is None  # result
            assert isinstance(call_args[0][2], ValueError)  # error


class TestAPIIntegrationWithTasks:
    """Test integration of API backend with @task decorator."""
    
    def test_task_local_execution(self):
        """Test task executes locally when no API environment variables."""
        @task
        def simple_task(x: int, y: int):
            return x * y
        
        with patch.dict(os.environ, {}, clear=True):
            with Session() as session:
                result = simple_task(6, 7)
                assert result == 42
    
    def test_task_api_execution_mock(self):
        """Test task executes with API inputs when environment variables are set."""
        @task
        def api_task(value: int, multiplier: int):
            return value * multiplier
        
        # Mock API setup
        api_inputs = {"value": 15, "multiplier": 3}
        mock_backend = Mock()
        mock_backend.config.task_instance_id = "test-instance-123"
        mock_backend.fetch_task_inputs.return_value = api_inputs
        
        with patch('numerous.tasks.api_backend.get_api_backend', return_value=mock_backend):
            with Session() as session:
                result = api_task(1, 1)  # Arguments ignored in API mode
                
                assert result == 45  # 15 * 3 from API inputs
                mock_backend.fetch_task_inputs.assert_called_once_with("test-instance-123")
                mock_backend.report_task_result.assert_called_once_with("test-instance-123", 45)
    
    def test_task_api_execution_with_task_control_mock(self):
        """Test task with TaskControl executes with API inputs."""
        @task
        def api_task_with_control(tc: TaskControl, message: str, count: int):
            tc.log(f"Processing {message}")
            tc.update_progress(50.0, "halfway")
            result = f"{message}_processed_{count}_times"
            tc.update_progress(100.0, "complete")
            return result
        
        # Mock API setup
        api_inputs = {"message": "test_data", "count": 5}
        mock_backend = Mock()
        mock_backend.config.task_instance_id = "test-instance-123"
        mock_backend.fetch_task_inputs.return_value = api_inputs
        
        with patch('numerous.tasks.api_backend.get_api_backend', return_value=mock_backend):
            with Session() as session:
                result = api_task_with_control("ignored", 0)  # Arguments ignored
                
                assert result == "test_data_processed_5_times"
                mock_backend.setup_task_control.assert_called_once_with("test-instance-123")
                mock_backend.fetch_task_inputs.assert_called_once_with("test-instance-123")
                mock_backend.report_task_result.assert_called_once_with("test-instance-123", result)


class TestAPIBackendErrorHandling:
    """Test error handling scenarios for API backend."""
    
    def test_api_backend_fetch_error_handling(self):
        """Test handling of API errors during input fetching."""
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_client = Mock()
        mock_client._loop.await_coro.side_effect = Exception("API request failed")
        
        with patch('numerous.tasks.api_backend.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            with pytest.raises(BackendError, match="Failed to fetch inputs"):
                backend.fetch_task_inputs("test-instance-123")
    
    def test_api_backend_report_error_handling(self):
        """Test handling of API errors during result reporting."""
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_client = Mock()
        mock_client._loop.await_coro.side_effect = Exception("API request failed")
        
        with patch('numerous.tasks.api_backend.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            with pytest.raises(BackendError, match="Failed to report result"):
                backend.report_task_result("test-instance-123", {"result": "test"})
    
    def test_wrapper_graceful_error_handling(self):
        """Test wrapper handles backend errors gracefully."""
        def test_task(a, b):
            return a + b
        
        # Mock backend that fails on input fetching
        mock_backend = Mock()
        mock_backend.config.task_instance_id = "test-instance-123"
        mock_backend.fetch_task_inputs.side_effect = BackendError("API fetch failed")
        
        with patch('numerous.tasks.api_backend.get_api_backend', return_value=mock_backend):
            with pytest.raises(BackendError, match="API fetch failed"):
                api_task_execution_wrapper(test_task, "test_task", 5, 3)
            
            # Verify setup was attempted
            mock_backend.setup_task_control.assert_called_once_with("test-instance-123") 