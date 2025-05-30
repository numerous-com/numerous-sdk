#!/usr/bin/env python3
"""
Docker validator tasks that can be executed as entrypoints.
These demonstrate different execution modes in Docker containers.
"""

import json
import os
import platform
import sys
from datetime import datetime
from pathlib import Path

# Import shared validation logic
# Note: In a real deployment, this would be the actual validator module
def validate_container():
    """Validate Docker container environment."""
    print("🐳 Validating Docker container environment...")
    
    results = {
        "timestamp": datetime.now().isoformat(),
        "platform": platform.platform(),
        "architecture": platform.architecture(),
        "python_version": sys.version,
        "container_info": {},
        "validation_status": "success"
    }
    
    # Check if running in container
    if os.path.exists('/.dockerenv'):
        results["container_info"]["in_docker"] = True
        print("✅ Running inside Docker container")
    else:
        results["container_info"]["in_docker"] = False
        print("⚠️ Not detected as running in Docker")
    
    # Check container OS
    try:
        with open('/etc/os-release', 'r') as f:
            os_info = f.read()
            results["container_info"]["os_release"] = os_info[:200]  # First 200 chars
        print("✅ Container OS info retrieved")
    except:
        results["container_info"]["os_release"] = "unavailable"
        print("⚠️ Could not read OS release info")
    
    # Check available commands
    commands = ["curl", "jq", "ps"]
    available_commands = {}
    for cmd in commands:
        if os.system(f"which {cmd} > /dev/null 2>&1") == 0:
            available_commands[cmd] = True
            print(f"✅ {cmd} available")
        else:
            available_commands[cmd] = False
            print(f"❌ {cmd} not available")
    
    results["container_info"]["available_commands"] = available_commands
    
    # Check environment variables
    env_vars = ["HOME", "PATH", "PYTHONPATH", "NUMEROUS_API_URL"]
    env_info = {}
    for var in env_vars:
        value = os.environ.get(var)
        env_info[var] = value is not None
        if value:
            print(f"📝 {var}: {value[:50]}{'...' if len(value) > 50 else ''}")
    
    results["container_info"]["environment_vars"] = env_info
    
    print("✅ Container validation completed")
    
    # Output JSON for programmatic consumption
    print(f"\n--- VALIDATION_RESULTS ---")
    print(json.dumps(results, indent=2))
    print(f"--- END_VALIDATION_RESULTS ---")
    
    return results

def process_data():
    """Process data in Docker environment."""
    print("📊 Processing data in Docker container...")
    
    # Sample data processing
    data = [
        {"id": 1, "value": 100, "category": "A"},
        {"id": 2, "value": 200, "category": "B"},
        {"id": 3, "value": 150, "category": "A"},
        {"id": 4, "value": 300, "category": "C"},
    ]
    
    results = {
        "timestamp": datetime.now().isoformat(),
        "input_count": len(data),
        "processing_mode": "docker_entrypoint",
        "results": {}
    }
    
    # Calculate statistics
    values = [item["value"] for item in data]
    results["results"]["statistics"] = {
        "total": sum(values),
        "average": sum(values) / len(values),
        "min": min(values),
        "max": max(values)
    }
    
    # Group by category
    categories = {}
    for item in data:
        cat = item["category"]
        if cat not in categories:
            categories[cat] = []
        categories[cat].append(item["value"])
    
    category_stats = {}
    for cat, vals in categories.items():
        category_stats[cat] = {
            "count": len(vals),
            "total": sum(vals),
            "average": sum(vals) / len(vals)
        }
    
    results["results"]["category_breakdown"] = category_stats
    
    print(f"✅ Processed {len(data)} records")
    print(f"📈 Found {len(categories)} categories")
    
    # Output JSON
    print(f"\n--- PROCESSING_RESULTS ---")
    print(json.dumps(results, indent=2))
    print(f"--- END_PROCESSING_RESULTS ---")
    
    return results

if __name__ == "__main__":
    """Execute tasks based on command line arguments."""
    if len(sys.argv) < 2:
        print("Usage: python validator.py <task_name>")
        print("Available tasks: validate_container, process_data")
        sys.exit(1)
    
    task_name = sys.argv[1]
    
    if task_name == "validate_container":
        validate_container()
    elif task_name == "process_data":
        process_data()
    else:
        print(f"Unknown task: {task_name}")
        sys.exit(1) 