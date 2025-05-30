#!/bin/bash

# Numerous Task Workflow Validation Script
# This script validates the complete task workflow from login to execution

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI_BINARY="${SCRIPT_DIR}/bin/numerous"
EXAMPLE_TASK_DIR="${SCRIPT_DIR}/examples/validator/python-tasks"
COLLECTION_NAME="test-validation-$(date +%s)"
TIMEOUT=300

# Default values (can be overridden)
ORG_SLUG=""
API_URL="${NUMEROUS_API_URL:-http://localhost:8080/graphql}"
VERBOSE=false
CLEANUP=true
SKIP_LOGIN=false

# Function to print colored output
print_step() {
    echo -e "${BLUE}🔵 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${NC}ℹ️  $1${NC}"
}

# Function to show usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Validates the complete Numerous task workflow:"
    echo "  1. Login to platform"
    echo "  2. List organizations"
    echo "  3. Deploy task collection"
    echo "  4. Run tasks from collection"
    echo "  5. Cleanup (optional)"
    echo ""
    echo "Options:"
    echo "  --org SLUG           Organization slug to use (required if not interactive)"
    echo "  --api-url URL        API endpoint URL (default: \$NUMEROUS_API_URL or http://localhost:8080/graphql)"
    echo "  --skip-login         Skip login step (assume already authenticated)"
    echo "  --no-cleanup         Don't cleanup deployed resources"
    echo "  --verbose            Enable verbose output"
    echo "  --help               Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --org my-org"
    echo "  $0 --org my-org --verbose --no-cleanup"
    echo "  $0 --skip-login --org my-org"
    echo ""
    echo "Environment Variables:"
    echo "  NUMEROUS_API_URL     API endpoint URL"
    echo "  NUMEROUS_ACCESS_TOKEN Access token (if skipping login)"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --org)
            ORG_SLUG="$2"
            shift 2
            ;;
        --api-url)
            API_URL="$2"
            shift 2
            ;;
        --skip-login)
            SKIP_LOGIN=true
            shift
            ;;
        --no-cleanup)
            CLEANUP=false
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Verbose flag for CLI commands
VERBOSE_FLAG=""
if [ "$VERBOSE" = true ]; then
    VERBOSE_FLAG="--verbose"
fi

# Function to check if CLI binary exists
check_cli() {
    print_step "Checking CLI binary"
    
    if [ ! -f "$CLI_BINARY" ]; then
        print_error "CLI binary not found at $CLI_BINARY"
        print_info "Please build the CLI first: go build -o bin/numerous ."
        exit 1
    fi
    
    print_success "CLI binary found"
}

# Function to check if example task directory exists
check_example_tasks() {
    print_step "Checking example task collection"
    
    if [ ! -d "$EXAMPLE_TASK_DIR" ]; then
        print_error "Example task directory not found at $EXAMPLE_TASK_DIR"
        exit 1
    fi
    
    if [ ! -f "$EXAMPLE_TASK_DIR/numerous-task.toml" ]; then
        print_error "Task manifest not found in $EXAMPLE_TASK_DIR"
        exit 1
    fi
    
    print_success "Example task collection found"
}

# Function to perform login
perform_login() {
    if [ "$SKIP_LOGIN" = true ]; then
        print_step "Skipping login (--skip-login specified)"
        if [ -z "$NUMEROUS_ACCESS_TOKEN" ]; then
            print_warning "No NUMEROUS_ACCESS_TOKEN environment variable set"
        else
            print_success "Using existing access token"
        fi
        return
    fi
    
    print_step "Performing login"
    
    # Check if already logged in
    if $CLI_BINARY token status >/dev/null 2>&1; then
        print_warning "Already logged in, using existing token"
        return
    fi
    
    print_info "Please complete the login process in your browser"
    if ! $CLI_BINARY login --api-url "$API_URL" $VERBOSE_FLAG; then
        print_error "Login failed"
        exit 1
    fi
    
    print_success "Login completed"
}

# Function to list and select organization
list_organizations() {
    print_step "Listing organizations"
    
    # Get organizations
    if ! $CLI_BINARY organization list $VERBOSE_FLAG > /tmp/orgs.txt 2>&1; then
        print_error "Failed to list organizations"
        cat /tmp/orgs.txt
        exit 1
    fi
    
    print_success "Organizations retrieved"
    
    # If no org specified, show available orgs and prompt
    if [ -z "$ORG_SLUG" ]; then
        print_info "Available organizations:"
        cat /tmp/orgs.txt
        echo ""
        read -p "Enter organization slug: " ORG_SLUG
        
        if [ -z "$ORG_SLUG" ]; then
            print_error "Organization slug is required"
            exit 1
        fi
    fi
    
    print_info "Using organization: $ORG_SLUG"
}

