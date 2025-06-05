# numerous.tasks.runner_entrypoint
# This script is the entry point for executing a Numerous Task in a backend-managed environment.

import os
import sys
import importlib
import json
import logging
import traceback
from pathlib import Path
from typing import Dict, Any, Optional

from .control import set_task_control_handler, PoCMockRemoteTaskControlHandler, LocalTaskControlHandler # Ensure LocalTaskControlHandler is imported
from .session import Session
from .exceptions import TaskError # Import base TaskError
from ..organization import get_client # Added for API client
# Assuming organization is one level up. If numerous.tasks is top-level for runner, it might be: from numerous.organization import get_client

# Setup basic logging for the runner itself
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - [TaskRunner] %(message)s',
    stream=sys.stdout # Ensure runner logs go to stdout for capture by platform
)
runner_logger = logging.getLogger("numerous.task_runner")

# Expected environment variables or config file paths:
# MANDATORY:
#   NUMEROUS_TASK_INSTANCE_ID
#   NUMEROUS_TASK_COLLECTION_NAME (or identifier for manifest)
#   NUMEROUS_TASK_FUNCTION_NAME (public name of the task to run)
#   NUMEROUS_INPUT_PAYLOAD_PATH (path to JSON file with inputs)
#   NUMEROUS_OUTPUT_PAYLOAD_PATH (path where output JSON should be written)
#   NUMEROUS_MANIFEST_PATH (path to the numerous-task.toml for this collection)
#
# For PoC (mocking API calls by printing):
#   NUMEROUS_MOCK_REMOTE_LOGGING=true (or similar to trigger PoCMockRemoteTaskControlHandler)
#
# For Real Backend:
#   NUMEROUS_STATUS_API_ENDPOINT
#   NUMEROUS_LOG_API_ENDPOINT
#   NUMEROUS_RESULT_API_ENDPOINT
#   NUMEROUS_API_TOKEN_PATH (path to a file containing the API token)
#   NUMEROUS_SHOULD_STOP_SIGNAL_PATH (path to a file whose existence means stop)

class RunnerError(Exception):
    """Custom exception for runner script errors."""
    pass

def _load_manifest(manifest_path_str: str) -> Dict[str, Any]:
    manifest_path = Path(manifest_path_str)
    if not manifest_path.is_file():
        raise RunnerError(f"Manifest file not found: {manifest_path_str}")
    try:
        import tomli
        with open(manifest_path, "rb") as f:
            return tomli.load(f)
    except ImportError:
        runner_logger.error("tomli library not found. Please ensure it's in the task environment or requirements.txt for manifest parsing.")
        raise RunnerError("TOML parsing library (tomli) not available.") 
    except Exception as e:
        runner_logger.error(f"Error parsing manifest {manifest_path_str} with tomli: {e}", exc_info=True)
        raise RunnerError(f"Error parsing manifest {manifest_path_str}: {e}")

def _find_task_details_in_manifest(manifest_data: Dict[str, Any], task_function_name: str) -> Dict[str, Any]:
    for task_def in manifest_data.get("task", []):
        if task_def.get("function_name") == task_function_name:
            return task_def
    raise RunnerError(f"Task function '{task_function_name}' not found in manifest.")

def _get_env_var(name: str, is_mandatory: bool = True, default: Optional[str] = None) -> Optional[str]:
    val = os.environ.get(name)
    if val is None and default is not None:
        return default
    if val is None and is_mandatory:
        runner_logger.critical(f"Mandatory environment variable {name} not set.")
        raise RunnerError(f"Mandatory environment variable {name} not set.")
    return val

