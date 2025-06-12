"""
MockTaskControlHandler for Task 1.0: Mock Backend Unit Testing.

Provides a comprehensive mock implementation of TaskControlHandler that:
- Records all log messages, progress updates, and status changes
- Simulates real TaskControl behavior with full state tracking  
- Allows inspection of all handler calls for test assertions
- Supports configurable behavior for different testing scenarios
"""

import time
import threading
from typing import Optional, Dict, List, Any
from dataclasses import dataclass, field
from enum import Enum
import logging

from numerous.tasks.control import TaskControlHandler, TaskControl

logger = logging.getLogger(__name__)


@dataclass
class MockLogEntry:
    """Represents a logged message with metadata."""
    timestamp: float
    task_id: str
    task_name: str
    message: str
    level: str
    extra_data: Dict[str, Any] = field(default_factory=dict)


@dataclass
class MockProgressUpdate:
    """Represents a progress update with metadata."""
    timestamp: float
    task_id: str
    task_name: str
    progress: float
    status: Optional[str]


@dataclass  
class MockStatusUpdate:
    """Represents a status update with metadata."""
    timestamp: float
    task_id: str
    task_name: str
    status: str


@dataclass
class MockStopRequest:
    """Represents a stop request with metadata."""
    timestamp: float
    task_id: str
    task_name: str


