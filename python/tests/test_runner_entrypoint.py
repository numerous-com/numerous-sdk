import pytest
import os
import sys
import json
import tempfile
from unittest.mock import Mock, patch, MagicMock, mock_open
from pathlib import Path
from typing import Dict, Any

from numerous.tasks.runner_entrypoint import (
    main, RunnerError, _load_manifest, _find_task_details_in_manifest,
    _get_env_var, _fetch_task_inputs, _report_task_outcome
)
from numerous.tasks import task, TaskControl, Session


class TestRunnerEntrypoint:
    """Test the runner entrypoint functionality."""

    def test_runner_error_exception(self):
        """Test RunnerError exception class."""
        error = RunnerError("Test error message")
        assert str(error) == "Test error message"
        assert isinstance(error, Exception)

    @patch('tomli.load')
    @patch('builtins.open', new_callable=mock_open, read_data=b'test data')
    @patch('pathlib.Path.is_file')
    def test_load_manifest_success(self, mock_is_file, mock_file, mock_tomli_load):
        """Test successful manifest loading."""
        mock_is_file.return_value = True
        mock_manifest_data = {"task": [{"function_name": "test_task"}]}
        mock_tomli_load.return_value = mock_manifest_data
        
        result = _load_manifest("/path/to/manifest.toml")
        assert result == mock_manifest_data
        mock_tomli_load.assert_called_once()

    @patch('pathlib.Path.is_file')
    def test_load_manifest_file_not_found(self, mock_is_file):
        """Test manifest loading when file doesn't exist."""
        mock_is_file.return_value = False
        
        with pytest.raises(RunnerError, match="Manifest file not found"):
            _load_manifest("/nonexistent/manifest.toml")

    @patch('pathlib.Path.is_file')
    def test_load_manifest_tomli_not_found(self, mock_is_file):
        """Test manifest loading when tomli is not available."""
        mock_is_file.return_value = True
        
        # Use patch.dict to temporarily remove tomli module and mock the import
        with patch.dict('sys.modules', {'tomli': None}):
            with pytest.raises(RunnerError, match="TOML parsing library.*not available"):
                _load_manifest("/path/to/manifest.toml")

    @patch('tomli.load')
    @patch('builtins.open', new_callable=mock_open, read_data=b'test data')
    @patch('pathlib.Path.is_file')
    def test_load_manifest_parse_error(self, mock_is_file, mock_file, mock_tomli_load):
        """Test manifest loading with parsing error."""
        mock_is_file.return_value = True
        mock_tomli_load.side_effect = Exception("Invalid TOML")
        
        with pytest.raises(RunnerError, match="Error parsing manifest"):
            _load_manifest("/path/to/manifest.toml")

    def test_find_task_details_in_manifest_success(self):
        """Test successful task details finding."""
        manifest_data = {
            "task": [
                {"function_name": "task1", "source_file": "task1.py"},
                {"function_name": "task2", "source_file": "task2.py"}
            ]
        }
        
        result = _find_task_details_in_manifest(manifest_data, "task2")
        assert result == {"function_name": "task2", "source_file": "task2.py"}

    def test_find_task_details_in_manifest_not_found(self):
        """Test task details finding when task doesn't exist."""
        manifest_data = {
            "task": [
                {"function_name": "task1", "source_file": "task1.py"}
            ]
        }
        
        with pytest.raises(RunnerError, match="Task function 'nonexistent' not found"):
            _find_task_details_in_manifest(manifest_data, "nonexistent")

    def test_get_env_var_success(self):
        """Test successful environment variable retrieval."""
        with patch.dict(os.environ, {"TEST_VAR": "test_value"}):
            result = _get_env_var("TEST_VAR")
            assert result == "test_value"

    def test_get_env_var_mandatory_missing(self):
        """Test mandatory environment variable missing."""
        with patch.dict(os.environ, {}, clear=True):
            with pytest.raises(RunnerError, match="Mandatory environment variable TEST_VAR not set"):
                _get_env_var("TEST_VAR", is_mandatory=True)

    def test_get_env_var_optional_missing_with_default(self):
        """Test optional environment variable with default."""
        with patch.dict(os.environ, {}, clear=True):
            result = _get_env_var("TEST_VAR", is_mandatory=False, default="default_value")
            assert result == "default_value"

    def test_get_env_var_optional_missing_no_default(self):
        """Test optional environment variable without default."""
        with patch.dict(os.environ, {}, clear=True):
            result = _get_env_var("TEST_VAR", is_mandatory=False)
            assert result is None

    def test_fetch_task_inputs_success(self):
        """Test successful task input fetching."""
        mock_client = Mock()
        mock_gql = Mock()
        mock_loop = Mock()
        
        # Mock the GraphQL response
        mock_response_data = {
            "getTaskInstance": {
                "id": "test_id",
                "inputs": "eyJ0ZXN0X2lucHV0IjogInRlc3RfdmFsdWUifQ==",  # base64 encoded {"test_input": "test_value"}
                "taskDefinitionName": "test_task",
                "status": "RUNNING"
            }
        }
        
        mock_client._gql = mock_gql
        mock_client._loop = mock_loop
        mock_client._headers = {}
        mock_loop.await_coro.return_value = mock_response_data
        mock_gql.get_data.return_value = mock_response_data
        
        result = _fetch_task_inputs(mock_client, "test_id")
        assert result == {"test_input": "test_value"}

    def test_fetch_task_inputs_no_instance(self):
        """Test task input fetching when instance not found."""
        mock_client = Mock()
        mock_gql = Mock()
        mock_loop = Mock()
        
        mock_client._gql = mock_gql
        mock_client._loop = mock_loop
        mock_client._headers = {}
        mock_loop.await_coro.return_value = {"getTaskInstance": None}
        mock_gql.get_data.return_value = {"getTaskInstance": None}
        
        with pytest.raises(RuntimeError, match="Task instance test_id not found"):
            _fetch_task_inputs(mock_client, "test_id")

    def test_fetch_task_inputs_no_inputs(self):
        """Test task input fetching when no inputs provided."""
        mock_client = Mock()
        mock_gql = Mock()
        mock_loop = Mock()
        
        mock_response_data = {
            "getTaskInstance": {
                "id": "test_id",
                "inputs": None,
                "taskDefinitionName": "test_task",
                "status": "RUNNING"
            }
        }
        
        mock_client._gql = mock_gql
        mock_client._loop = mock_loop
        mock_client._headers = {}
        mock_loop.await_coro.return_value = mock_response_data
        mock_gql.get_data.return_value = mock_response_data
        
        result = _fetch_task_inputs(mock_client, "test_id")
        assert result == {}

    def test_report_task_outcome_success(self):
        """Test successful task outcome reporting."""
        mock_client = Mock()
        mock_gql = Mock()
        mock_loop = Mock()
        
        mock_response_data = {
            "reportTaskOutcome": {
                "id": "test_id",
                "status": "COMPLETED",
                "completedAt": "2023-01-01T00:00:00Z"
            }
        }
        
        mock_client._gql = mock_gql
        mock_client._loop = mock_loop
        mock_client._headers = {}
        mock_loop.await_coro.return_value = mock_response_data
        mock_gql.get_data.return_value = mock_response_data
        
        # Should not raise an exception
        _report_task_outcome(mock_client, "test_id", "COMPLETED", result={"output": "test"})

    def test_report_task_outcome_with_error(self):
        """Test task outcome reporting with error."""
        mock_client = Mock()
        mock_gql = Mock()
        mock_loop = Mock()
        
        mock_response_data = {
            "reportTaskOutcome": {
                "id": "test_id",
                "status": "FAILED",
                "completedAt": "2023-01-01T00:00:00Z"
            }
        }
        
        mock_client._gql = mock_gql
        mock_client._loop = mock_loop
        mock_client._headers = {}
        mock_loop.await_coro.return_value = mock_response_data
        mock_gql.get_data.return_value = mock_response_data
        
        error_info = {
            "error_type": "ValueError",
            "message": "Test error",
            "traceback": "Traceback info"
        }
        
        # Should not raise an exception
        _report_task_outcome(mock_client, "test_id", "FAILED", error=error_info)

    def test_report_task_outcome_api_failure(self):
        """Test task outcome reporting when API fails."""
        mock_client = Mock()
        mock_client._loop.await_coro.side_effect = Exception("API Error")
        
        with pytest.raises(RuntimeError, match="Failed to report outcome"):
            _report_task_outcome(mock_client, "test_id", "COMPLETED")