# Placeholder for fetching inputs via API
def _fetch_task_inputs(api_client: Any, task_instance_id: str) -> Dict[str, Any]:
    """Fetches task inputs using the API client."""
    runner_logger.info(f"Fetching inputs for task instance {task_instance_id} via API...")
    # Placeholder: Construct and execute GraphQL query
    # query = """
    #     query GetTaskInputs($id: UUID!) {
    #         getTaskInstance(id: $id) {
    #             inputs
    #         }
    #     }
    # """
    # variables = {"id": task_instance_id}
    # response = api_client.execute(query, variable_values=variables) # Conceptual
    # inputs_data = response.get("getTaskInstance", {}).get("inputs")
    # if inputs_data is None:
    #     raise RunnerError(f"Failed to fetch inputs for task instance {task_instance_id} or inputs are null.")
    # runner_logger.info(f"Successfully fetched inputs: {inputs_data}")
    # return inputs_data

    # --- Dummy implementation for now ---
    runner_logger.warning(f"DUMMY: _fetch_task_inputs called for {task_instance_id}. Returning dummy inputs.")
    # In a real scenario, this would come from an API call that might fail.
    # This dummy data should match what your actual task expects.
    # If the task 'add_numbers' expects {"a": 1, "b": 2}, return that.
    # If it's for a generic task, an empty dict might be okay for initial testing IF tasks handle it.
    return {"message": "This is a dummy input from runner_entrypoint._fetch_task_inputs"}
    # --- End Dummy implementation ---


# Placeholder for reporting outcome via API
def _report_task_outcome(api_client: Any, task_instance_id: str, status: str, result: Optional[Any] = None, error: Optional[Dict[str, Any]] = None) -> None:
    """Reports the final task outcome using the API client."""
    runner_logger.info(f"Reporting outcome for task instance {task_instance_id} via API: Status - {status}")
    # Placeholder: Construct and execute GraphQL mutation
    # mutation = """ ... ReportTaskOutcomeInput ... """
    # variables = { ... }
    # api_client.execute(mutation, variable_values=variables) # Conceptual
    # runner_logger.info(f"Successfully reported outcome for task instance {task_instance_id}.")

    # --- Dummy implementation for now ---
    runner_logger.warning(
        f"DUMMY: _report_task_outcome called for {task_instance_id}. Status: {status}. "
        f"Result: {result if result else 'N/A'}. Error: {error if error else 'N/A'}."
    )
    # --- End Dummy implementation ---


