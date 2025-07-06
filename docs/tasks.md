# Tasks

With Numerous Tasks, you can define, execute, and manage long-running Python functions as distributed tasks.
Tasks provide a powerful way to handle asynchronous operations, background processing, and scalable workloads with seamless integration across development, testing, and production environments.

## Key Features

1. **Define tasks** using the simple `@task` decorator on Python functions
2. **Execute tasks** locally for development or remotely on the Numerous platform
3. **Monitor progress** with real-time status updates, progress tracking, and structured logging
4. **Manage sessions** to coordinate multiple related tasks with concurrency control
5. **Handle cancellation** and error recovery with graceful shutdown patterns
6. **Integrate seamlessly** with FastAPI, Streamlit, and other Python frameworks
7. **Deploy easily** using the Numerous CLI for production workloads

!!! tip
    Remember to add `numerous` as a dependency in your project; most likely to your `requirements.txt` file.

## Quick Start

Import the [Numerous SDK](http://www.pypi.org/project/numerous) in your Python code:

```python
import time
from numerous.tasks import task, Session, TaskControl

# Define a simple task
@task
def process_data(data: list) -> dict:
    """Process a list of data items."""
    result = {"processed": len(data), "items": data}
    return result

# Define a task with progress tracking
@task
def long_running_task(tc: TaskControl, iterations: int) -> str:
    """Execute a long-running task with progress updates."""
    for i in range(iterations):
        # Check if task should stop
        if tc.should_stop:
            tc.log("Task was cancelled", "warning")
            break
            
        # Update progress
        progress = (i + 1) / iterations * 100
        tc.update_progress(progress, f"Processing item {i+1}/{iterations}")
        
        # Simulate work
        time.sleep(0.1)
    
    tc.log("Task completed successfully", "info")
    return f"Processed {iterations} items"

# Execute tasks within a session
with Session() as session:
    # Direct execution (for simple cases)
    result = process_data([1, 2, 3, 4, 5])
    print(result)  # {"processed": 5, "items": [1, 2, 3, 4, 5]}
    
    # Task instance execution (for advanced features)
    instance = long_running_task.instance()
    future = instance.start(10)
    
    # Get result (blocks until complete)
    result = future.result()
    print(result)  # "Processed 10 items"
```

## Development Workflows

Numerous Tasks support multiple execution modes to accommodate different stages of your development workflow:

### 1. Direct Execution (Development & Testing)

Perfect for rapid development and unit testing. Tasks execute synchronously as regular function calls:

```python
from numerous.tasks import task

@task
def validate_data(data: dict) -> bool:
    """Validate data structure."""
    return all(key in data for key in ['id', 'name', 'value'])

# Direct execution - no session required
# Great for development and testing
is_valid = validate_data({'id': 1, 'name': 'test', 'value': 42})
print(is_valid)  # True
```

**Use cases:**
- Unit testing task logic
- Interactive development in Jupyter notebooks
- Quick validation of task behavior

### 2. Local Task Instances (Integration Testing)

Execute tasks as instances with full TaskControl features while running locally:

```python
from numerous.tasks import task, Session, TaskControl

@task
def integration_test_task(tc: TaskControl, config: dict) -> dict:
    """Test task with full monitoring capabilities."""
    tc.log("Starting integration test", "info")
    tc.update_progress(50, "Validating configuration")
    
    # Your test logic here
    result = {"status": "success", "config": config}
    
    tc.update_progress(100, "Test completed")
    tc.log("Integration test finished", "info")
    return result

# Local execution with task control
with Session() as session:
    instance = integration_test_task.instance()
    future = instance.start({"env": "test"})
    result = future.result()
    print(result)
```

**Use cases:**
- Integration testing with monitoring
- Local development with progress tracking
- Testing cancellation and error handling

### 3. Remote Execution (Production)

Execute tasks on the Numerous platform with automatic scaling and resource management:

```python
import os

# Configure for remote execution
os.environ['NUMEROUS_TASK_BACKEND'] = 'remote'

# Same code works for remote execution
with Session() as session:
    instance = process_large_dataset.instance()
    future = instance.start(massive_dataset)
    result = future.result()  # Executes on Numerous platform
```

**Use cases:**
- Production workloads
- Resource-intensive tasks
- Scaling beyond local machine capabilities

## Framework Integration

### FastAPI Integration

Integrate tasks seamlessly with FastAPI applications for robust web APIs:

```python
from fastapi import FastAPI, Request, BackgroundTasks
from numerous.tasks import task, Session, TaskControl
from numerous.frameworks.fastapi import get_session

app = FastAPI()

@task
def process_upload(tc: TaskControl, file_data: bytes, filename: str) -> dict:
    """Process uploaded file in background."""
    tc.log(f"Processing file: {filename}", "info")
    tc.update_progress(50, "Analyzing file")
    
    # Your processing logic
    result = {"filename": filename, "size": len(file_data), "status": "processed"}
    
    tc.update_progress(100, "Processing complete")
    return result

@app.post("/upload")
async def upload_file(request: Request, file_data: bytes, filename: str):
    """Upload endpoint that triggers background processing."""
    # Get user session from request
    session = get_session(request)
    
    # Start background task
    with session:
        instance = process_upload.instance()
        future = instance.start(file_data, filename)
        
        return {
            "message": "File uploaded successfully",
            "task_id": instance.id,
            "status": "processing"
        }

@app.get("/status/{task_id}")
async def get_task_status(task_id: str):
    """Check task status."""
    # Implementation depends on your task tracking needs
    return {"task_id": task_id, "status": "completed"}
```

### Streamlit Integration

Create interactive data processing applications with Streamlit:

```python
import streamlit as st
from numerous.tasks import task, Session, TaskControl
from numerous.frameworks.streamlit import get_session

@task
def analyze_data(tc: TaskControl, data: list, analysis_type: str) -> dict:
    """Analyze data with progress updates."""
    tc.log(f"Starting {analysis_type} analysis", "info")
    
    results = {}
    total_steps = len(data)
    
    for i, item in enumerate(data):
        if tc.should_stop:
            break
            
        # Simulate analysis
        progress = (i + 1) / total_steps * 100
        tc.update_progress(progress, f"Analyzing item {i+1}/{total_steps}")
        
        # Add to results
        results[f"item_{i}"] = {"value": item, "analysis": analysis_type}
    
    tc.log("Analysis completed", "info")
    return results

# Streamlit app
st.title("Data Analysis with Tasks")

# Get user session
session = get_session()

# UI elements
data_input = st.text_area("Enter data (one item per line)")
analysis_type = st.selectbox("Analysis Type", ["statistical", "qualitative"])

if st.button("Start Analysis"):
    if data_input:
        data = [line.strip() for line in data_input.split('\n') if line.strip()]
        
        with session:
            instance = analyze_data.instance()
            future = instance.start(data, analysis_type)
            
            # Show progress
            progress_bar = st.progress(0)
            status_text = st.empty()
            
            # In a real app, you'd want to poll for progress
            # This is a simplified example
            with st.spinner("Processing..."):
                result = future.result()
                
            st.success("Analysis completed!")
            st.json(result)
```

## CLI Integration and Deployment

### Local Development with CLI

Use the Numerous CLI to manage your task-based applications:

```bash
# Initialize a new project
numerous init

# Configure your numerous.toml
```

```toml
# numerous.toml
name = "Task Processing App"
description = "An application that processes data using Numerous Tasks"
port = 8000

[python]
version = "3.11"
library = "fastapi"
app_file = "main.py"
requirements_file = "requirements.txt"

[deploy]
app = "my-task-app"
organization = "my-org-slug"
```

### Deployment Workflow

1. **Develop locally** using direct execution and local task instances
2. **Test integration** with your chosen framework (FastAPI, Streamlit)
3. **Deploy to production** using the Numerous CLI

```bash
# Deploy your application
numerous deploy

# Monitor logs
numerous logs

# Check deployment status
numerous list
```

### Environment Configuration

Configure task execution based on your environment:

```python
import os
from numerous.tasks import task, Session, TaskControl

# Configure backend based on environment
if os.getenv('ENVIRONMENT') == 'production':
    os.environ['NUMEROUS_TASK_BACKEND'] = 'remote'
else:
    os.environ['NUMEROUS_TASK_BACKEND'] = 'local'

@task
def environment_aware_task(tc: TaskControl, data: dict) -> dict:
    """Task that adapts to different environments."""
    env = os.getenv('ENVIRONMENT', 'development')
    tc.log(f"Running in {env} environment", "info")
    
    if env == 'production':
        # Use production-specific logic
        tc.log("Using production optimizations", "info")
    else:
        # Use development-specific logic
        tc.log("Using development settings", "info")
    
    return {"environment": env, "data": data}
```

## Session Management

### Understanding the Two Types of Sessions

It's important to distinguish between two different session concepts:

#### 1. **User Sessions** (Application/Authentication Sessions)
- User authentication and application state management
- Already in production and managed by your application framework
- Examples: HTTP sessions, login sessions, browser sessions
- Typically managed by Flask, FastAPI, Django, etc.

#### 2. **Task Sessions** (Task Grouping and Management)
- Organizational contexts for grouping related task instances  
- Provide concurrency control and coordination for task execution
- Can be named, persisted, and retrieved for later use
- Independent from user authentication - one user might have multiple task sessions

### Task Session Core Concepts

Task Sessions provide context for managing related tasks with built-in concurrency control:

```python
from numerous.tasks import Session, task, TaskControl

@task(max_parallel=3)
def batch_processor(tc: TaskControl, batch_id: str, items: list) -> dict:
    """Process a batch of items."""
    tc.log(f"Processing batch {batch_id}", "info")
    
    results = []
    for i, item in enumerate(items):
        if tc.should_stop:
            break
            
        # Process item
        result = {"item": item, "processed": True}
        results.append(result)
        
        # Update progress
        progress = (i + 1) / len(items) * 100
        tc.update_progress(progress, f"Batch {batch_id}: {i+1}/{len(items)}")
    
    return {"batch_id": batch_id, "results": results}

# Create a named task session
with Session(name="batch-processing-session") as session:
    futures = []
    
    # Start multiple batches (respecting max_parallel=3)
    for i, batch_data in enumerate(dataset_batches):
        instance = batch_processor.instance()
        future = instance.start(f"batch_{i}", batch_data)
        futures.append(future)
    
    # Collect results
    results = []
    for future in futures:
        result = future.result()
        results.append(result)
    
    print(f"Processed {len(results)} batches in session: {session.name}")
```

### Integrating User Sessions with Task Sessions

Here are common patterns for connecting user sessions with task sessions:

#### Pattern 1: Task Sessions Tagged with User Session ID

For short-lived, cheap tasks that are tied to a specific user session:

```python
from numerous.tasks import Session, task, TaskControl
from flask import request, session as user_session  # User session from Flask

@task
def user_data_task(tc: TaskControl, user_id: str, data: dict) -> dict:
    """Process user-specific data."""
    tc.log(f"Processing data for user {user_id}", "info")
    # Process data specific to this user
    return {"user_id": user_id, "processed": data}

def process_user_data(data):
    """Process data within the context of the current user session."""
    user_id = user_session.get('user_id')  # From user's HTTP session
    
    # Create task session tagged with user session info
    task_session_name = f"user_{user_id}_session_{user_session.get('session_id')}"
    
    with Session(name=task_session_name) as task_session:
        instance = user_data_task.instance()
        future = instance.start(user_id, data)
        result = future.result()
        
        return result
```

#### Pattern 2: Persistent Named Task Sessions (Long-running Workflows)

For longer-running or user-managed task sessions that can be saved and loaded:

```python
from numerous.tasks import Session, task, TaskControl
from typing import Dict, List
import uuid

# In-memory store (use database/collection in production)
SAVED_TASK_SESSIONS: Dict[str, Dict] = {}

@task(max_parallel=2)
def data_analysis_task(tc: TaskControl, dataset_name: str, analysis_type: str) -> dict:
    """Perform data analysis."""
    tc.log(f"Analyzing {dataset_name} with {analysis_type}", "info")
    # Analysis logic here
    return {"dataset": dataset_name, "analysis": analysis_type, "result": "analyzed"}

class TaskSessionManager:
    """Manager for persistent task sessions."""
    
    @staticmethod
    def create_named_session(session_name: str, user_id: str) -> str:
        """Create a new named task session."""
        session_id = str(uuid.uuid4())
        
        session_metadata = {
            "session_id": session_id,
            "name": session_name,
            "user_id": user_id,
            "created_at": "2024-01-01T00:00:00Z",
            "status": "active",
            "task_instances": []
        }
        
        SAVED_TASK_SESSIONS[session_id] = session_metadata
        return session_id
    
    @staticmethod
    def get_user_sessions(user_id: str) -> List[Dict]:
        """Get all task sessions for a specific user."""
        return [
            session for session in SAVED_TASK_SESSIONS.values()
            if session["user_id"] == user_id
        ]
    
    @staticmethod
    def resume_session(session_id: str) -> Dict:
        """Resume a previously created task session."""
        return SAVED_TASK_SESSIONS.get(session_id)

# Usage in your application
def start_analysis_project(user_id: str, project_name: str, datasets: List[str]):
    """Start a new analysis project for a user."""
    
    # Create persistent task session
    session_id = TaskSessionManager.create_named_session(
        session_name=f"{user_id}_{project_name}_analysis",
        user_id=user_id
    )
    
    # Execute tasks within the named session
    with Session(name=f"analysis_project_{session_id}") as task_session:
        futures = []
        
        for dataset in datasets:
            instance = data_analysis_task.instance()
            future = instance.start(dataset, "statistical_analysis")
            futures.append(future)
        
        # Update session metadata
        SAVED_TASK_SESSIONS[session_id]["task_instances"] = [
            {"task": "data_analysis_task", "dataset": dataset} 
            for dataset in datasets
        ]
        
        # Collect results
        results = [future.result() for future in futures]
        
        # Mark session as completed
        SAVED_TASK_SESSIONS[session_id]["status"] = "completed"
        SAVED_TASK_SESSIONS[session_id]["results"] = results
        
        return session_id, results

def load_user_sessions_for_ui(user_id: str):
    """Get user's task sessions for UI display (Load Session - select from list)."""
    user_sessions = TaskSessionManager.get_user_sessions(user_id)
    
    # Format for UI dropdown/selection
    return [
        {
            "id": session["session_id"],
            "name": session["name"],
            "created": session["created_at"],
            "status": session["status"],
            "task_count": len(session["task_instances"])
        }
        for session in user_sessions
    ]
```

#### Pattern 2b: Task Session Persistence with Numerous Collections

For production applications, persist task sessions using Numerous collection documents:

```python
from numerous.tasks import Session, task, TaskControl
from numerous.collections import collection
from typing import Dict, List, Optional
import uuid
import datetime
import json

@task(max_parallel=3)
def document_processing_task(tc: TaskControl, doc_id: str, operation: str) -> dict:
    """Process a document with specified operation."""
    tc.log(f"Processing document {doc_id} with operation {operation}", "info")
    
    # Simulate document processing
    for i in range(5):
        if tc.should_stop:
            tc.log("Task cancelled during processing", "warning")
            break
        
        progress = (i + 1) * 20
        tc.update_progress(progress, f"Processing step {i+1}/5")
        
        # Simulate processing time
        import time
        time.sleep(0.2)
    
    return {
        "document_id": doc_id,
        "operation": operation,
        "status": "completed",
        "processed_at": datetime.datetime.now().isoformat()
    }

class PersistentTaskSessionManager:
    """Manage task sessions with persistence to Numerous collections."""
    
    def __init__(self, collection_key: str = "task-sessions"):
        self.sessions_collection = collection(collection_key)
    
    def create_user_workflow(self, user_id: str, workflow_name: str, description: str = "") -> str:
        """Create a new named workflow that can be resumed later."""
        session_id = str(uuid.uuid4())
        
        # Create task session metadata
        workflow_metadata = {
            "session_id": session_id,
            "user_id": user_id,
            "workflow_name": workflow_name,
            "description": description,
            "created_at": datetime.datetime.now().isoformat(),
            "updated_at": datetime.datetime.now().isoformat(),
            "status": "active",
            "task_instances": [],
            "results": {},
            "progress": 0,
            "tags": {
                "user_id": user_id,
                "workflow_type": "document_processing",
                "status": "active"
            }
        }
        
        # Save to collection with tags for filtering
        doc_ref = self.sessions_collection.document(session_id)
        doc_ref.set(workflow_metadata)
        
        # Add tags for efficient querying
        doc_ref.tag("user_id", user_id)
        doc_ref.tag("workflow_name", workflow_name)
        doc_ref.tag("status", "active")
        doc_ref.tag("workflow_type", "document_processing")
        
        return session_id
    
    def get_user_workflows(self, user_id: str, status: Optional[str] = None) -> List[Dict]:
        """Get all workflows for a specific user (for UI 'Load Session' dropdown)."""
        workflows = []
        
        # Query documents by user_id tag
        for doc_ref in self.sessions_collection.documents(tag_key="user_id", tag_value=user_id):
            workflow_data = doc_ref.get()
            if workflow_data:
                # Filter by status if specified
                if status is None or workflow_data.get("status") == status:
                    workflows.append({
                        "id": workflow_data["session_id"],
                        "name": workflow_data["workflow_name"],
                        "description": workflow_data.get("description", ""),
                        "created": workflow_data["created_at"],
                        "updated": workflow_data["updated_at"],
                        "status": workflow_data["status"],
                        "progress": workflow_data.get("progress", 0),
                        "task_count": len(workflow_data.get("task_instances", []))
                    })
        
        # Sort by most recently updated
        workflows.sort(key=lambda w: w["updated"], reverse=True)
        return workflows
    
    def load_workflow(self, session_id: str) -> Optional[Dict]:
        """Load a workflow session for resumption."""
        doc_ref = self.sessions_collection.document(session_id)
        return doc_ref.get()
    
    def update_workflow_progress(self, session_id: str, progress: int, task_instances: List[Dict] = None):
        """Update workflow progress and task instances."""
        doc_ref = self.sessions_collection.document(session_id)
        workflow_data = doc_ref.get()
        
        if workflow_data:
            workflow_data["progress"] = progress
            workflow_data["updated_at"] = datetime.datetime.now().isoformat()
            
            if task_instances:
                workflow_data["task_instances"] = task_instances
            
            doc_ref.set(workflow_data)
    
    def complete_workflow(self, session_id: str, results: Dict):
        """Mark workflow as completed with final results."""
        doc_ref = self.sessions_collection.document(session_id)
        workflow_data = doc_ref.get()
        
        if workflow_data:
            workflow_data["status"] = "completed"
            workflow_data["progress"] = 100
            workflow_data["results"] = results
            workflow_data["completed_at"] = datetime.datetime.now().isoformat()
            workflow_data["updated_at"] = datetime.datetime.now().isoformat()
            
            doc_ref.set(workflow_data)
            
            # Update status tag
            doc_ref.tag("status", "completed")
    
    def delete_workflow(self, session_id: str):
        """Delete a workflow session."""
        doc_ref = self.sessions_collection.document(session_id)
        doc_ref.delete()

# Usage examples
def start_document_processing_workflow(user_id: str, workflow_name: str, document_ids: List[str]):
    """Start a new document processing workflow with persistent session."""
    
    # Initialize session manager
    session_manager = PersistentTaskSessionManager()
    
    # Create persistent workflow
    session_id = session_manager.create_user_workflow(
        user_id=user_id,
        workflow_name=workflow_name,
        description=f"Processing {len(document_ids)} documents"
    )
    
    print(f"Created workflow session: {session_id}")
    
    # Execute tasks within named session
    task_session_name = f"workflow_{session_id}"
    
    with Session(name=task_session_name) as task_session:
        futures = []
        task_instances = []
        
        # Start document processing tasks
        for i, doc_id in enumerate(document_ids):
            instance = document_processing_task.instance()
            future = instance.start(doc_id, "extract_text")
            futures.append(future)
            
            # Track task instance metadata
            task_instances.append({
                "task_id": instance.id,
                "document_id": doc_id,
                "operation": "extract_text",
                "status": "running",
                "started_at": datetime.datetime.now().isoformat()
            })
        
        # Update progress as tasks complete
        completed_tasks = 0
        results = []
        
        for i, future in enumerate(futures):
            result = future.result()
            results.append(result)
            completed_tasks += 1
            
            # Update task instance status
            task_instances[i]["status"] = "completed"
            task_instances[i]["completed_at"] = datetime.datetime.now().isoformat()
            
            # Update workflow progress
            progress = int((completed_tasks / len(document_ids)) * 100)
            session_manager.update_workflow_progress(session_id, progress, task_instances)
            
            print(f"Completed {completed_tasks}/{len(document_ids)} tasks ({progress}%)")
        
        # Mark workflow as completed
        final_results = {
            "processed_documents": len(results),
            "successful": sum(1 for r in results if r["status"] == "completed"),
            "results": results
        }
        
        session_manager.complete_workflow(session_id, final_results)
        print(f"Workflow {workflow_name} completed successfully!")
        
        return session_id, final_results

def resume_workflow(session_id: str):
    """Resume a previously saved workflow."""
    session_manager = PersistentTaskSessionManager()
    
    # Load workflow metadata
    workflow_data = session_manager.load_workflow(session_id)
    if not workflow_data:
        print(f"Workflow {session_id} not found")
        return
    
    print(f"Resuming workflow: {workflow_data['workflow_name']}")
    print(f"Status: {workflow_data['status']}")
    print(f"Progress: {workflow_data['progress']}%")
    print(f"Tasks: {len(workflow_data['task_instances'])}")
    
    # Resume logic here (if workflow supports resumption)
    # This would depend on your specific workflow requirements
    
    return workflow_data

def get_user_workflows_for_ui(user_id: str):
    """Get user's workflows for UI 'Load Session' feature."""
    session_manager = PersistentTaskSessionManager()
    
    # Get all workflows for user
    all_workflows = session_manager.get_user_workflows(user_id)
    
    # Get active workflows
    active_workflows = session_manager.get_user_workflows(user_id, status="active")
    
    # Get completed workflows
    completed_workflows = session_manager.get_user_workflows(user_id, status="completed")
    
    return {
        "all": all_workflows,
        "active": active_workflows,
        "completed": completed_workflows
    }

# Example usage in a web application
def document_processing_endpoint(user_id: str, workflow_name: str, document_ids: List[str]):
    """Web endpoint for starting document processing."""
    try:
        session_id, results = start_document_processing_workflow(
            user_id=user_id,
            workflow_name=workflow_name,
            document_ids=document_ids
        )
        
        return {
            "success": True,
            "session_id": session_id,
            "workflow_name": workflow_name,
            "processed_documents": results["processed_documents"],
            "message": f"Successfully processed {results['processed_documents']} documents"
        }
        
    except Exception as e:
        return {
            "success": False,
            "error": str(e),
            "message": "Failed to process documents"
        }

def load_session_endpoint(user_id: str):
    """Web endpoint for loading user's sessions (for UI dropdown)."""
    try:
        workflows = get_user_workflows_for_ui(user_id)
        
        return {
            "success": True,
            "workflows": workflows,
            "message": f"Found {len(workflows['all'])} workflows"
        }
        
    except Exception as e:
        return {
            "success": False,
            "error": str(e),
            "message": "Failed to load workflows"
        }
```

This pattern provides:

#### **Key Benefits**
- **Persistent Storage**: Workflows survive application restarts
- **User Filtering**: Efficient querying by user ID using collection tags
- **Progress Tracking**: Real-time updates to workflow progress
- **UI Integration**: Ready-to-use data structure for "Load Session" dropdowns
- **Metadata Rich**: Store comprehensive workflow information
- **Scalable**: Uses Numerous collections for distributed storage

#### **Collection Document Structure**
```json
{
  "session_id": "uuid-here",
  "user_id": "user123", 
  "workflow_name": "Document Analysis",
  "description": "Processing 50 documents",
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:15:00Z",
  "status": "completed",
  "progress": 100,
  "task_instances": [
    {
      "task_id": "task-uuid",
      "document_id": "doc1",
      "operation": "extract_text",
      "status": "completed",
      "started_at": "2024-01-01T10:00:00Z",
      "completed_at": "2024-01-01T10:01:00Z"
    }
  ],
  "results": {
    "processed_documents": 50,
    "successful": 48,
    "results": [...]
  }
}
```

#### **Collection Tags for Efficient Querying**
- `user_id`: Filter workflows by user
- `workflow_name`: Find workflows by name  
- `status`: Filter by active/completed/failed
- `workflow_type`: Group by workflow type

#### Pattern 3: Framework Integration with Session Distinction

Example showing both session types in a FastAPI application:

```python
from fastapi import FastAPI, Request, Depends
from numerous.tasks import Session, task, TaskControl
from numerous.frameworks.fastapi import get_session

app = FastAPI()

@task(max_parallel=5)
def process_user_document(tc: TaskControl, doc_id: str, user_id: str) -> dict:
    """Process a document for a specific user."""
    tc.log(f"Processing document {doc_id} for user {user_id}", "info")
    # Document processing logic
    return {"doc_id": doc_id, "user_id": user_id, "status": "processed"}

@app.post("/process-documents")
async def process_documents(
    request: Request,
    document_ids: List[str],
    task_session_name: str = None  # Optional: user can name their task session
):
    """Process multiple documents for the authenticated user."""
    
    # 1. Get user from HTTP/authentication session
    user_id = request.session.get("user_id")  # From HTTP session (User Session)
    if not user_id:
        return {"error": "Not authenticated"}
    
    # 2. Create or get task session name
    if not task_session_name:
        task_session_name = f"user_{user_id}_docs_{len(document_ids)}"
    
    # 3. Use framework-specific task session
    task_session = get_session(request)  # Gets/creates task session for this request
    
    # Alternative: Create explicitly named task session
    # with Session(name=task_session_name) as task_session:
    
    futures = []
    for doc_id in document_ids:
        instance = process_user_document.instance()
        future = instance.start(doc_id, user_id)
        futures.append(future)
    
    # Wait for all documents to be processed
    results = [future.result() for future in futures]
    
    return {
        "user_id": user_id,  # From user session
        "task_session_name": task_session_name,  # Task session identifier
        "processed_documents": len(results),
        "results": results
    }

@app.get("/my-task-sessions")
async def get_my_task_sessions(request: Request):
    """Get task sessions for the current user (for UI 'Load Session' dropdown)."""
    user_id = request.session.get("user_id")
    if not user_id:
        return {"error": "Not authenticated"}
    
    # Return user's task sessions for UI selection
    return get_user_workflows_for_ui(user_id)
```

### Task Session Lifecycle Management

```python
from numerous.tasks import Session, task, TaskControl

@task
def cleanup_task(tc: TaskControl, temp_files: List[str]) -> dict:
    """Clean up temporary files."""
    tc.log(f"Cleaning up {len(temp_files)} temporary files", "info")
    # Cleanup logic
    return {"cleaned_files": len(temp_files)}

class TaskSessionLifecycle:
    """Manage task session lifecycle with proper cleanup."""
    
    def __init__(self, session_name: str, user_id: str = None):
        self.session_name = session_name
        self.user_id = user_id
        self.task_session = None
    
    def __enter__(self):
        """Start the task session."""
        self.task_session = Session(name=self.session_name)
        return self.task_session.__enter__()
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Clean up the task session."""
        try:
            # Perform cleanup tasks
            if self.task_session:
                cleanup_instance = cleanup_task.instance()
                cleanup_future = cleanup_instance.start([])  # Files to clean
                cleanup_future.result()  # Wait for cleanup
                
        finally:
            # Close the task session
            if self.task_session:
                self.task_session.__exit__(exc_type, exc_val, exc_tb)

# Usage
with TaskSessionLifecycle("data_processing_session", user_id="user123") as session:
    # Your task processing here
    pass  # Session automatically cleaned up
```

### Best Practices for Session Management

#### 1. **Naming Conventions**
- **User Sessions**: Use framework conventions (`session`, `request.session`, etc.)
- **Task Sessions**: Use descriptive names that indicate purpose
  ```python
  # Good task session names
  "user_123_data_analysis_2024"
  "batch_processing_session_456"
  "document_processing_user_789"
  ```

#### 2. **Session Scope Guidelines**
- **User Sessions**: Authentication, user preferences, HTTP state
- **Task Sessions**: Task coordination, concurrency control, progress tracking
- **Don't mix**: Keep user authentication separate from task management

#### 3. **Persistence Patterns**
- **Short-lived tasks**: Tag task sessions with user session ID
- **Long-running workflows**: Use named task sessions with database persistence
- **UI-managed sessions**: Store task session metadata for user selection

#### 4. **Error Handling**
```python
from numerous.tasks import Session, TaskError

def safe_task_session_execution(session_name: str, user_id: str):
    """Execute tasks with proper error handling."""
    try:
        with Session(name=session_name) as task_session:
            # Task execution here
            pass
    except TaskError as e:
        # Log task-specific errors
        print(f"Task session {session_name} failed for user {user_id}: {e}")
        # Notify user or trigger recovery
    except Exception as e:
        # Handle unexpected errors
        print(f"Unexpected error in task session {session_name}: {e}")
        # Cleanup and notify
```

By clearly separating User Sessions (authentication/application state) from Task Sessions (task grouping/management), you can build robust applications that leverage both concepts effectively while avoiding confusion between the two distinct session types.

## Task Definition and Configuration

### Basic Task Definition

```python
from numerous.tasks import task, TaskControl

@task
def basic_task(value: int) -> int:
    """A simple task that doubles a value."""
    return value * 2

# With configuration
@task(name="configured_task", max_parallel=5, size="medium")
def configured_task(tc: TaskControl, data: list) -> dict:
    """A task with custom configuration."""
    tc.log("Processing data", "info")
    return {"processed": len(data), "items": data}
```

### Configuration Options

- **`name`**: Custom name for the task (default: function name)
- **`max_parallel`**: Maximum concurrent instances (default: 1)
- **`size`**: Resource size hint - "small", "medium", "large" (default: "small")

### Advanced Task Control

```python
@task
def advanced_task(tc: TaskControl, items: list) -> dict:
    """Task demonstrating all TaskControl features."""
    tc.log("Starting advanced task", "info")
    tc.update_status("Initializing")
    
    results = []
    total = len(items)
    
    for i, item in enumerate(items):
        # Cancellation check
        if tc.should_stop:
            tc.log("Task cancelled by user", "warning")
            break
            
        # Process item
        processed = process_item(item)
        results.append(processed)
        
        # Progress updates
        progress = (i + 1) / total * 100
        tc.update_progress(progress, f"Processed {i+1}/{total} items")
        
        # Structured logging
        if i % 10 == 0:
            tc.log(f"Checkpoint: processed {i+1} items", "info")
    
    tc.update_status("Completed")
    tc.log("Task finished successfully", "info")
    return {"processed": len(results), "results": results}
```

## Error Handling and Cancellation

### Graceful Cancellation

```python
from numerous.tasks import task, TaskControl
from numerous.tasks.exceptions import TaskCancelledError

@task
def cancellable_task(tc: TaskControl, duration: int) -> str:
    """Task that handles cancellation gracefully."""
    for i in range(duration):
        if tc.should_stop:
            tc.log("Graceful shutdown initiated", "warning")
            # Cleanup logic here
            raise TaskCancelledError("Task cancelled by user request")
        
        tc.update_progress(i / duration * 100, f"Step {i+1}/{duration}")
        time.sleep(1)
    
    return "Task completed successfully"

# Usage with cancellation
with Session() as session:
    instance = cancellable_task.instance()
    future = instance.start(30)
    
    # Cancel after 5 seconds
    import threading
    def cancel_later():
        time.sleep(5)
        instance.stop()
    
    threading.Thread(target=cancel_later).start()
    
    try:
        result = future.result()
        print(result)
    except TaskCancelledError:
        print("Task was cancelled")
```

### Error Recovery

```python
from numerous.tasks import task, TaskControl

@task
def resilient_task(tc: TaskControl, items: list, max_retries: int = 3) -> dict:
    """Task with built-in error recovery."""
    tc.log(f"Processing {len(items)} items with {max_retries} max retries", "info")
    
    results = []
    failures = []
    
    for i, item in enumerate(items):
        if tc.should_stop:
            break
            
        retries = 0
        while retries < max_retries:
            try:
                # Process item (might fail)
                result = risky_operation(item)
                results.append(result)
                break
            except Exception as e:
                retries += 1
                tc.log(f"Attempt {retries} failed for item {i}: {e}", "warning")
                
                if retries >= max_retries:
                    failures.append({"item": item, "error": str(e)})
                    tc.log(f"Max retries exceeded for item {i}", "error")
                else:
                    time.sleep(1)  # Wait before retry
        
        # Update progress
        progress = (i + 1) / len(items) * 100
        tc.update_progress(progress, f"Processed {i+1}/{len(items)} items")
    
    return {
        "successful": len(results),
        "failed": len(failures),
        "results": results,
        "failures": failures
    }
```

## Code Examples

Explore comprehensive examples in the repository:

- **[Basic Local Task Example](https://github.com/numerous-com/numerous-sdk/blob/main/python/examples/basic_local_task.py)**: Demonstrates task definition, execution, and session management
- **[Cancellation and Logging Example](https://github.com/numerous-com/numerous-sdk/blob/main/python/examples/cancellation_and_logging_task.py)**: Shows cancellation handling and structured logging
- **[Failing Task Example](https://github.com/numerous-com/numerous-sdk/blob/main/python/examples/failing_task_example.py)**: Demonstrates error handling and recovery patterns

## Common Patterns

### Data Pipeline Processing

```python
@task
def extract_data(tc: TaskControl, source: str) -> dict:
    """Extract data from source."""
    tc.log(f"Extracting data from {source}", "info")
    # Extraction logic
    return {"source": source, "data": extracted_data}

@task
def transform_data(tc: TaskControl, raw_data: dict) -> dict:
    """Transform extracted data."""
    tc.log("Transforming data", "info")
    # Transformation logic
    return {"transformed": raw_data}

@task
def load_data(tc: TaskControl, processed_data: dict) -> dict:
    """Load processed data to destination."""
    tc.log("Loading data", "info")
    # Loading logic
    return {"loaded": True, "records": len(processed_data)}

# Execute pipeline
with Session(name="etl-pipeline") as session:
    # Extract
    extract_instance = extract_data.instance()
    extract_future = extract_instance.start("database")
    raw_data = extract_future.result()
    
    # Transform
    transform_instance = transform_data.instance()
    transform_future = transform_instance.start(raw_data)
    processed_data = transform_future.result()
    
    # Load
    load_instance = load_data.instance()
    load_future = load_instance.start(processed_data)
    final_result = load_future.result()
```

### Parallel Batch Processing

```python
@task(max_parallel=4)
def process_batch(tc: TaskControl, batch_id: str, items: list) -> dict:
    """Process a batch of items in parallel."""
    tc.log(f"Processing batch {batch_id} with {len(items)} items", "info")
    
    results = []
    for i, item in enumerate(items):
        if tc.should_stop:
            break
            
        # Process individual item
        result = process_single_item(item)
        results.append(result)
        
        # Update progress
        progress = (i + 1) / len(items) * 100
        tc.update_progress(progress, f"Batch {batch_id}: {i+1}/{len(items)}")
    
    tc.log(f"Completed batch {batch_id}", "info")
    return {"batch_id": batch_id, "processed": len(results), "results": results}

# Process multiple batches concurrently
with Session(name="parallel-processing") as session:
    futures = []
    
    # Start up to 4 batches concurrently (max_parallel=4)
    for i, batch_data in enumerate(data_batches):
        instance = process_batch.instance()
        future = instance.start(f"batch_{i}", batch_data)
        futures.append(future)
    
    # Wait for all batches to complete
    results = []
    for future in futures:
        result = future.result()
        results.append(result)
    
    print(f"Processed {len(results)} batches")
```

## Best Practices

### 1. Task Design

- **Single Responsibility**: Each task should have a clear, single purpose
- **Idempotency**: Tasks should be safe to retry and produce consistent results
- **Statelessness**: Avoid shared state between task instances

### 2. Progress and Logging

- **Regular Progress Updates**: Use `tc.update_progress()` for long-running tasks
- **Structured Logging**: Use appropriate log levels and meaningful messages
- **Status Updates**: Keep users informed with `tc.update_status()`

### 3. Error Handling

- **Graceful Degradation**: Handle partial failures appropriately
- **Retry Logic**: Implement exponential backoff for transient failures
- **Cancellation Checks**: Regularly check `tc.should_stop` in loops

### 4. Resource Management

- **Appropriate Sizing**: Set correct `size` parameter for resource requirements
- **Concurrency Limits**: Configure `max_parallel` based on available resources
- **Session Management**: Use sessions to organize related tasks

### 5. Development Workflow

- **Start Simple**: Begin with direct execution for rapid development
- **Test Locally**: Use local task instances for integration testing
- **Deploy Incrementally**: Test thoroughly before switching to remote execution

## API Reference

### Core Classes

- **@task** - Decorator for defining tasks ([reference](reference/numerous/tasks/task.md))
- **Task** - Task definition class ([reference](reference/numerous/tasks/task.md))
- **TaskInstance** - Individual task execution instance ([reference](reference/numerous/tasks/task.md))
- **TaskControl** - Task control and monitoring ([reference](reference/numerous/tasks/control.md))
- **Session** - Task session management ([reference](reference/numerous/tasks/session.md))
- **Future** - Asynchronous result handling ([reference](reference/numerous/tasks/future.md))

### Task Control Methods

- **tc.log(message, level)** - Log messages with levels: "debug", "info", "warning", "error"
- **tc.update_progress(percentage, status)** - Update progress (0-100) and status message
- **tc.update_status(status)** - Update just the status message
- **tc.should_stop** - Check if task should stop (boolean property)

### Future Methods

- **future.result(timeout=None)** - Get result (blocks until complete)
- **future.status** - Current status ("pending", "running", "completed", "failed", "cancelled")
- **future.done** - Boolean indicating if task is complete
- **future.error** - Exception if task failed, None otherwise
- **future.cancel()** - Attempt to cancel the task

### Framework Integration

- **numerous.frameworks.fastapi.get_session(request)** - Get session for FastAPI applications ([reference](reference/numerous/frameworks/fastapi.md))
- **numerous.frameworks.streamlit.get_session()** - Get session for Streamlit applications ([reference](reference/numerous/frameworks/streamlit.md))

### Exceptions

- **TaskError** - Base task exception ([reference](reference/numerous/tasks/exceptions.md))
- **MaxInstancesReachedError** - Concurrency limit exceeded ([reference](reference/numerous/tasks/exceptions.md))
- **SessionNotFoundError** - No active session ([reference](reference/numerous/tasks/exceptions.md))
- **TaskCancelledError** - Task was cancelled ([reference](reference/numerous/tasks/exceptions.md))

## Advanced Topics

### Custom Task Control Handlers

For advanced use cases, customize how TaskControl operations are handled:

```python
from numerous.tasks.control import TaskControlHandler, set_task_control_handler

class CustomTaskControlHandler(TaskControlHandler):
    """Custom handler for task control operations."""
    
    def log(self, task_control, message, level, **extra_data):
        """Custom logging implementation."""
        print(f"[{level.upper()}] {task_control.task_definition_name}: {message}")
    
    def update_progress(self, task_control, progress, status):
        """Custom progress tracking."""
        print(f"Progress: {progress:.1f}% - {status}")
    
    def update_status(self, task_control, status):
        """Custom status updates."""
        print(f"Status: {status}")

# Set custom handler globally
set_task_control_handler(CustomTaskControlHandler())
```

### Backend Configuration

Control task execution environment:

```python
import os

# Local execution (default)
os.environ['NUMEROUS_TASK_BACKEND'] = 'local'

# Remote execution on Numerous platform
os.environ['NUMEROUS_TASK_BACKEND'] = 'remote'
```

### Task Metadata

Tasks can include metadata and documentation for better organization:

```python
from numerous.tasks import task

@task(name="data_processor", max_parallel=3, size="medium")
def data_processing_task(tc: TaskControl, data: dict) -> dict:
    """Process data with comprehensive logging and monitoring."""
    tc.log(f"Processing data batch with {len(data)} items", "info")
    tc.update_status("Initializing data processing")
    
    # Your processing logic here
    processed_data = {"processed": data, "timestamp": "2024-01-01"}
    
    tc.update_progress(100, "Processing complete")
    return processed_data
```

For complete API documentation, see the [API reference](reference/numerous/tasks/index.md). 