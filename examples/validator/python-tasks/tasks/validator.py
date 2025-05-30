"""
Validator tasks for testing Numerous Tasks functionality.
These tasks demonstrate different types of operations and can be used
to validate that the task execution system works correctly.
"""

import json
import os
import platform
import sys
import tempfile
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional

try:
    import pandas as pd
    import numpy as np
    import requests
except ImportError as e:
    print(f"Warning: Some dependencies not available: {e}")
    pd = None
    np = None
    requests = None


def validate_environment(context: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
    """
    Validate the Python environment and available dependencies.
    
    Args:
        context: Optional execution context
        
    Returns:
        Dict containing environment validation results
    """
    print("🔍 Validating Python environment...")
    
    results = {
        "timestamp": datetime.now().isoformat(),
        "python_version": sys.version,
        "platform": platform.platform(),
        "architecture": platform.architecture(),
        "executable": sys.executable,
        "dependencies": {},
        "environment_vars": {},
        "validation_status": "success"
    }
    
    # Check Python version
    python_version = sys.version_info
    results["python_info"] = {
        "major": python_version.major,
        "minor": python_version.minor,
        "micro": python_version.micro,
        "version_string": f"{python_version.major}.{python_version.minor}.{python_version.micro}"
    }
    
    # Check dependencies
    dependencies = {
        "pandas": pd,
        "numpy": np,
        "requests": requests
    }
    
    for name, module in dependencies.items():
        if module is not None:
            try:
                version = getattr(module, "__version__", "unknown")
                results["dependencies"][name] = {
                    "available": True,
                    "version": version
                }
                print(f"✅ {name}: {version}")
            except Exception as e:
                results["dependencies"][name] = {
                    "available": False,
                    "error": str(e)
                }
                print(f"❌ {name}: {e}")
        else:
            results["dependencies"][name] = {
                "available": False,
                "error": "Module not imported"
            }
            print(f"❌ {name}: Not available")
    
    # Check some environment variables
    env_vars = ["HOME", "PATH", "PYTHONPATH", "NUMEROUS_API_URL"]
    for var in env_vars:
        value = os.environ.get(var)
        results["environment_vars"][var] = value is not None
        if value:
            # Only show first 50 chars for security
            display_value = value[:50] + "..." if len(value) > 50 else value
            print(f"📝 {var}: {display_value}")
    
    print("✅ Environment validation completed")
    return results


def process_data(
    data: Optional[List[Dict[str, Any]]] = None,
    context: Optional[Dict[str, Any]] = None
) -> Dict[str, Any]:
    """
    Process sample data to test data handling capabilities.
    
    Args:
        data: Optional input data, defaults to sample data
        context: Optional execution context
        
    Returns:
        Dict containing processed data results
    """
    print("📊 Processing sample data...")
    
    # Use sample data if none provided
    if data is None:
        data = [
            {"id": 1, "name": "Alice", "score": 85, "department": "Engineering"},
            {"id": 2, "name": "Bob", "score": 92, "department": "Sales"},
            {"id": 3, "name": "Charlie", "score": 78, "department": "Engineering"},
            {"id": 4, "name": "Diana", "score": 96, "department": "Marketing"},
            {"id": 5, "name": "Eve", "score": 88, "department": "Sales"},
        ]
    
    results = {
        "timestamp": datetime.now().isoformat(),
        "input_count": len(data),
        "processing_steps": [],
        "output_data": {},
        "statistics": {}
    }
    
    try:
        # Convert to pandas DataFrame if available
        if pd is not None:
            df = pd.DataFrame(data)
            results["processing_steps"].append("Converted to pandas DataFrame")
            
            # Basic statistics
            if "score" in df.columns:
                stats = {
                    "mean_score": float(df["score"].mean()),
                    "median_score": float(df["score"].median()),
                    "std_score": float(df["score"].std()),
                    "min_score": float(df["score"].min()),
                    "max_score": float(df["score"].max())
                }
                results["statistics"]["scores"] = stats
                results["processing_steps"].append("Calculated score statistics")
            
            # Group by department if available
            if "department" in df.columns:
                dept_counts = df["department"].value_counts().to_dict()
                results["statistics"]["department_counts"] = dept_counts
                results["processing_steps"].append("Grouped by department")
                
                if "score" in df.columns:
                    dept_scores = df.groupby("department")["score"].mean().to_dict()
                    results["statistics"]["department_avg_scores"] = {
                        k: float(v) for k, v in dept_scores.items()
                    }
                    results["processing_steps"].append("Calculated department averages")
            
            # Export processed data
            results["output_data"]["processed_records"] = df.to_dict("records")
            
        else:
            # Fallback processing without pandas
            results["processing_steps"].append("Processing without pandas")
            
            # Manual statistics calculation
            scores = [item.get("score", 0) for item in data if "score" in item]
            if scores:
                results["statistics"]["scores"] = {
                    "mean_score": sum(scores) / len(scores),
                    "min_score": min(scores),
                    "max_score": max(scores),
                    "count": len(scores)
                }
            
            # Manual grouping
            departments = {}
            for item in data:
                dept = item.get("department", "Unknown")
                if dept not in departments:
                    departments[dept] = []
                departments[dept].append(item)
            
            results["statistics"]["department_counts"] = {
                dept: len(items) for dept, items in departments.items()
            }
            
            results["output_data"]["processed_records"] = data
        
        print(f"✅ Processed {len(data)} records")
        print(f"📈 Statistics: {len(results['statistics'])} categories")
        
    except Exception as e:
        results["error"] = str(e)
        results["processing_steps"].append(f"Error: {e}")
        print(f"❌ Data processing failed: {e}")
    
    return results


def file_operations(
    test_content: Optional[str] = None,
    context: Optional[Dict[str, Any]] = None
) -> Dict[str, Any]:
    """
    Test file I/O operations.
    
    Args:
        test_content: Optional content to write to test file
        context: Optional execution context
        
    Returns:
        Dict containing file operation results
    """
    print("📁 Testing file I/O operations...")
    
    if test_content is None:
        test_content = f"Test file created at {datetime.now().isoformat()}\nThis is a validation test.\n"
    
    results = {
        "timestamp": datetime.now().isoformat(),
        "operations": [],
        "files_created": [],
        "success": True
    }
    
    try:
        # Create temporary directory
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)
            results["operations"].append(f"Created temporary directory: {temp_dir}")
            
            # Write test file
            test_file = temp_path / "test_file.txt"
            test_file.write_text(test_content, encoding="utf-8")
            results["operations"].append(f"Wrote test file: {test_file}")
            results["files_created"].append(str(test_file))
            
            # Read test file
            read_content = test_file.read_text(encoding="utf-8")
            results["operations"].append("Read test file back")
            
            # Verify content
            content_match = read_content == test_content
            results["operations"].append(f"Content verification: {'✅ PASS' if content_match else '❌ FAIL'}")
            
            # Create subdirectory and nested file
            sub_dir = temp_path / "subdir"
            sub_dir.mkdir()
            nested_file = sub_dir / "nested.json"
            
            test_data = {
                "test": True,
                "timestamp": datetime.now().isoformat(),
                "nested_level": 1
            }
            
            nested_file.write_text(json.dumps(test_data, indent=2), encoding="utf-8")
            results["operations"].append(f"Created nested JSON file: {nested_file}")
            results["files_created"].append(str(nested_file))
            
            # Read and parse JSON
            read_data = json.loads(nested_file.read_text(encoding="utf-8"))
            json_match = read_data == test_data
            results["operations"].append(f"JSON verification: {'✅ PASS' if json_match else '❌ FAIL'}")
            
            # List directory contents
            dir_contents = list(temp_path.iterdir())
            results["operations"].append(f"Directory listing: {len(dir_contents)} items")
            
            # File size checks
            file_size = test_file.stat().st_size
            results["file_stats"] = {
                "test_file_size": file_size,
                "nested_file_size": nested_file.stat().st_size,
                "total_files": len(results["files_created"])
            }
            
            results["operations"].append(f"File stats collected")
            
        results["operations"].append("Temporary directory cleaned up")
        print("✅ File operations completed successfully")
        
    except Exception as e:
        results["success"] = False
        results["error"] = str(e)
        results["operations"].append(f"Error: {e}")
        print(f"❌ File operations failed: {e}")
    
    return results


