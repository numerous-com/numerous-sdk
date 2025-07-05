"""
Tests for the task execution layer functionality.

This tests the integration between the SDK and the new task execution layer
with conflict detection, force operations, and session scoping.
"""

import pytest
import json
import base64
from unittest.mock import Mock, patch, MagicMock
from typing import Dict, Any

from numerous.tasks import task, TaskControl, Session
from numerous.tasks.api_backend import APIConfig, APIConnectedBackend
from numerous.tasks.exceptions import BackendError, TaskExecutionConflictError


class TestTaskExecutionLayerTypes:
    """Test the new execution layer type definitions."""
    
    def test_execution_conflict_error(self):
        """Test ExecutionConflictError exception."""
        # Arrange
        error = TaskExecutionConflictError(
            message="Task instance already has active execution",
            active_execution_id="exec-123",
            conflicting_instance_id="instance-456"
        )
        
        # Act & Assert
        assert str(error) == "Task instance already has active execution"
        assert error.active_execution_id == "exec-123"
        assert error.conflicting_instance_id == "instance-456"
        assert isinstance(error, Exception)


class TestTaskRegistration:
    """Test idempotent task registration with the execution layer."""
    
    def test_upsert_task_success(self):
        """Test successful task registration (upsert)."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token",
            organization_id="org-123"
        )
        
        mock_response = {
            "upsertTask": {
                "id": "task-456",
                "name": "test_task",
                "version": "1.0.0",
                "functionName": "test_task",
                "module": "test_module"
            }
        }
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act
            result = backend.upsert_task(
                name="test_task",
                version="1.0.0",
                function_name="test_task",
                module="test_module"
            )
            
            # Assert
            assert result["id"] == "task-456"
            assert result["name"] == "test_task"
            mock_client._loop.await_coro.assert_called_once()
    
    def test_upsert_task_instance_success(self):
        """Test successful task instance registration (upsert)."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_response = {
            "upsertInstance": {
                "id": "instance-789",
                "apiId": "instance-api-123",
                "sessionId": "session-456",
                "taskDefinitionName": "test_task",
                "status": "PENDING"
            }
        }
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act
            result = backend.upsert_task_instance(
                instance_id="instance-api-123",
                session_id="session-456",
                task_name="test_task",
                task_version="1.0.0",
                inputs={"param": "value"}
            )
            
            # Assert
            assert result["id"] == "instance-789"
            assert result["status"] == "PENDING"
            mock_client._loop.await_coro.assert_called_once()


class TestExecutionConflictDetection:
    """Test execution conflict detection functionality."""
    
    def test_start_execution_success_no_conflict(self):
        """Test starting execution successfully when no conflict exists."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_response = {
            "startExecution": {
                "id": "execution-123",
                "status": "RUNNING",
                "taskInstanceId": "instance-456",
                "startedAt": "2024-01-01T00:00:00Z"
            }
        }
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act
            result = backend.start_execution(
                task_instance_id="instance-456",
                session_id="session-789",
                client_id="client-abc",
                force=False
            )
            
            # Assert
            assert result["id"] == "execution-123"
            assert result["status"] == "RUNNING"
            mock_client._loop.await_coro.assert_called_once()
    
    def test_start_execution_conflict_error(self):
        """Test execution conflict when active execution exists."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        # Mock GraphQL error response for conflict
        mock_client = Mock()
        mock_error = Exception("ExecutionConflictError: Task instance instance-456 already has an active execution")
        mock_client._loop.await_coro.side_effect = mock_error
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act & Assert
            with pytest.raises(TaskExecutionConflictError):
                backend.start_execution(
                    task_instance_id="instance-456",
                    session_id="session-789",
                    client_id="client-abc",
                    force=False
                )
    
    def test_check_execution_conflict_exists(self):
        """Test checking for execution conflicts when one exists."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_response = {
            "checkExecutionConflict": {
                "conflictType": "ACTIVE_EXECUTION",
                "activeExecutionId": "execution-456",
                "instanceId": "instance-123"
            }
        }
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act
            conflict = backend.check_execution_conflict("instance-123")
            
            # Assert
            assert conflict is not None
            assert conflict["conflictType"] == "ACTIVE_EXECUTION"
            assert conflict["activeExecutionId"] == "execution-456"
    
    def test_check_execution_conflict_none(self):
        """Test checking for execution conflicts when none exist."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_response = {"checkExecutionConflict": None}
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act
            conflict = backend.check_execution_conflict("instance-123")
            
            # Assert
            assert conflict is None