class TestRunnerEntrypointIntegration:
    """Integration tests for the runner entrypoint main function."""

    def setup_method(self):
        """Setup test environment."""
        self.test_env = {
            "NUMEROUS_TASK_INSTANCE_ID": "test_instance_123",
            "NUMEROUS_TASK_COLLECTION_NAME": "test_collection",
            "NUMEROUS_TASK_FUNCTION_NAME": "test_task",
            "NUMEROUS_MANIFEST_PATH": "/path/to/manifest.toml",
            "NUMEROUS_API_URL": "https://api.test.com",
            "NUMEROUS_API_ACCESS_TOKEN": "test_token",
            "NUMEROUS_ORGANIZATION_ID": "test_org",
            "NUMEROUS_MOCK_REMOTE_LOGGING": "false"
        }

    def create_test_task_file(self, temp_dir):
        """Create a test task file."""
        task_file = temp_dir / "test_task.py"
        task_file.write_text("""
from numerous.tasks import task, TaskControl

@task
def test_task(value: int):
    return value * 2

@task
def test_task_with_control(tc: TaskControl, value: int):
    tc.update_progress(50.0, "halfway")
    tc.log("Processing value", level="info")
    return value * 3
""")
        return task_file

    def create_test_manifest(self, temp_dir, task_file):
        """Create a test manifest file."""
        manifest_file = temp_dir / "manifest.toml"
        manifest_content = f"""
[[task]]
function_name = "test_task"
source_file = "{task_file.name}"
decorated_function = "test_task"

[[task]]
function_name = "test_task_with_control"
source_file = "{task_file.name}"
decorated_function = "test_task_with_control"
"""
        manifest_file.write_text(manifest_content)
        return manifest_file

    @patch('tomli.load')
    @patch('numerous.tasks.runner_entrypoint.get_client')
    @patch('sys.exit')
    def test_main_success_simple_task(self, mock_exit, mock_get_client, mock_tomli_load):
        """Test successful main execution with simple task."""
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)
            
            # Create test files
            task_file = self.create_test_task_file(temp_path)
            manifest_file = self.create_test_manifest(temp_path, task_file)
            
            # Update environment
            test_env = self.test_env.copy()
            test_env["NUMEROUS_MANIFEST_PATH"] = str(manifest_file)
            
            # Mock manifest loading
            mock_manifest_data = {
                "task": [{
                    "function_name": "test_task",
                    "source_file": task_file.name,
                    "decorated_function": "test_task"
                }]
            }
            mock_tomli_load.return_value = mock_manifest_data
            
            # Mock API client
            mock_client = Mock()
            mock_get_client.return_value = mock_client
            
            # Mock fetch inputs
            mock_client._gql = Mock()
            mock_client._loop = Mock()
            mock_client._headers = {}
            
            # Mock successful input fetch
            input_response = {
                "getTaskInstance": {
                    "id": "test_instance_123",
                    "inputs": "eyJ2YWx1ZSI6IDV9",  # base64 encoded {"value": 5}
                    "taskDefinitionName": "test_task",
                    "status": "RUNNING"
                }
            }
            mock_client._loop.await_coro.return_value = input_response
            mock_client._gql.get_data.return_value = input_response
            
            # Mock successful outcome reporting
            outcome_response = {
                "reportTaskOutcome": {
                    "id": "test_instance_123",
                    "status": "COMPLETED",
                    "completedAt": "2023-01-01T00:00:00Z"
                }
            }
            
            def mock_await_coro(coro):
                # Return input response for first call, outcome response for second
                if not hasattr(mock_await_coro, 'call_count'):
                    mock_await_coro.call_count = 0
                mock_await_coro.call_count += 1
                if mock_await_coro.call_count == 1:
                    return input_response
                else:
                    return outcome_response
            
            mock_client._loop.await_coro.side_effect = mock_await_coro
            
            # Change to temp directory to make task imports work
            original_cwd = os.getcwd()
            os.chdir(temp_dir)
            
            try:
                with patch.dict(os.environ, test_env):
                    main()
                
                # Verify successful exit
                mock_exit.assert_called_once_with(0)
                
            finally:
                os.chdir(original_cwd)

    @patch('tomli.load')
    @patch('numerous.tasks.runner_entrypoint.get_client')
    @patch('sys.exit')
    def test_main_success_task_with_control(self, mock_exit, mock_get_client, mock_tomli_load):
        """Test successful main execution with task that uses TaskControl."""
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)
            
            # Create test files
            task_file = self.create_test_task_file(temp_path)
            manifest_file = self.create_test_manifest(temp_path, task_file)
            
            # Update environment for task with control
            test_env = self.test_env.copy()
            test_env["NUMEROUS_MANIFEST_PATH"] = str(manifest_file)
            test_env["NUMEROUS_TASK_FUNCTION_NAME"] = "test_task_with_control"
            test_env["NUMEROUS_MOCK_REMOTE_LOGGING"] = "true"
            
            # Mock manifest loading
            mock_manifest_data = {
                "task": [{
                    "function_name": "test_task_with_control",
                    "source_file": task_file.name,
                    "decorated_function": "test_task_with_control"
                }]
            }
            mock_tomli_load.return_value = mock_manifest_data
            
            # Mock API client
            mock_client = Mock()
            mock_get_client.return_value = mock_client
            
            # Mock API interactions
            mock_client._gql = Mock()
            mock_client._loop = Mock()
            mock_client._headers = {}
            
            # Mock successful input fetch
            input_response = {
                "getTaskInstance": {
                    "id": "test_instance_123",
                    "inputs": "eyJ2YWx1ZSI6IDV9",  # base64 encoded {"value": 5}
                    "taskDefinitionName": "test_task_with_control",
                    "status": "RUNNING"
                }
            }
            
            # Mock successful outcome reporting
            outcome_response = {
                "reportTaskOutcome": {
                    "id": "test_instance_123",
                    "status": "COMPLETED",
                    "completedAt": "2023-01-01T00:00:00Z"
                }
            }
            
            def mock_await_coro(coro):
                if not hasattr(mock_await_coro, 'call_count'):
                    mock_await_coro.call_count = 0
                mock_await_coro.call_count += 1
                if mock_await_coro.call_count == 1:
                    return input_response
                else:
                    return outcome_response
            
            mock_client._loop.await_coro.side_effect = mock_await_coro
            mock_client._gql.get_data.side_effect = [input_response, outcome_response]
            
            # Change to temp directory
            original_cwd = os.getcwd()
            os.chdir(temp_dir)
            
            try:
                with patch.dict(os.environ, test_env):
                    main()
                
                # Verify successful exit
                mock_exit.assert_called_once_with(0)
                
            finally:
                os.chdir(original_cwd)

    @patch('sys.exit')
    def test_main_missing_env_var(self, mock_exit):
        """Test main function with missing environment variable."""
        with patch.dict(os.environ, {}, clear=True):
            main()
            
            # Should exit with error code
            mock_exit.assert_called_once_with(1)

    @patch('numerous.tasks.runner_entrypoint.get_client')
    @patch('sys.exit')
    def test_main_api_client_failure(self, mock_exit, mock_get_client):
        """Test main function when API client initialization fails."""
        mock_get_client.side_effect = Exception("API client failed")
        
        with patch.dict(os.environ, self.test_env):
            main()
            
            # Should exit with error code
            mock_exit.assert_called_once_with(1)

    @patch('numerous.tasks.runner_entrypoint.get_client')
    @patch('sys.exit')
    def test_main_manifest_not_found(self, mock_exit, mock_get_client):
        """Test main function when manifest file is not found."""
        mock_client = Mock()
        mock_get_client.return_value = mock_client
        
        # Mock API interactions
        mock_client._gql = Mock()
        mock_client._loop = Mock()
        mock_client._headers = {}
        
        input_response = {
            "getTaskInstance": {
                "id": "test_instance_123",
                "inputs": "eyJ2YWx1ZSI6IDV9",
                "taskDefinitionName": "test_task",
                "status": "RUNNING"
            }
        }
        mock_client._loop.await_coro.return_value = input_response
        mock_client._gql.get_data.return_value = input_response
        
        # Use non-existent manifest path
        test_env = self.test_env.copy()
        test_env["NUMEROUS_MANIFEST_PATH"] = "/nonexistent/manifest.toml"
        
        with patch.dict(os.environ, test_env):
            main()
            
            # Should exit with error code
            mock_exit.assert_called_once_with(1)

    @patch('tomli.load')
    @patch('numerous.tasks.runner_entrypoint.get_client')
    @patch('sys.exit')
    def test_main_task_execution_failure(self, mock_exit, mock_get_client, mock_tomli_load):
        """Test main function when task execution fails."""
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)
            
            # Create failing task file
            task_file = temp_path / "failing_task.py"
            task_file.write_text("""
from numerous.tasks import task

@task
def failing_task(value: int):
    raise ValueError("Task intentionally failed")
""")
            
            # Create manifest for failing task
            manifest_file = temp_path / "manifest.toml"
            manifest_content = f"""
[[task]]
function_name = "failing_task"
source_file = "{task_file.name}"
decorated_function = "failing_task"
"""
            manifest_file.write_text(manifest_content)
            
            # Update environment
            test_env = self.test_env.copy()
            test_env["NUMEROUS_MANIFEST_PATH"] = str(manifest_file)
            test_env["NUMEROUS_TASK_FUNCTION_NAME"] = "failing_task"
            
            # Mock manifest loading
            mock_manifest_data = {
                "task": [{
                    "function_name": "failing_task",
                    "source_file": task_file.name,
                    "decorated_function": "failing_task"
                }]
            }
            mock_tomli_load.return_value = mock_manifest_data
            
            # Mock API client
            mock_client = Mock()
            mock_get_client.return_value = mock_client
            
            # Mock API interactions
            mock_client._gql = Mock()
            mock_client._loop = Mock()
            mock_client._headers = {}
            
            input_response = {
                "getTaskInstance": {
                    "id": "test_instance_123",
                    "inputs": "eyJ2YWx1ZSI6IDV9",
                    "taskDefinitionName": "failing_task",
                    "status": "RUNNING"
                }
            }
            
            outcome_response = {
                "reportTaskOutcome": {
                    "id": "test_instance_123",
                    "status": "FAILED",
                    "completedAt": "2023-01-01T00:00:00Z"
                }
            }
            
            def mock_await_coro(coro):
                if not hasattr(mock_await_coro, 'call_count'):
                    mock_await_coro.call_count = 0
                mock_await_coro.call_count += 1
                if mock_await_coro.call_count == 1:
                    return input_response
                else:
                    return outcome_response
            
            mock_client._loop.await_coro.side_effect = mock_await_coro
            mock_client._gql.get_data.side_effect = [input_response, outcome_response]
            
            # Change to temp directory
            original_cwd = os.getcwd()
            os.chdir(temp_dir)
            
            try:
                with patch.dict(os.environ, test_env):
                    main()
                
                # Should exit with error code due to task failure
                mock_exit.assert_called_once_with(1)
                
            finally:
                os.chdir(original_cwd)

    @patch('numerous.tasks.runner_entrypoint.get_client')
    @patch('sys.exit')
    def test_main_input_fetch_failure(self, mock_exit, mock_get_client):
        """Test main function when input fetching fails."""
        mock_client = Mock()
        mock_get_client.return_value = mock_client
        
        # Mock API client to fail on input fetch
        mock_client._loop.await_coro.side_effect = Exception("Input fetch failed")
        
        with patch.dict(os.environ, self.test_env):
            main()
            
            # Should exit with error code
            mock_exit.assert_called_once_with(1)

    @patch('tomli.load')
    @patch('numerous.tasks.runner_entrypoint.get_client')
    @patch('sys.exit')
    def test_main_outcome_reporting_failure(self, mock_exit, mock_get_client, mock_tomli_load):
        """Test main function when outcome reporting fails."""
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)
            
            # Create test files
            task_file = self.create_test_task_file(temp_path)
            manifest_file = self.create_test_manifest(temp_path, task_file)
            
            # Update environment
            test_env = self.test_env.copy()
            test_env["NUMEROUS_MANIFEST_PATH"] = str(manifest_file)
            
            # Mock manifest loading
            mock_manifest_data = {
                "task": [{
                    "function_name": "test_task",
                    "source_file": task_file.name,
                    "decorated_function": "test_task"
                }]
            }
            mock_tomli_load.return_value = mock_manifest_data
            
            # Mock API client
            mock_client = Mock()
            mock_get_client.return_value = mock_client
            
            # Mock API interactions
            mock_client._gql = Mock()
            mock_client._loop = Mock()
            mock_client._headers = {}
            
            input_response = {
                "getTaskInstance": {
                    "id": "test_instance_123",
                    "inputs": "eyJ2YWx1ZSI6IDV9",
                    "taskDefinitionName": "test_task",
                    "status": "RUNNING"
                }
            }
            
            def mock_await_coro(coro):
                if not hasattr(mock_await_coro, 'call_count'):
                    mock_await_coro.call_count = 0
                mock_await_coro.call_count += 1
                if mock_await_coro.call_count == 1:
                    return input_response
                else:
                    # Fail on outcome reporting
                    raise Exception("Outcome reporting failed")
            
            mock_client._loop.await_coro.side_effect = mock_await_coro
            mock_client._gql.get_data.return_value = input_response
            
            # Change to temp directory
            original_cwd = os.getcwd()
            os.chdir(temp_dir)
            
            try:
                with patch.dict(os.environ, test_env):
                    with patch('sys.stderr'):  # Suppress stderr output during test
                        main()
                
                # Should exit with error code due to reporting failure
                mock_exit.assert_called_once_with(1)
                
            finally:
                os.chdir(original_cwd)