def main():
    # Define early for use in final logging/output, even if setup fails.
    # task_instance_id_for_logging will be properly set after reading env var.
    task_instance_id_for_logging = os.environ.get("NUMEROUS_TASK_INSTANCE_ID", "unknown-instance-early")
    # output_payload_path_str = os.environ.get("NUMEROUS_OUTPUT_PAYLOAD_PATH") # No longer used for primary output

    # Initialize these for the finally block
    api_client_for_reporting = None
    current_task_instance_id_for_reporting = task_instance_id_for_logging # Use early value if setup fails

    final_status_for_reporting: str = "RUNNER_INIT_FAILURE"
    final_result_for_reporting: Optional[Any] = None
    final_error_for_reporting: Optional[Dict[str, Any]] = None

    try:
        # STAGE 1: Configuration and Setup
        task_instance_id = _get_env_var("NUMEROUS_TASK_INSTANCE_ID")
        task_instance_id_for_logging = task_instance_id # Update with actual ID
        current_task_instance_id_for_reporting = task_instance_id # For use in finally block

        collection_name = _get_env_var("NUMEROUS_TASK_COLLECTION_NAME") # Still used for logging/context
        task_function_name = _get_env_var("NUMEROUS_TASK_FUNCTION_NAME")
        manifest_path_str = _get_env_var("NUMEROUS_MANIFEST_PATH")
        
        # API related env vars (used by get_client() implicitly, or explicitly if needed)
        _ = _get_env_var("NUMEROUS_API_URL") # Ensure it's set, get_client will use it
        _ = _get_env_var("NUMEROUS_API_ACCESS_TOKEN") # Ensure it's set, get_client will use it
        # NUMEROUS_ORGANIZATION_ID is also used by get_client, ensure it's set if your get_client requires it.
        _ = _get_env_var("NUMEROUS_ORGANIZATION_ID", is_mandatory=False) # Often optional for get_client if token is encompassing

        runner_logger.info(f"Numerous Task Runner Entrypoint started for instance: {task_instance_id}.")

        # Initialize API client
        try:
            api_client = get_client()
            api_client_for_reporting = api_client # For use in finally block
            runner_logger.info("Successfully initialized API client.")
        except Exception as e:
            runner_logger.critical(f"Failed to initialize API client: {e}", exc_info=True)
            raise RunnerError(f"API client initialization failed: {e}")

        # Fetch inputs using API client
        try:
            task_kwargs = _fetch_task_inputs(api_client, task_instance_id)
            runner_logger.info(f"Task inputs fetched (or dummied): {task_kwargs}")
        except Exception as e:
            runner_logger.error(f"Failed to fetch task inputs for {task_instance_id}: {e}", exc_info=True)
            # This is a critical error before task logic, set status for reporting
            final_status_for_reporting = "RUNNER_INPUT_FETCH_FAILED"
            final_error_for_reporting = {"error_type": type(e).__name__, "message": str(e), "traceback": traceback.format_exc()}
            raise # Re-raise to be caught by the main try-except and trigger final reporting

        # Load manifest and find task details (remains the same)
        runner_logger.info(f"Loading manifest from: {manifest_path_str}")
        manifest_data = _load_manifest(manifest_path_str)
        task_details = _find_task_details_in_manifest(manifest_data, task_function_name)
        
        source_file_rel_path = task_details.get("source_file")
        decorated_function_name = task_details.get("decorated_function", task_function_name)
        if not source_file_rel_path:
            raise RunnerError(f"Task '{task_function_name}' in manifest is missing 'source_file'.")

        # Setup TaskControlHandler (remains largely the same, uses api_client if needed by a real RemoteTaskControlHandler)
        use_mock_remote_handler_env = _get_env_var("NUMEROUS_MOCK_REMOTE_LOGGING", False, "false").lower() == "true"
        
        # Import task module and get task_object (remains the same, but done *before* handler setup to get expects_task_control)
        source_file_path = Path(source_file_rel_path)
        if not source_file_path.is_file():
            raise RunnerError(f"Source file '{source_file_path}' for task '{task_function_name}' not found (CWD: {Path.cwd()}).")
        
        module_name = source_file_path.stem
        module_dir = source_file_path.parent.resolve()
        if str(module_dir) not in sys.path: sys.path.insert(0, str(module_dir))
        if str(Path.cwd()) not in sys.path: sys.path.insert(0, str(Path.cwd()))

        runner_logger.info(f"Importing module '{module_name}' from '{source_file_path}' (effective dir: {module_dir})")
        task_module = importlib.import_module(module_name)
        task_object = getattr(task_module, decorated_function_name, None)
        if task_object is None or not hasattr(task_object, 'instance') or not hasattr(task_object, 'expects_task_control'):
            raise RunnerError(f"Decorated function '{decorated_function_name}' not found, not a Numerous Task, or missing 'expects_task_control' in '{source_file_path}'.")

        if task_object.expects_task_control and use_mock_remote_handler_env:
            runner_logger.info("Task expects TaskControl and mock remote logging is ON. Using PoCMockRemoteTaskControlHandler.")
            # PoCMockRemoteTaskControlHandler might also need api_client and task_instance_id for its own reporting in a real scenario
            set_task_control_handler(PoCMockRemoteTaskControlHandler(api_client=api_client, task_instance_id=task_instance_id))
        else:
            log_msg_handler = "Using default LocalTaskControlHandler because "
            if task_object.expects_task_control and not use_mock_remote_handler_env:
                log_msg_handler += "task expects TaskControl, but mock remote logging is OFF."
            elif not task_object.expects_task_control:
                log_msg_handler += "task does not expect TaskControl."
            else: # Should not happen
                 log_msg_handler += "of an unspecified condition."
            runner_logger.info(log_msg_handler)
            set_task_control_handler(LocalTaskControlHandler()) # Explicitly use Local

        # STAGE 2: Task Execution
        runner_logger.info(f"Starting task execution for {collection_name}/{task_function_name} (Instance: {task_instance_id})")
        final_status_for_reporting = "TASK_EXECUTION_FAILURE" # Default if this block has issues
        
        try:
            with Session(name=f"task_runner_session_{task_instance_id}") as session:
                runner_logger.info(f"Created session: {session.id}")
                task_instance = task_object.instance()
                runner_logger.info(f"Task instance {task_instance.id} created for {collection_name}/{task_function_name}. TC ID: {task_instance.task_control.instance_id if task_instance.task_control else 'N/A'}")
                
                future = task_instance.start(**task_kwargs)
                runner_logger.info(f"Task {task_instance.id} started. Waiting for result...")
                
                output = future.result() # This will re-raise task exceptions
                final_result_for_reporting = output
                final_status_for_reporting = "COMPLETED"
                runner_logger.info(f"Task {task_instance.id} completed successfully.")

        except TaskError as te:
            runner_logger.error(f"Task {task_instance_id} execution failed with TaskError: {te}", exc_info=True)
            final_error_for_reporting = {"error_type": type(te).__name__, "message": str(te), "traceback": traceback.format_exc()}
            final_status_for_reporting = "FAILED"
        except Exception as e:
            runner_logger.error(f"Task {task_instance_id} execution failed with unexpected Exception: {e}", exc_info=True)
            final_error_for_reporting = {"error_type": type(e).__name__, "message": str(e), "traceback": traceback.format_exc()}
            final_status_for_reporting = "FAILED"

    except RunnerError as re:
        runner_logger.critical(f"Runner script critical error (Instance: {task_instance_id_for_logging}): {re}", exc_info=False)
        final_error_for_reporting = {"error_type": type(re).__name__, "message": str(re), "traceback": traceback.format_exc()}
        final_status_for_reporting = "RUNNER_SETUP_FAILED"
    except Exception as e:
        runner_logger.critical(f"Unexpected critical error in runner (Instance: {task_instance_id_for_logging}): {e}", exc_info=True)
        final_error_for_reporting = {"error_type": type(e).__name__, "message": str(e), "traceback": traceback.format_exc()}
        final_status_for_reporting = "RUNNER_UNEXPECTED_ERROR"
    finally:
        # STAGE 3: Report outcome via API (replaces writing to output file)
        runner_logger.info(f"Reporting final outcome for instance {current_task_instance_id_for_reporting}. Status: {final_status_for_reporting}")
        if api_client_for_reporting:
            try:
                _report_task_outcome(
                    api_client_for_reporting,
                    current_task_instance_id_for_reporting,
                    status=final_status_for_reporting,
                    result=final_result_for_reporting,
                    error=final_error_for_reporting
                )
                runner_logger.info(f"Successfully reported outcome for {current_task_instance_id_for_reporting}.")
            except Exception as report_err:
                runner_logger.error(f"FATAL: Could not report outcome via API for {current_task_instance_id_for_reporting}: {report_err}", exc_info=True)
                # If API reporting fails, we might have to rely on platform to mark as failed after timeout, or log to stderr as last resort.
                final_status_for_reporting = "RUNNER_REPORTING_FAILURE" # Update status to reflect this specific failure
                print(json.dumps({
                    "runner_final_status_before_report_failure": final_status_for_reporting, # The status before this reporting error
                    "runner_reporting_error_type": type(report_err).__name__,
                    "runner_reporting_error_message": str(report_err),
                    "task_instance_id": current_task_instance_id_for_reporting,
                    "original_result": final_result_for_reporting, # Include original data if possible
                    "original_error": final_error_for_reporting
                }), file=sys.stderr)
        else:
            runner_logger.error(f"API client not available. Cannot report outcome for {current_task_instance_id_for_reporting}. Status was {final_status_for_reporting}")
            # Log to stderr as a last resort if API client was never initialized
            print(json.dumps({
                "runner_final_status": final_status_for_reporting,
                "runner_error": "API client not available for reporting",
                "task_instance_id": current_task_instance_id_for_reporting,
                "original_result": final_result_for_reporting,
                "original_error": final_error_for_reporting
            }), file=sys.stderr)

        exit_code = 0
        if final_status_for_reporting not in ["COMPLETED"]:
            exit_code = 1
        
        runner_logger.info(f"Numerous Task Runner Entrypoint for instance {current_task_instance_id_for_reporting} finished with reported status: {final_status_for_reporting}. Exiting with code {exit_code}.")
        sys.exit(exit_code)

if __name__ == "__main__":
    main() 