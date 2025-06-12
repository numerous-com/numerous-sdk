"""
MockSessionManager for Task 1.0: Mock Backend Unit Testing.

Provides mock session management with in-memory state persistence:
- Tracks session lifecycle (create, active, cleanup) 
- Manages task instance associations with sessions
- Records all session operations for test inspection
- Simulates session state without external dependencies
"""

import time
import threading
from typing import Optional, Dict, List, Set, Any
from dataclasses import dataclass, field
from enum import Enum
import logging
from uuid import uuid4

logger = logging.getLogger(__name__)


class MockSessionState(Enum):
    """Mock session states."""
    CREATED = "created"
    ACTIVE = "active"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


@dataclass
class MockSessionInfo:
    """Information about a mock session."""
    session_id: str
    session_name: Optional[str]
    created_at: float
    started_at: Optional[float] = None
    completed_at: Optional[float] = None
    state: MockSessionState = MockSessionState.CREATED
    task_instances: Set[str] = field(default_factory=set)
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class MockSessionOperation:
    """Represents a session operation for tracking."""
    timestamp: float
    session_id: str
    operation: str  # 'create', 'start', 'add_task', 'remove_task', 'complete', etc.
    details: Dict[str, Any] = field(default_factory=dict)


class MockSessionManager:
    """
    Mock session manager for comprehensive unit testing.
    
    Features:
    - In-memory session state persistence
    - Full session lifecycle tracking
    - Task instance association management
    - Operation history for test assertions
    - Thread-safe operation for concurrent testing
    - Session cleanup and resource management simulation
    """
    
    def __init__(self):
        """Initialize the mock session manager."""
        self._sessions: Dict[str, MockSessionInfo] = {}
        self._operations: List[MockSessionOperation] = []
        self._current_session: Optional[str] = None
        self._lock = threading.Lock()
        
        # Statistics
        self._session_counter = 0
        self._total_task_instances = 0
        
        logger.info("MockSessionManager initialized")
    
    def create_session(self, session_name: Optional[str] = None, metadata: Optional[Dict[str, Any]] = None) -> str:
        """Create a new mock session."""
        with self._lock:
            self._session_counter += 1
            session_id = f"mock_session_{self._session_counter}_{uuid4().hex[:8]}"
            
            session_info = MockSessionInfo(
                session_id=session_id,
                session_name=session_name or f"Session {self._session_counter}",
                created_at=time.time(),
                metadata=metadata.copy() if metadata else {}
            )
            
            self._sessions[session_id] = session_info
            
            operation = MockSessionOperation(
                timestamp=time.time(),
                session_id=session_id,
                operation="create",
                details={
                    "session_name": session_name,
                    "metadata": metadata
                }
            )
            self._operations.append(operation)
            
            logger.info(f"MockSessionManager created session {session_id}")
            return session_id
    
    def start_session(self, session_id: str) -> bool:
        """Start (activate) a session."""
        with self._lock:
            session = self._sessions.get(session_id)
            if session is None:
                return False
            
            if session.state != MockSessionState.CREATED:
                return False
            
            session.state = MockSessionState.ACTIVE
            session.started_at = time.time()
            self._current_session = session_id
            
            operation = MockSessionOperation(
                timestamp=time.time(),
                session_id=session_id,
                operation="start"
            )
            self._operations.append(operation)
            
            logger.info(f"MockSessionManager started session {session_id}")
            return True
    
    def complete_session(self, session_id: str, success: bool = True) -> bool:
        """Complete a session."""
        with self._lock:
            session = self._sessions.get(session_id)
            if session is None:
                return False
            
            if session.state != MockSessionState.ACTIVE:
                return False
            
            session.state = MockSessionState.COMPLETED if success else MockSessionState.FAILED
            session.completed_at = time.time()
            
            if self._current_session == session_id:
                self._current_session = None
            
            operation = MockSessionOperation(
                timestamp=time.time(),
                session_id=session_id,
                operation="complete",
                details={"success": success}
            )
            self._operations.append(operation)
            
            logger.info(f"MockSessionManager completed session {session_id} (success={success})")
            return True
    
    def cancel_session(self, session_id: str) -> bool:
        """Cancel a session."""
        with self._lock:
            session = self._sessions.get(session_id)
            if session is None:
                return False
            
            if session.state in [MockSessionState.COMPLETED, MockSessionState.FAILED, MockSessionState.CANCELLED]:
                return False
            
            session.state = MockSessionState.CANCELLED
            session.completed_at = time.time()
            
            if self._current_session == session_id:
                self._current_session = None
            
            operation = MockSessionOperation(
                timestamp=time.time(),
                session_id=session_id,
                operation="cancel"
            )
            self._operations.append(operation)
            
            logger.info(f"MockSessionManager cancelled session {session_id}")
            return True
    
    def add_task_instance(self, session_id: str, instance_id: str) -> bool:
        """Add a task instance to a session."""
        with self._lock:
            session = self._sessions.get(session_id)
            if session is None or session.state != MockSessionState.ACTIVE:
                return False
            
            session.task_instances.add(instance_id)
            self._total_task_instances += 1
            
            operation = MockSessionOperation(
                timestamp=time.time(),
                session_id=session_id,
                operation="add_task",
                details={"instance_id": instance_id}
            )
            self._operations.append(operation)
            
            logger.debug(f"MockSessionManager added task {instance_id} to session {session_id}")
            return True
    
    def remove_task_instance(self, session_id: str, instance_id: str) -> bool:
        """Remove a task instance from a session."""
        with self._lock:
            session = self._sessions.get(session_id)
            if session is None:
                return False
            
            if instance_id in session.task_instances:
                session.task_instances.remove(instance_id)
                
                operation = MockSessionOperation(
                    timestamp=time.time(),
                    session_id=session_id,
                    operation="remove_task",
                    details={"instance_id": instance_id}
                )
                self._operations.append(operation)
                
                logger.debug(f"MockSessionManager removed task {instance_id} from session {session_id}")
                return True
            
            return False
    
    def get_current_session(self) -> Optional[str]:
        """Get the current active session ID."""
        with self._lock:
            return self._current_session
    
    def get_session_info(self, session_id: str) -> Optional[MockSessionInfo]:
        """Get information about a session."""
        with self._lock:
            session = self._sessions.get(session_id)
            return session
    
    def get_session_task_instances(self, session_id: str) -> Set[str]:
        """Get task instances associated with a session."""
        with self._lock:
            session = self._sessions.get(session_id)
            return session.task_instances.copy() if session else set()
    
    def get_active_sessions(self) -> List[MockSessionInfo]:
        """Get all active sessions."""
        with self._lock:
            return [s for s in self._sessions.values() if s.state == MockSessionState.ACTIVE]
    
    def get_all_sessions(self) -> List[MockSessionInfo]:
        """Get all sessions."""
        with self._lock:
            return list(self._sessions.values())
    
    def get_session_operations(self, session_id: Optional[str] = None) -> List[MockSessionOperation]:
        """Get session operations with optional filtering."""
        with self._lock:
            operations = self._operations.copy()
            
            if session_id is not None:
                operations = [op for op in operations if op.session_id == session_id]
            
            return operations
    
    def get_session_count(self) -> int:
        """Get total number of sessions created."""
        with self._lock:
            return len(self._sessions)
    
    def get_statistics(self) -> Dict[str, Any]:
        """Get comprehensive statistics about session usage."""
        with self._lock:
            states = {}
            for session in self._sessions.values():
                state = session.state.value
                states[state] = states.get(state, 0) + 1
            
            return {
                'total_sessions': len(self._sessions),
                'current_session': self._current_session,
                'session_states': states,
                'total_operations': len(self._operations),
                'total_task_instances': self._total_task_instances,
                'active_session_count': len([s for s in self._sessions.values() if s.state == MockSessionState.ACTIVE])
            }
    
    def clear_state(self) -> None:
        """Clear all session state (for test cleanup)."""
        with self._lock:
            self._sessions.clear()
            self._operations.clear()
            self._current_session = None
            self._session_counter = 0
            self._total_task_instances = 0
            
            logger.info("MockSessionManager state cleared")
    
    # Context manager support for Session integration
    
    def __enter__(self):
        """Support for 'with' statement (if needed)."""
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Cleanup on exit."""
        # Could auto-complete active sessions here if needed
        pass 