class TestRunnerEntrypointEdgeCases:
    """Test edge cases and error conditions."""

    def test_fetch_task_inputs_invalid_base64(self):
        """Test task input fetching with invalid base64 data."""
        mock_client = Mock()
        mock_gql = Mock()
        mock_loop = Mock()
        
        mock_response_data = {
            "getTaskInstance": {
                "id": "test_id",
                "inputs": "invalid_base64",
                "taskDefinitionName": "test_task",
                "status": "RUNNING"
            }
        }
        
        mock_client._gql = mock_gql
        mock_client._loop = mock_loop
        mock_client._headers = {}
        mock_loop.await_coro.return_value = mock_response_data
        mock_gql.get_data.return_value = mock_response_data
        
        with pytest.raises(Exception):  # Should raise some decoding error
            _fetch_task_inputs(mock_client, "test_id")

    def test_fetch_task_inputs_invalid_json(self):
        """Test task input fetching with invalid JSON data."""
        mock_client = Mock()
        mock_gql = Mock()
        mock_loop = Mock()
        
        # Base64 encoded "invalid json"
        invalid_json_b64 = "aW52YWxpZCBqc29u"
        
        mock_response_data = {
            "getTaskInstance": {
                "id": "test_id",
                "inputs": invalid_json_b64,
                "taskDefinitionName": "test_task",
                "status": "RUNNING"
            }
        }
        
        mock_client._gql = mock_gql
        mock_client._loop = mock_loop
        mock_client._headers = {}
        mock_loop.await_coro.return_value = mock_response_data
        mock_gql.get_data.return_value = mock_response_data
        
        with pytest.raises(Exception):  # Should raise JSON decode error
            _fetch_task_inputs(mock_client, "test_id")

    def test_find_task_details_empty_manifest(self):
        """Test finding task details in empty manifest."""
        manifest_data = {"task": []}
        
        with pytest.raises(RunnerError, match="Task function 'test_task' not found"):
            _find_task_details_in_manifest(manifest_data, "test_task")

    def test_find_task_details_no_task_section(self):
        """Test finding task details when no task section exists."""
        manifest_data = {}
        
        with pytest.raises(RunnerError, match="Task function 'test_task' not found"):
            _find_task_details_in_manifest(manifest_data, "test_task") 