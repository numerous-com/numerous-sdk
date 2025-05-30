#!/bin/bash
# Shell script entrypoint for system validation tasks

set -e

function system_info() {
    echo "🖥️ Gathering system information..."
    
    echo "--- SYSTEM_INFO ---"
    
    # Basic system info
    echo "Hostname: $(hostname)"
    echo "Uptime: $(uptime)"
    echo "Date: $(date)"
    
    # OS information
    echo "OS: $(cat /etc/os-release | grep PRETTY_NAME | cut -d= -f2 | tr -d '\"')"
    echo "Kernel: $(uname -r)"
    echo "Architecture: $(uname -m)"
    
    # Memory info
    echo "Memory:"
    free -h | head -2
    
    # Disk info
    echo "Disk usage:"
    df -h | head -2
    
    # CPU info
    echo "CPU:"
    cat /proc/cpuinfo | grep "model name" | head -1 | cut -d: -f2 | xargs
    echo "CPU cores: $(nproc)"
    
    # Process info
    echo "Processes: $(ps aux | wc -l)"
    
    # Network interfaces
    echo "Network interfaces:"
    ip addr show | grep "inet " | awk '{print $2}' | head -5
    
    # Environment variables (safe subset)
    echo "Key environment variables:"
    env | grep -E "^(HOME|PATH|USER|PWD)" | head -5
    
    echo "--- END_SYSTEM_INFO ---"
    
    echo "✅ System information gathered successfully"
}

function usage() {
    echo "Usage: $0 <command>"
    echo "Available commands:"
    echo "  system_info - Gather system information"
    echo "  help        - Show this help message"
}

# Main script logic
case "${1:-}" in
    "system_info")
        system_info
        ;;
    "help"|"--help"|"-h")
        usage
        ;;
    "")
        echo "❌ No command provided"
        usage
        exit 1
        ;;
    *)
        echo "❌ Unknown command: $1"
        usage
        exit 1
        ;;
esac 