class TestForceOperations:
    """Test force execution operations."""
    
    def test_force_start_execution_success(self):
        """Test force starting execution (kills existing executions)."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_response = {
            "startExecution": {
                "id": "execution-new-123",
                "status": "RUNNING",
                "taskInstanceId": "instance-456",
                "startedAt": "2024-01-01T00:00:00Z"
            }
        }
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act
            result = backend.force_start_execution(
                task_instance_id="instance-456",
                session_id="session-789",
                client_id="client-abc"
            )
            
            # Assert
            assert result["id"] == "execution-new-123"
            assert result["status"] == "RUNNING"
            mock_client._loop.await_coro.assert_called_once()
    
    def test_force_stop_execution_success(self):
        """Test force stopping an execution."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_response = {
            "forceStopExecution": {
                "id": "execution-123",
                "status": "KILLED",
                "completedAt": "2024-01-01T00:00:00Z",
                "statusMessage": "Force stopped by user"
            }
        }
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act
            result = backend.force_stop_execution(
                execution_id="execution-123",
                session_id="session-789",
                reason="Force stopped by user"
            )
            
            # Assert
            assert result["id"] == "execution-123"
            assert result["status"] == "KILLED"
            assert result["statusMessage"] == "Force stopped by user"


class TestExecutionLifecycle:
    """Test complete execution lifecycle operations."""
    
    def test_report_execution_progress(self):
        """Test reporting execution progress."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_response = {
            "reportExecutionProgress": {
                "id": "execution-123",
                "progress": 75.0,
                "statusMessage": "Processing data"
            }
        }
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act
            result = backend.report_execution_progress(
                execution_id="execution-123",
                progress=75.0,
                status_message="Processing data"
            )
            
            # Assert
            assert result["progress"] == 75.0
            assert result["statusMessage"] == "Processing data"
    
    def test_complete_execution(self):
        """Test completing an execution with results."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_response = {
            "completeExecution": {
                "id": "execution-123",
                "status": "COMPLETED",
                "result": json.dumps({"output": "processed_data"}),
                "completedAt": "2024-01-01T00:00:00Z"
            }
        }
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act
            result = backend.complete_execution(
                execution_id="execution-123",
                result={"output": "processed_data"}
            )
            
            # Assert
            assert result["status"] == "COMPLETED"
            assert json.loads(result["result"]) == {"output": "processed_data"}
    
    def test_fail_execution(self):
        """Test failing an execution with error details."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_response = {
            "failExecution": {
                "id": "execution-123",
                "status": "FAILED",
                "error": "Task processing failed: Invalid input",
                "completedAt": "2024-01-01T00:00:00Z"
            }
        }
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act
            result = backend.fail_execution(
                execution_id="execution-123",
                error="Task processing failed: Invalid input"
            )
            
            # Assert
            assert result["status"] == "FAILED"
            assert result["error"] == "Task processing failed: Invalid input"


class TestSessionScoping:
    """Test session scoping and client management."""
    
    def test_validate_session_ownership_success(self):
        """Test successful session ownership validation."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_response = {"validateSessionOwnership": True}
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act
            is_valid = backend.validate_session_ownership(
                session_id="session-123",
                instance_id="instance-456"
            )
            
            # Assert
            assert is_valid is True
    
    def test_validate_session_ownership_failure(self):
        """Test session ownership validation failure."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_client = Mock()
        mock_error = Exception("SessionOwnershipError: Instance does not belong to session")
        mock_client._loop.await_coro.side_effect = mock_error
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act & Assert
            with pytest.raises(Exception, match="SessionOwnershipError"):
                backend.validate_session_ownership(
                    session_id="session-123",
                    instance_id="instance-456"
                )
    
    def test_disconnect_client(self):
        """Test disconnecting a client and killing its executions."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_response = {
            "disconnectClient": {
                "killedExecutions": ["execution-1", "execution-2"],
                "message": "Client disconnected successfully"
            }
        }
        
        mock_client = Mock()
        mock_client._loop.await_coro.return_value = mock_response
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act
            result = backend.disconnect_client(
                client_id="client-123",
                session_id="session-456"
            )
            
            # Assert
            assert result["killedExecutions"] == ["execution-1", "execution-2"]
            assert "successfully" in result["message"]


