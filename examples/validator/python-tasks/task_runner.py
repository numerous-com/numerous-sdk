#!/usr/bin/env python3
"""
Simple task runner for Numerous task collections.
This script can be used as a Docker entrypoint to execute specific tasks.
"""

import importlib.util
import json
import os
import sys
from pathlib import Path


def load_task_manifest():
    """Load the task manifest from numerous-task.toml"""
    try:
        import toml
        # Prefer manifest in current working directory (e.g. /app within container)
        # then fallback to a relative path if not found (less likely in container).
        manifest_path = Path("numerous-task.toml")
        if not manifest_path.exists():
            # Fallback for cases where CWD might not be /app
            manifest_path_alt = Path("/app/numerous-task.toml")
            if manifest_path_alt.exists():
                manifest_path = manifest_path_alt
            elif not manifest_path.exists(): # if original numerous-task.toml also didn't exist
                 print("Warning: numerous-task.toml not found in CWD or /app.")
                 return None


        with open(manifest_path, 'r') as f:
            return toml.load(f)
    except ImportError:
        # Fallback to simple parsing if toml not available
        print("Warning: toml module not available, using fallback parser")
        return None


def find_task_by_name(manifest, function_name):
    """Find a task definition by function name"""
    if not manifest or 'task' not in manifest:
        return None
    
    for task in manifest['task']:
        if task.get('function_name') == function_name:
            return task
    return None


def load_and_execute_task(source_file, function_name, task_script_args=None):
    """Load a Python module and execute the specified function"""
    if task_script_args is None:
        task_script_args = []
    
    try:
        # Add /app to Python path if it exists
        app_path = Path("/app")
        if app_path.exists():
            sys.path.insert(0, str(app_path))
        
        # Load the module
        module_path = Path(source_file)
        if not module_path.is_absolute():
            if app_path.exists():
                module_path = app_path / source_file
            else:
                module_path = Path.cwd() / source_file
        
        spec = importlib.util.spec_from_file_location("task_module", module_path)
        if spec is None:
            raise ImportError(f"Could not load module from {module_path}")
        
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        
        # Get the function
        if not hasattr(module, function_name):
            raise AttributeError(f"Function '{function_name}' not found in {source_file}")
        
        func = getattr(module, function_name)
        
        # Execute the function
        print(f"🚀 Executing {function_name} from {source_file} with args: {task_script_args}")
        result = func(*task_script_args)
        
        print(f"✅ Task completed successfully")
        return result
        
    except Exception as e:
        print(f"❌ Task execution failed: {e}")
        raise


def main():
    """Main entrypoint for the task runner"""
    # Get function name from command line (first argument)
    if len(sys.argv) > 1:
        function_name = sys.argv[1]
    else:
        function_name = os.environ.get('TASK_FUNCTION_NAME') # Fallback for older mechanism if needed

    if not function_name:
        print("Error: No function name specified.")
        print(f"Usage: python {sys.argv[0]} <function_name> [arg1 arg2 ...]")
        print("   or: Set TASK_FUNCTION_NAME environment variable (fallback)")
        sys.exit(1)

    # Subsequent arguments are for the task itself
    task_script_args = sys.argv[2:]
    print(f"Task arguments received: {task_script_args}")

    # Load manifest to find task details
    manifest = load_task_manifest()
    task_def = find_task_by_name(manifest, function_name) if manifest else None
    
    if task_def:
        source_file = task_def['source_file']
        print(f"📋 Found task definition: {function_name} in {source_file}")
    else:
        # Fallback: try common patterns
        possible_files = [
            f"tasks/{function_name}.py",
            f"tasks/validator.py",  # Common case for our example
            f"{function_name}.py"
        ]
        
        source_file = None
        for candidate in possible_files:
            candidate_path = Path(candidate)
            if candidate_path.exists():
                source_file = candidate
                break
        
        if not source_file:
            print(f"⚠️  No manifest found and no source file found. Tried: {possible_files}")
            source_file = "tasks/validator.py"  # Default for our example
        
        print(f"⚠️  No manifest found, using source file: {source_file}")
    
    # Execute the task
    try:
        result = load_and_execute_task(source_file, function_name, task_script_args)
        
        # Output result as JSON
        output = {
            "status": "success",
            "result": result
        }
        
        print("📤 Task output:")
        print(json.dumps(output, indent=2, default=str))
        
    except Exception as e:
        output = {
            "status": "error",
            "error": str(e)
        }
        
        print("📤 Task output:")
        print(json.dumps(output, indent=2))
        sys.exit(1)


if __name__ == "__main__":
    main() 