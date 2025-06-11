"""
Integration tests for different task execution modes.

These tests validate actual task execution with different backends and configurations.
Run with: pytest -m integration tests/integration/
"""

import pytest
import os
import tempfile
import subprocess
import json
import time
from pathlib import Path
from unittest.mock import patch, Mock

from numerous.tasks import task, TaskControl, Session
from numerous.tasks.runner_entrypoint import main


pytestmark = pytest.mark.integration


class TestLocalBackendExecution:
    """Test tasks running with local backend (developer mode)."""
    
    def test_simple_task_local_execution(self):
        """Test simple task execution in local backend."""
        @task
        def add_numbers(a: int, b: int):
            return a + b
        
        with Session() as session:
            result = add_numbers(10, 5)
            assert result == 15
            
            # Verify session tracking
            assert session.get_running_instances_count("add_numbers") == 0
    
    def test_task_with_control_local_execution(self):
        """Test TaskControl-enabled task in local backend."""
        progress_updates = []
        log_messages = []
        
        @task
        def processing_task(tc: TaskControl, items: list):
            tc.log("Starting processing", level="info")
            log_messages.append("Starting processing")
            
            for i, item in enumerate(items):
                progress = (i + 1) / len(items) * 100
                tc.update_progress(progress, f"Processing {item}")
                progress_updates.append((progress, f"Processing {item}"))
                
                if tc.should_stop:
                    tc.log("Task stopped early", level="warning")
                    return f"stopped_at_{i}"
            
            tc.log("Processing complete", level="info")
            log_messages.append("Processing complete")
            return f"processed_{len(items)}_items"
        
        with Session() as session:
            result = processing_task(["a", "b", "c"])
            assert result == "processed_3_items"
            assert len(progress_updates) == 3
            assert progress_updates[-1] == (100.0, "Processing c")
            assert len(log_messages) == 2
    
    def test_parallel_task_execution_local(self):
        """Test multiple tasks running in parallel with local backend."""
        import threading
        import concurrent.futures
        
        @task
        def concurrent_task(task_id: int, delay: float = 0.1):
            time.sleep(delay)
            return f"task_{task_id}_done"
        
        # Test concurrent execution
        with concurrent.futures.ThreadPoolExecutor(max_workers=3) as executor:
            futures = []
            for i in range(5):
                # Each thread gets its own session
                def run_task(task_id):
                    with Session() as session:
                        return concurrent_task(task_id, 0.05)
                
                future = executor.submit(run_task, i)
                futures.append(future)
            
            results = [f.result() for f in futures]
            assert len(results) == 5
            assert all(f"task_{i}_done" in results for i in range(5))