class TestTaskExecutionLayerIntegration:
    """Test integration of execution layer with @task decorator."""
    
    def test_task_execution_with_conflict_detection(self):
        """Test task execution with conflict detection enabled."""
        @task
        def conflict_test_task(data: str):
            return f"processed_{data}"
        
        # Mock execution layer responses
        mock_backend = Mock()
        mock_backend.config.task_instance_id = "instance-123"
        mock_backend.config.session_id = "session-456"
        
        # Mock successful execution start (no conflict)
        mock_backend.start_execution.return_value = {
            "id": "execution-789",
            "status": "RUNNING"
        }
        
        # Mock input fetching
        mock_backend.fetch_task_inputs.return_value = {"data": "test_input"}
        
        with patch('numerous.tasks.api_backend.get_api_backend', return_value=mock_backend):
            with Session() as session:
                result = conflict_test_task("ignored")  # API inputs used instead
                
                assert result == "processed_test_input"
                mock_backend.start_execution.assert_called_once()
                mock_backend.complete_execution.assert_called_once()
    
    def test_task_execution_with_force_mode(self):
        """Test task execution with force mode enabled."""
        @task
        def force_test_task(value: int):
            return value * 2
        
        # Mock execution layer with force mode
        mock_backend = Mock()
        mock_backend.config.task_instance_id = "instance-123"
        mock_backend.config.force_execution = True
        
        mock_backend.force_start_execution.return_value = {
            "id": "execution-new",
            "status": "RUNNING"
        }
        
        mock_backend.fetch_task_inputs.return_value = {"value": 21}
        
        with patch('numerous.tasks.api_backend.get_api_backend', return_value=mock_backend):
            with Session() as session:
                result = force_test_task(10)  # API inputs used instead
                
                assert result == 42  # 21 * 2
                mock_backend.force_start_execution.assert_called_once()
    
    def test_task_execution_error_handling(self):
        """Test error handling in task execution layer."""
        @task
        def error_test_task(should_fail: bool):
            if should_fail:
                raise ValueError("Task intentionally failed")
            return "success"
        
        mock_backend = Mock()
        mock_backend.config.task_instance_id = "instance-123"
        mock_backend.fetch_task_inputs.return_value = {"should_fail": True}
        
        # Mock start_execution to return execution ID for fail_execution
        mock_backend.start_execution.return_value = {
            "id": "execution-error-123",
            "status": "RUNNING"
        }
        
        with patch('numerous.tasks.api_backend.get_api_backend', return_value=mock_backend):
            with Session() as session:
                with pytest.raises(ValueError, match="Task intentionally failed"):
                    error_test_task(False)  # API inputs override
                
                # Verify error was reported to execution layer
                mock_backend.fail_execution.assert_called_once()
                call_args = mock_backend.fail_execution.call_args
                assert "Task intentionally failed" in str(call_args)


class TestRealTimeUpdates:
    """Test real-time execution updates and subscriptions."""
    
    def test_subscribe_to_task_execution_updates(self):
        """Test subscribing to task execution updates."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        # Mock subscription response
        mock_updates = [
            {"id": "execution-123", "status": "RUNNING", "progress": 25.0},
            {"id": "execution-123", "status": "RUNNING", "progress": 75.0},
            {"id": "execution-123", "status": "COMPLETED", "progress": 100.0}
        ]
        
        mock_client = Mock()
        mock_client.subscribe.return_value = iter(mock_updates)
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act
            updates = list(backend.subscribe_to_execution_updates("execution-123"))
            
            # Assert
            assert len(updates) == 3
            assert updates[0]["progress"] == 25.0
            assert updates[1]["progress"] == 75.0
            assert updates[2]["status"] == "COMPLETED"
    
    def test_task_instance_updates_subscription(self):
        """Test subscribing to task instance updates."""
        # Arrange
        config = APIConfig(
            api_url="https://api.test.com",
            access_token="test-token"
        )
        
        mock_updates = [
            {"id": "instance-123", "status": "PENDING"},
            {"id": "instance-123", "status": "RUNNING"},
            {"id": "instance-123", "status": "COMPLETED"}
        ]
        
        mock_client = Mock()
        mock_client.subscribe.return_value = iter(mock_updates)
        
        with patch('numerous.collections._get_client.get_client', return_value=mock_client):
            backend = APIConnectedBackend(config)
            
            # Act
            updates = list(backend.subscribe_to_instance_updates("instance-123"))
            
            # Assert
            assert len(updates) == 3
            assert updates[0]["status"] == "PENDING"
            assert updates[2]["status"] == "COMPLETED" 