#!/usr/bin/env python3
"""
Flask web application for Docker validator API endpoints.
Provides HTTP API endpoints for task execution validation.
"""

import json
import os
import platform
import sys
from datetime import datetime
from flask import Flask, jsonify, request

# Import task functions
from tasks.validator import validate_container, process_data

app = Flask(__name__)

@app.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint for task validation."""
    return jsonify({
        "status": "healthy",
        "timestamp": datetime.now().isoformat(),
        "service": "docker-validator",
        "version": "1.0.0",
        "python_version": sys.version,
        "platform": platform.platform()
    })

@app.route('/process', methods=['POST'])
def api_process():
    """Data processing API endpoint."""
    try:
        # Get input data from request
        input_data = request.get_json() if request.is_json else {}
        
        # Call the processing function
        result = process_data()
        
        # Add API-specific metadata
        result["api_info"] = {
            "endpoint": "/process",
            "method": "POST",
            "input_provided": bool(input_data),
            "processing_mode": "api_endpoint"
        }
        
        return jsonify(result)
        
    except Exception as e:
        return jsonify({
            "error": str(e),
            "timestamp": datetime.now().isoformat(),
            "endpoint": "/process"
        }), 500

@app.route('/validate', methods=['GET'])
def api_validate():
    """Container validation API endpoint."""
    try:
        result = validate_container()
        
        # Add API-specific metadata
        result["api_info"] = {
            "endpoint": "/validate",
            "method": "GET",
            "processing_mode": "api_endpoint"
        }
        
        return jsonify(result)
        
    except Exception as e:
        return jsonify({
            "error": str(e),
            "timestamp": datetime.now().isoformat(),
            "endpoint": "/validate"
        }), 500

@app.route('/info', methods=['GET'])
def info():
    """Get general information about the API."""
    return jsonify({
        "service": "docker-validator",
        "version": "1.0.0",
        "description": "Docker task collection validator API",
        "endpoints": {
            "/health": "Health check endpoint",
            "/process": "Data processing endpoint (POST)",
            "/validate": "Container validation endpoint (GET)",
            "/info": "API information endpoint (GET)"
        },
        "timestamp": datetime.now().isoformat(),
        "environment": {
            "python_version": sys.version,
            "platform": platform.platform(),
            "in_docker": os.path.exists('/.dockerenv')
        }
    })

@app.errorhandler(404)
def not_found(error):
    """Handle 404 errors."""
    return jsonify({
        "error": "Endpoint not found",
        "available_endpoints": ["/health", "/process", "/validate", "/info"],
        "timestamp": datetime.now().isoformat()
    }), 404

@app.errorhandler(500)
def internal_error(error):
    """Handle 500 errors."""
    return jsonify({
        "error": "Internal server error",
        "timestamp": datetime.now().isoformat()
    }), 500

if __name__ == "__main__":
    # Configuration
    host = os.environ.get("HOST", "0.0.0.0")
    port = int(os.environ.get("PORT", 8080))
    debug = os.environ.get("DEBUG", "false").lower() == "true"
    
    print(f"🚀 Starting Docker validator API server...")
    print(f"📍 Server: http://{host}:{port}")
    print(f"🔧 Debug mode: {debug}")
    print(f"🐳 In Docker: {os.path.exists('/.dockerenv')}")
    
    # Start the Flask application
    app.run(host=host, port=port, debug=debug) 