class TestMockBackendExecution:
    """Test task-runner with mock backend (Numerous dev team mode)."""
    
    def create_test_task_and_manifest(self, temp_dir, task_name="test_task", use_control=False):
        """Helper to create task file and manifest."""
        task_file = temp_dir / f"{task_name}.py"
        
        if use_control:
            task_content = f"""
from numerous.tasks import task, TaskControl

@task
def {task_name}(tc: TaskControl, value: int):
    tc.log(f"Processing value: {{value}}", level="info")
    tc.update_progress(25.0, "started")
    tc.update_progress(50.0, "halfway")
    tc.update_progress(100.0, "completed")
    return value * 2
"""
        else:
            task_content = f"""
from numerous.tasks import task

@task
def {task_name}(value: int):
    return value * 2
"""
        
        task_file.write_text(task_content)
        
        # Create manifest
        manifest_file = temp_dir / "numerous-task.toml"
        manifest_content = f"""
[[task]]
function_name = "{task_name}"
source_file = "{task_file.name}"
decorated_function = "{task_name}"
"""
        manifest_file.write_text(manifest_content)
        
        return task_file, manifest_file
    
    @patch('numerous.tasks.runner_entrypoint.get_client')
    @patch('tomli.load')
    def test_runner_mock_backend_simple_task(self, mock_tomli_load, mock_get_client):
        """Test task-runner with mock backend for simple task."""
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)
            
            # Create test task and manifest
            task_file, manifest_file = self.create_test_task_and_manifest(temp_path)
            
            # Mock manifest loading
            mock_tomli_load.return_value = {
                "task": [{
                    "function_name": "test_task",
                    "source_file": task_file.name,
                    "decorated_function": "test_task"
                }]
            }
            
            # Mock API client
            mock_client = Mock()
            mock_get_client.return_value = mock_client
            
            # Mock API responses
            input_response = {
                "getTaskInstance": {
                    "id": "integration_test_123",
                    "inputs": "eyJ2YWx1ZSI6IDEwfQ==",  # base64 {"value": 10}
                    "taskDefinitionName": "test_task",
                    "status": "RUNNING"
                }
            }
            
            outcome_response = {
                "reportTaskOutcome": {
                    "id": "integration_test_123", 
                    "status": "COMPLETED",
                    "completedAt": "2023-01-01T00:00:00Z"
                }
            }
            
            call_count = 0
            def mock_await_coro(coro):
                nonlocal call_count
                call_count += 1
                return input_response if call_count == 1 else outcome_response
            
            mock_client._gql = Mock()
            mock_client._loop = Mock()
            mock_client._headers = {}
            mock_client._loop.await_coro.side_effect = mock_await_coro
            mock_client._gql.get_data.side_effect = [input_response, outcome_response]
            
            # Set up environment for task-runner
            test_env = {
                "NUMEROUS_TASK_INSTANCE_ID": "integration_test_123",
                "NUMEROUS_TASK_COLLECTION_NAME": "integration_test",
                "NUMEROUS_TASK_FUNCTION_NAME": "test_task",
                "NUMEROUS_MANIFEST_PATH": str(manifest_file),
                "NUMEROUS_API_URL": "https://api.test.com",
                "NUMEROUS_API_ACCESS_TOKEN": "test_token",
                "NUMEROUS_ORGANIZATION_ID": "test_org",
                "NUMEROUS_MOCK_REMOTE_LOGGING": "false"
            }
            
            # Change to temp directory and run
            original_cwd = os.getcwd()
            os.chdir(temp_dir)
            
            try:
                with patch.dict(os.environ, test_env):
                    with patch('sys.exit') as mock_exit:
                        main()
                        mock_exit.assert_called_once_with(0)
                
                # Verify API interactions
                assert mock_client._loop.await_coro.call_count == 2
                
            finally:
                os.chdir(original_cwd)
    
    @patch('numerous.tasks.runner_entrypoint.get_client')
    @patch('tomli.load')
    def test_runner_mock_backend_with_control(self, mock_tomli_load, mock_get_client):
        """Test task-runner with mock backend for TaskControl-enabled task."""
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)
            
            # Create TaskControl-enabled task
            task_file, manifest_file = self.create_test_task_and_manifest(
                temp_path, "control_task", use_control=True
            )
            
            # Mock manifest loading
            mock_tomli_load.return_value = {
                "task": [{
                    "function_name": "control_task",
                    "source_file": task_file.name,
                    "decorated_function": "control_task"
                }]
            }
            
            # Mock API client
            mock_client = Mock()
            mock_get_client.return_value = mock_client
            
            # Mock API responses
            input_response = {
                "getTaskInstance": {
                    "id": "control_test_456",
                    "inputs": "eyJ2YWx1ZSI6IDE1fQ==",  # base64 {"value": 15}
                    "taskDefinitionName": "control_task",
                    "status": "RUNNING"
                }
            }
            
            outcome_response = {
                "reportTaskOutcome": {
                    "id": "control_test_456",
                    "status": "COMPLETED", 
                    "completedAt": "2023-01-01T00:00:00Z"
                }
            }
            
            call_count = 0
            def mock_await_coro(coro):
                nonlocal call_count
                call_count += 1
                return input_response if call_count == 1 else outcome_response
            
            mock_client._gql = Mock()
            mock_client._loop = Mock()
            mock_client._headers = {}
            mock_client._loop.await_coro.side_effect = mock_await_coro
            mock_client._gql.get_data.side_effect = [input_response, outcome_response]
            
            # Environment with mock remote logging enabled
            test_env = {
                "NUMEROUS_TASK_INSTANCE_ID": "control_test_456",
                "NUMEROUS_TASK_COLLECTION_NAME": "integration_test",
                "NUMEROUS_TASK_FUNCTION_NAME": "control_task",
                "NUMEROUS_MANIFEST_PATH": str(manifest_file),
                "NUMEROUS_API_URL": "https://api.test.com", 
                "NUMEROUS_API_ACCESS_TOKEN": "test_token",
                "NUMEROUS_ORGANIZATION_ID": "test_org",
                "NUMEROUS_MOCK_REMOTE_LOGGING": "true"  # Enable mock remote logging
            }
            
            original_cwd = os.getcwd()
            os.chdir(temp_dir)
            
            try:
                with patch.dict(os.environ, test_env):
                    with patch('sys.exit') as mock_exit:
                        # Capture stdout to verify mock remote logging
                        import io
                        from contextlib import redirect_stdout
                        
                        stdout_capture = io.StringIO()
                        with redirect_stdout(stdout_capture):
                            main()
                        
                        mock_exit.assert_called_once_with(0)
                        
                        # Verify mock remote logging output
                        output = stdout_capture.getvalue()
                        assert "[PoCMockRemoteTC][LOG][INFO]" in output
                        assert "[PoCMockRemoteTC][PROGRESS]" in output
                        assert "Processing value: 15" in output
                
            finally:
                os.chdir(original_cwd)


@pytest.mark.slow
class TestTaskRunnerSubprocess:
    """Test task-runner as a subprocess (closer to real execution)."""
    
    def test_subprocess_task_execution(self):
        """Test running task-runner as subprocess with mock environment."""
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)
            
            # Create a simple task
            task_file = temp_path / "subprocess_task.py"
            task_file.write_text("""
from numerous.tasks import task

@task
def subprocess_task(message: str):
    return f"subprocess processed: {message}"
""")
            
            # Create manifest
            manifest_file = temp_path / "numerous-task.toml"
            manifest_file.write_text("""
[[task]]
function_name = "subprocess_task"
source_file = "subprocess_task.py"
decorated_function = "subprocess_task"
""")
            
            # This test would require more complex setup to mock the API
            # For now, we'll skip it in CI but it shows the structure
            pytest.skip("Subprocess testing requires full API mock setup")


# Pytest configuration for integration tests
def pytest_configure(config):
    """Register custom markers."""
    config.addinivalue_line(
        "markers", "integration: mark test as integration test"
    )
    config.addinivalue_line(
        "markers", "slow: mark test as slow running test"
    ) 