def network_check(
    test_urls: Optional[List[str]] = None,
    context: Optional[Dict[str, Any]] = None
) -> Dict[str, Any]:
    """
    Check network connectivity and API calls.
    
    Args:
        test_urls: Optional list of URLs to test
        context: Optional execution context
        
    Returns:
        Dict containing network check results
    """
    print("🌐 Testing network connectivity...")
    
    if test_urls is None:
        test_urls = [
            "https://httpbin.org/get",
            "https://api.github.com/repos/numerous-com/numerous-sdk",
            "https://jsonplaceholder.typicode.com/posts/1"
        ]
    
    results = {
        "timestamp": datetime.now().isoformat(),
        "tests": [],
        "summary": {
            "total_tests": len(test_urls),
            "successful": 0,
            "failed": 0
        }
    }
    
    if requests is None:
        results["error"] = "requests module not available"
        results["tests"].append({
            "test": "dependency_check",
            "status": "failed",
            "error": "requests module not imported"
        })
        print("❌ Network tests skipped: requests module not available")
        return results
    
    for url in test_urls:
        test_result = {
            "url": url,
            "timestamp": datetime.now().isoformat()
        }
        
        try:
            print(f"🔗 Testing {url}...")
            response = requests.get(url, timeout=10)
            
            test_result.update({
                "status": "success",
                "status_code": response.status_code,
                "response_time": response.elapsed.total_seconds(),
                "content_length": len(response.content),
                "headers": dict(response.headers)
            })
            
            # Try to parse JSON if possible
            try:
                if response.headers.get("content-type", "").startswith("application/json"):
                    json_data = response.json()
                    test_result["json_keys"] = list(json_data.keys()) if isinstance(json_data, dict) else None
                    test_result["json_parseable"] = True
                else:
                    test_result["json_parseable"] = False
            except:
                test_result["json_parseable"] = False
            
            results["summary"]["successful"] += 1
            print(f"✅ {url} - Status: {response.status_code}, Time: {response.elapsed.total_seconds():.2f}s")
            
        except requests.exceptions.Timeout:
            test_result.update({
                "status": "timeout",
                "error": "Request timed out"
            })
            results["summary"]["failed"] += 1
            print(f"⏱️ {url} - Timeout")
            
        except requests.exceptions.RequestException as e:
            test_result.update({
                "status": "failed",
                "error": str(e)
            })
            results["summary"]["failed"] += 1
            print(f"❌ {url} - Error: {e}")
        
        except Exception as e:
            test_result.update({
                "status": "failed",
                "error": f"Unexpected error: {e}"
            })
            results["summary"]["failed"] += 1
            print(f"💥 {url} - Unexpected error: {e}")
        
        results["tests"].append(test_result)
    
    success_rate = (results["summary"]["successful"] / results["summary"]["total_tests"]) * 100
    results["summary"]["success_rate"] = success_rate
    
    print(f"🎯 Network tests completed: {success_rate:.1f}% success rate")
    return results


if __name__ == "__main__":
    """
    Run all validation tasks when script is executed directly.
    This is useful for local testing.
    """
    print("🚀 Running all validation tasks...")
    
    print("\n" + "="*50)
    env_results = validate_environment()
    
    print("\n" + "="*50)
    data_results = process_data()
    
    print("\n" + "="*50)
    file_results = file_operations()
    
    print("\n" + "="*50)
    network_results = network_check()
    
    print("\n" + "="*50)
    print("📋 Validation Summary:")
    print(f"✅ Environment validation: {'PASS' if env_results['validation_status'] == 'success' else 'FAIL'}")
    print(f"✅ Data processing: {'PASS' if 'error' not in data_results else 'FAIL'}")
    print(f"✅ File operations: {'PASS' if file_results['success'] else 'FAIL'}")
    print(f"✅ Network connectivity: {'PASS' if network_results['summary']['success_rate'] > 50 else 'FAIL'}")
    print("🏁 All validation tasks completed!") 