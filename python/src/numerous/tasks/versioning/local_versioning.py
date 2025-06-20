"""
Local development versioning for tasks.

Provides automatic version generation for development tasks using content hashing.
Local versions use the format: local-{hash} where hash is derived from task definition.
"""

import hashlib
import json
from typing import Dict, Any, Optional


def generate_local_version(task_definition: Dict[str, Any]) -> str:
    """
    Generate a local development version based on task definition hash.
    
    Args:
        task_definition: Task definition dictionary containing function metadata
        
    Returns:
        Version string in format "local-{hash}"
    """
    # Create a stable hash of the task definition
    # Sort keys to ensure consistent hashing
    definition_str = json.dumps(task_definition, sort_keys=True, separators=(',', ':'))
    
    # Generate SHA-256 hash and take first 8 characters
    hash_obj = hashlib.sha256(definition_str.encode('utf-8'))
    hash_hex = hash_obj.hexdigest()[:8]
    
    return f"local-{hash_hex}"


def is_local_version(version: str) -> bool:
    """
    Check if a version string is a local development version.
    
    Args:
        version: Version string to check
        
    Returns:
        True if version is a local development version
    """
    return version.startswith("local-")


def extract_task_definition(task_func) -> Dict[str, Any]:
    """
    Extract task definition metadata from a task function.
    
    Args:
        task_func: Task function with @task decorator
        
    Returns:
        Dictionary containing task definition metadata
    """
    definition = {
        "function_name": getattr(task_func, '__name__', 'unknown'),
        "module": getattr(task_func, '__module__', 'unknown'),
        "doc": getattr(task_func, '__doc__', ''),
    }
    
    # Add task-specific metadata if available
    if hasattr(task_func, '_task_config'):
        config = task_func._task_config
        definition.update({
            "max_parallel": getattr(config, 'max_parallel', None),
            "resource_sizing": getattr(config, 'resource_sizing', None),
            "timeout": getattr(config, 'timeout', None),
        })
    
    # Add function signature information
    import inspect
    try:
        sig = inspect.signature(task_func)
        definition["parameters"] = {
            name: {
                "annotation": str(param.annotation) if param.annotation != inspect.Parameter.empty else None,
                "default": str(param.default) if param.default != inspect.Parameter.empty else None,
            }
            for name, param in sig.parameters.items()
        }
    except (ValueError, TypeError):
        # Handle cases where signature inspection fails
        definition["parameters"] = {}
    
    return definition


def get_task_version(task_func, explicit_version: Optional[str] = None) -> str:
    """
    Get the version for a task, either explicit or auto-generated local version.
    
    Args:
        task_func: Task function
        explicit_version: Explicit version if provided in @task decorator
        
    Returns:
        Version string (explicit version or local-{hash})
    """
    if explicit_version:
        return explicit_version
    
    # Generate local version from task definition
    task_definition = extract_task_definition(task_func)
    return generate_local_version(task_definition) 