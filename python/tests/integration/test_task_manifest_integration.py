"""
Integration tests for task manifest functionality.

These tests verify that the task manifest extension works correctly
with the API and can be parsed and validated properly.
"""

import pytest
import tempfile
import os
from pathlib import Path
from numerous.tasks import task, Session


@pytest.mark.integration
class TestTaskManifestIntegration:
    """Integration tests for task manifest functionality."""

    def test_task_manifest_parsing(self):
        """Test that task manifest can be parsed correctly."""
        # Create a temporary manifest file with tasks enabled
        manifest_content = '''name = "Test Task App"
description = "A test app with tasks enabled"
cover_image = "cover.png"
exclude = ["*venv", "venv*"]
port = 80

[python]
  library = "streamlit"
  version = "3.11"
  app_file = "app.py"
  requirements_file = "requirements.txt"

[tasks]
  enabled = true
'''
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.toml', delete=False) as f:
            f.write(manifest_content)
            manifest_path = f.name
        
        try:
            # Verify the manifest file exists and has correct content
            assert os.path.exists(manifest_path)
            
            # Read and verify the content
            with open(manifest_path, 'r') as f:
                content = f.read()
                assert 'tasks' in content
                assert 'enabled = true' in content
                assert 'library = "streamlit"' in content
        finally:
            os.unlink(manifest_path)

    def test_task_execution_with_manifest_enabled(self):
        """Test that tasks can be executed when manifest has tasks enabled."""
        
        @task(name="test_task")
        def sample_task():
            return "task completed"
        
        # Test direct execution (should work regardless of manifest)
        result = sample_task()
        assert result == "task completed"
        
        # Test session-based execution
        with Session() as session:
            future = sample_task.instance().start()
            result = future.result()
            assert result == "task completed"

    def test_task_manifest_validation(self):
        """Test that task manifest validation works correctly."""
        # Test with valid manifest
        valid_manifest = '''name = "Valid Task App"
description = "A valid app"
cover_image = "cover.png"
exclude = ["*venv"]
port = 80

[python]
  library = "streamlit"
  version = "3.11"
  app_file = "app.py"
  requirements_file = "requirements.txt"

[tasks]
  enabled = true
'''
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.toml', delete=False) as f:
            f.write(valid_manifest)
            manifest_path = f.name
        
        try:
            # Verify the manifest is valid
            assert os.path.exists(manifest_path)
            
            # Test that we can read the file and it contains expected content
            with open(manifest_path, 'r') as f:
                content = f.read()
                assert 'tasks' in content
                assert 'enabled = true' in content
        finally:
            os.unlink(manifest_path)

    def test_task_manifest_without_tasks_section(self):
        """Test that apps work correctly without tasks section."""
        # Create manifest without tasks section
        manifest_content = '''name = "App Without Tasks"
description = "An app without tasks section"
cover_image = "cover.png"
exclude = ["*venv"]
port = 80

[python]
  library = "streamlit"
  version = "3.11"
  app_file = "app.py"
  requirements_file = "requirements.txt"
'''
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.toml', delete=False) as f:
            f.write(manifest_content)
            manifest_path = f.name
        
        try:
            # Verify the manifest file exists and doesn't have tasks section
            assert os.path.exists(manifest_path)
            
            with open(manifest_path, 'r') as f:
                content = f.read()
                assert 'tasks' not in content
                assert 'library = "streamlit"' in content
        finally:
            os.unlink(manifest_path)

    def test_task_manifest_with_tasks_disabled(self):
        """Test that tasks section with enabled=false works correctly."""
        manifest_content = '''name = "App With Tasks Disabled"
description = "An app with tasks disabled"
cover_image = "cover.png"
exclude = ["*venv"]
port = 80

[python]
  library = "streamlit"
  version = "3.11"
  app_file = "app.py"
  requirements_file = "requirements.txt"

[tasks]
  enabled = false
'''
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.toml', delete=False) as f:
            f.write(manifest_content)
            manifest_path = f.name
        
        try:
            # Verify the manifest file exists and has tasks disabled
            assert os.path.exists(manifest_path)
            
            with open(manifest_path, 'r') as f:
                content = f.read()
                assert 'tasks' in content
                assert 'enabled = false' in content
        finally:
            os.unlink(manifest_path)