class MockTaskControlHandler(TaskControlHandler):
    """
    Mock TaskControl handler for comprehensive unit testing.
    
    Features:
    - Records all log messages with timestamps and metadata
    - Tracks progress updates with full history
    - Records status changes with timing information
    - Captures stop requests
    - Provides rich inspection API for test assertions
    - Thread-safe operation for concurrent testing
    """
    
    def __init__(self, simulate_delays: bool = False):
        """
        Initialize the mock handler.
        
        Args:
            simulate_delays: If True, add small delays to simulate real handler behavior
        """
        self._log_entries: List[MockLogEntry] = []
        self._progress_updates: List[MockProgressUpdate] = []
        self._status_updates: List[MockStatusUpdate] = []
        self._stop_requests: List[MockStopRequest] = []
        self._simulate_delays = simulate_delays
        self._lock = threading.Lock()
        
        # Statistics tracking
        self._call_counts: Dict[str, int] = {
            'log': 0,
            'update_progress': 0,
            'update_status': 0,
            'request_stop': 0
        }
        
        logger.info("MockTaskControlHandler initialized")
    
    def log(self, task_control: TaskControl, message: str, level: str, **extra_data: Any) -> None:
        """Record a log message with full metadata."""
        with self._lock:
            if self._simulate_delays:
                time.sleep(0.001)  # 1ms delay to simulate I/O
            
            entry = MockLogEntry(
                timestamp=time.time(),
                task_id=task_control.instance_id,
                task_name=task_control.task_definition_name,
                message=message,
                level=level,
                extra_data=extra_data.copy() if extra_data else {}
            )
            
            self._log_entries.append(entry)
            self._call_counts['log'] += 1
            
            logger.debug(f"MockHandler logged: [{level.upper()}] {task_control.task_definition_name}/{task_control.instance_id}: {message}")
    
    def update_progress(self, task_control: TaskControl, progress: float, status: Optional[str]) -> None:
        """Record a progress update with metadata."""
        with self._lock:
            if self._simulate_delays:
                time.sleep(0.001)  # 1ms delay to simulate I/O
            
            update = MockProgressUpdate(
                timestamp=time.time(),
                task_id=task_control.instance_id,
                task_name=task_control.task_definition_name,
                progress=progress,
                status=status
            )
            
            self._progress_updates.append(update)
            self._call_counts['update_progress'] += 1
            
            logger.debug(f"MockHandler progress: {task_control.task_definition_name}/{task_control.instance_id}: {progress}% - {status}")
    
    def update_status(self, task_control: TaskControl, status: str) -> None:
        """Record a status update with metadata."""
        with self._lock:
            if self._simulate_delays:
                time.sleep(0.001)  # 1ms delay to simulate I/O
            
            update = MockStatusUpdate(
                timestamp=time.time(),
                task_id=task_control.instance_id,
                task_name=task_control.task_definition_name,
                status=status
            )
            
            self._status_updates.append(update)
            self._call_counts['update_status'] += 1
            
            logger.debug(f"MockHandler status: {task_control.task_definition_name}/{task_control.instance_id}: {status}")
    
    def request_stop(self, task_control: TaskControl) -> None:
        """Record a stop request and set the internal stop flag."""
        with self._lock:
            if self._simulate_delays:
                time.sleep(0.001)  # 1ms delay to simulate I/O
            
            request = MockStopRequest(
                timestamp=time.time(),
                task_id=task_control.instance_id,
                task_name=task_control.task_definition_name
            )
            
            self._stop_requests.append(request)
            self._call_counts['request_stop'] += 1
            
            # Set the stop flag (this is the actual behavior)
            task_control._should_stop_internal = True
            
            logger.debug(f"MockHandler stop requested: {task_control.task_definition_name}/{task_control.instance_id}")
    
    # Inspection methods for testing
    
    def get_log_entries(self, task_id: Optional[str] = None, level: Optional[str] = None) -> List[MockLogEntry]:
        """Get log entries with optional filtering."""
        with self._lock:
            entries = self._log_entries.copy()
            
            if task_id is not None:
                entries = [e for e in entries if e.task_id == task_id]
            
            if level is not None:
                entries = [e for e in entries if e.level.lower() == level.lower()]
            
            return entries
    
    def get_progress_updates(self, task_id: Optional[str] = None) -> List[MockProgressUpdate]:
        """Get progress updates with optional filtering."""
        with self._lock:
            updates = self._progress_updates.copy()
            
            if task_id is not None:
                updates = [u for u in updates if u.task_id == task_id]
            
            return updates
    
    def get_status_updates(self, task_id: Optional[str] = None) -> List[MockStatusUpdate]:
        """Get status updates with optional filtering."""
        with self._lock:
            updates = self._status_updates.copy()
            
            if task_id is not None:
                updates = [u for u in updates if u.task_id == task_id]
            
            return updates
    
    def get_stop_requests(self, task_id: Optional[str] = None) -> List[MockStopRequest]:
        """Get stop requests with optional filtering."""
        with self._lock:
            requests = self._stop_requests.copy()
            
            if task_id is not None:
                requests = [r for r in requests if r.task_id == task_id]
            
            return requests
    
    def get_call_counts(self) -> Dict[str, int]:
        """Get call counts for all handler methods."""
        with self._lock:
            return self._call_counts.copy()
    
    def get_latest_progress(self, task_id: str) -> Optional[MockProgressUpdate]:
        """Get the latest progress update for a task."""
        updates = self.get_progress_updates(task_id)
        return updates[-1] if updates else None
    
    def get_latest_status(self, task_id: str) -> Optional[MockStatusUpdate]:
        """Get the latest status update for a task."""
        updates = self.get_status_updates(task_id)
        return updates[-1] if updates else None
    
    def get_latest_log(self, task_id: str, level: Optional[str] = None) -> Optional[MockLogEntry]:
        """Get the latest log entry for a task."""
        entries = self.get_log_entries(task_id, level)
        return entries[-1] if entries else None
    
    def has_logged_message(self, task_id: str, message_substring: str, level: Optional[str] = None) -> bool:
        """Check if a message containing substring was logged."""
        entries = self.get_log_entries(task_id, level)
        return any(message_substring in entry.message for entry in entries)
    
    def get_task_progress_history(self, task_id: str) -> List[float]:
        """Get progress value history for a task."""
        updates = self.get_progress_updates(task_id)
        return [u.progress for u in updates]
    
    def get_task_status_history(self, task_id: str) -> List[str]:
        """Get status history for a task."""
        updates = self.get_status_updates(task_id)
        return [u.status for u in updates]
    
    def was_stop_requested(self, task_id: str) -> bool:
        """Check if stop was requested for a task."""
        return len(self.get_stop_requests(task_id)) > 0
    
    def clear_history(self) -> None:
        """Clear all recorded history (for test cleanup)."""
        with self._lock:
            self._log_entries.clear()
            self._progress_updates.clear()
            self._status_updates.clear()
            self._stop_requests.clear()
            
            for key in self._call_counts:
                self._call_counts[key] = 0
            
            logger.info("MockTaskControlHandler history cleared")
    
    def get_statistics(self) -> Dict[str, Any]:
        """Get comprehensive statistics about handler usage."""
        with self._lock:
            return {
                'total_calls': sum(self._call_counts.values()),
                'call_counts': self._call_counts.copy(),
                'total_log_entries': len(self._log_entries),
                'total_progress_updates': len(self._progress_updates),
                'total_status_updates': len(self._status_updates),
                'total_stop_requests': len(self._stop_requests),
                'unique_tasks': len(set(e.task_id for e in self._log_entries + 
                                      [MockLogEntry(0, u.task_id, u.task_name, '', '') for u in self._progress_updates] +
                                      [MockLogEntry(0, u.task_id, u.task_name, '', '') for u in self._status_updates] +
                                      [MockLogEntry(0, r.task_id, r.task_name, '', '') for r in self._stop_requests])),
                'log_levels': list(set(e.level for e in self._log_entries))
            } 