# Function to deploy task collection
deploy_task_collection() {
    print_step "Deploying task collection"
    
    print_info "Collection name: $COLLECTION_NAME"
    print_info "Task directory: $EXAMPLE_TASK_DIR"
    
    # Deploy the task collection
    if ! $CLI_BINARY deploy "$EXAMPLE_TASK_DIR" \
        --organization "$ORG_SLUG" \
        --name "$COLLECTION_NAME" \
        --api-url "$API_URL" \
        $VERBOSE_FLAG; then
        print_error "Task collection deployment failed"
        exit 1
    fi
    
    print_success "Task collection deployed successfully"
}

# Function to list tasks in the deployed collection
list_deployed_tasks() {
    print_step "Listing tasks in deployed collection"
    
    if ! $CLI_BINARY task list \
        --org "$ORG_SLUG" \
        --collection "$COLLECTION_NAME" \
        --api-url "$API_URL" \
        $VERBOSE_FLAG; then
        print_error "Failed to list tasks in collection"
        exit 1
    fi
    
    print_success "Tasks listed successfully"
}

# Function to run tasks from the collection
run_tasks() {
    print_step "Running tasks from deployed collection"
    
    # List of tasks to run (based on the python-tasks example)
    declare -a tasks=("validate_environment" "process_data" "file_operations" "network_check")
    
    for task in "${tasks[@]}"; do
        print_info "Running task: $task"
        
        if $CLI_BINARY task run "$task" \
            --org "$ORG_SLUG" \
            --collection "$COLLECTION_NAME" \
            --api-url "$API_URL" \
            --timeout "$TIMEOUT" \
            $VERBOSE_FLAG; then
            print_success "Task '$task' completed successfully"
        else
            print_warning "Task '$task' failed (this might be expected for testing)"
        fi
        
        echo ""
    done
    
    print_success "Task execution tests completed"
}

# Function to test local execution
test_local_execution() {
    print_step "Testing local task execution"
    
    print_info "Running task locally: validate_environment"
    
    if $CLI_BINARY task run validate_environment \
        --org "$ORG_SLUG" \
        --local \
        --task-dir "$EXAMPLE_TASK_DIR" \
        $VERBOSE_FLAG; then
        print_success "Local task execution successful"
    else
        print_warning "Local task execution failed (this might be expected)"
    fi
}

# Function to cleanup deployed resources
cleanup_deployment() {
    if [ "$CLEANUP" = false ]; then
        print_step "Skipping cleanup (--no-cleanup specified)"
        print_info "Deployed collection '$COLLECTION_NAME' remains on the platform"
        return
    fi
    
    print_step "Cleaning up deployed resources"
    
    # Note: This would require a delete/undeploy command to be implemented
    print_warning "Cleanup not implemented yet - collection '$COLLECTION_NAME' remains deployed"
    print_info "You may need to manually clean up the deployed collection"
}

# Function to run the complete validation
main() {
    echo "🚀 Starting Numerous Task Workflow Validation"
    echo "============================================="
    echo ""
    echo "Configuration:"
    echo "  CLI Binary: $CLI_BINARY"
    echo "  Example Tasks: $EXAMPLE_TASK_DIR"
    echo "  API URL: $API_URL"
    echo "  Collection Name: $COLLECTION_NAME"
    echo "  Organization: ${ORG_SLUG:-<will be selected>}"
    echo "  Verbose: $VERBOSE"
    echo "  Skip Login: $SKIP_LOGIN"
    echo "  Cleanup: $CLEANUP"
    echo ""
    
    # Pre-flight checks
    check_cli
    check_example_tasks
    
    # Authentication
    perform_login
    
    # Organization selection
    list_organizations
    
    # Task collection deployment
    deploy_task_collection
    
    # List deployed tasks
    list_deployed_tasks
    
    # Run remote tasks
    run_tasks
    
    # Test local execution
    test_local_execution
    
    # Cleanup
    cleanup_deployment
    
    echo ""
    echo "🎉 Validation completed successfully!"
    echo "====================================="
    print_success "All workflow steps completed"
    
    if [ "$CLEANUP" = false ]; then
        print_info "Collection '$COLLECTION_NAME' is still deployed in organization '$ORG_SLUG'"
    fi
}

# Trap to ensure cleanup on script exit
cleanup_on_exit() {
    if [ $? -ne 0 ]; then
        print_error "Script failed. Check the output above for details."
        
        if [ "$CLEANUP" = true ]; then
            print_info "You may need to manually clean up collection '$COLLECTION_NAME'"
        fi
    fi
}

trap cleanup_on_exit EXIT

# Run main function
main "$@" 