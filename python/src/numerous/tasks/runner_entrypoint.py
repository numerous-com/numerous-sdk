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

from .control import set_task_control_handler, PoCMockRemoteTaskControlHandler # For PoC
# from .control import RemoteTaskControlHandler # In a real scenario
from .session import Session
from .exceptions import TaskError # Import base TaskError

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
        # Re-raise as RunnerError to be caught by the main try/except in main()
        raise RunnerError("TOML parsing library (tomli) not available.") 
    except Exception as e:
        # Catch other potential errors during tomli.load (e.g., TomlDecodeError)
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
        # Log critical because this will likely lead to script termination via exception.
        runner_logger.critical(f"Mandatory environment variable {name} not set.")
        raise RunnerError(f"Mandatory environment variable {name} not set.")
    return val

def main():
    runner_logger.info("Numerous Task Runner Entrypoint started.")

    # Define early for use in final logging/output, even if setup fails.
    task_instance_id_for_logging = os.environ.get("NUMEROUS_TASK_INSTANCE_ID", "unknown-instance-early")
    output_payload_path_str = os.environ.get("NUMEROUS_OUTPUT_PAYLOAD_PATH")

    result_payload: Optional[Dict[str, Any]] = None
    error_payload: Optional[Dict[str, Any]] = None
    task_final_status: str = "RUNNER_INIT_FAILURE" # Default status if errors occur before task logic.
    final_payload_to_write: Dict[str, Any] = {}

    try:
        # STAGE 1: Configuration and Setup
        # Expected env vars:
        #   NUMEROUS_TASK_INSTANCE_ID
        #   NUMEROUS_TASK_COLLECTION_NAME
        #   NUMEROUS_TASK_FUNCTION_NAME
        #   NUMEROUS_INPUT_PAYLOAD_PATH
        #   NUMEROUS_OUTPUT_PAYLOAD_PATH (already fetched for final output)
        #   NUMEROUS_MANIFEST_PATH
        #   NUMEROUS_MOCK_REMOTE_LOGGING (optional)

        task_instance_id = _get_env_var("NUMEROUS_TASK_INSTANCE_ID")
        task_instance_id_for_logging = task_instance_id # Update with actual ID if successfully read
        collection_name = _get_env_var("NUMEROUS_TASK_COLLECTION_NAME")
        task_function_name = _get_env_var("NUMEROUS_TASK_FUNCTION_NAME")
        input_payload_path_str = _get_env_var("NUMEROUS_INPUT_PAYLOAD_PATH")
        manifest_path_str = _get_env_var("NUMEROUS_MANIFEST_PATH")
        
        if output_payload_path_str is None: # Check explicitly after _get_env_var structure
             runner_logger.critical("NUMEROUS_OUTPUT_PAYLOAD_PATH environment variable not set.")
             raise RunnerError("NUMEROUS_OUTPUT_PAYLOAD_PATH environment variable not set.")
        output_payload_path = Path(output_payload_path_str)


        if _get_env_var("NUMEROUS_MOCK_REMOTE_LOGGING", False, "false").lower() == "true":
            runner_logger.info("Using PoCMockRemoteTaskControlHandler.")
            set_task_control_handler(PoCMockRemoteTaskControlHandler())
        else:
            runner_logger.info("Using default LocalTaskControlHandler.")
            set_task_control_handler(None) # Resets to default LocalTaskControlHandler

        runner_logger.info(f"Loading manifest from: {manifest_path_str}")
        manifest_data = _load_manifest(manifest_path_str)
        task_details = _find_task_details_in_manifest(manifest_data, task_function_name)
        
        source_file_rel_path = task_details.get("source_file")
        decorated_function_name = task_details.get("decorated_function", task_function_name)
        if not source_file_rel_path:
            raise RunnerError(f"Task '{task_function_name}' in manifest is missing 'source_file'.")

        source_file_path = Path(source_file_rel_path)
        if not source_file_path.is_file(): # Assumes CWD is the root of the extracted task package
            raise RunnerError(f"Source file '{source_file_path}' for task '{task_function_name}' not found (expected at CWD: {Path.cwd()}).")
        
        module_name = source_file_path.stem
        module_dir = source_file_path.parent.resolve()
        if str(module_dir) not in sys.path: sys.path.insert(0, str(module_dir))
        if str(Path.cwd()) not in sys.path: sys.path.insert(0, str(Path.cwd()))

        runner_logger.info(f"Importing module '{module_name}' from '{source_file_path}' (effective dir: {module_dir})")
        task_module = importlib.import_module(module_name)
        
        task_object = getattr(task_module, decorated_function_name, None)
        if task_object is None or not hasattr(task_object, 'instance'):
            raise RunnerError(f"Decorated function '{decorated_function_name}' not found or not a Numerous Task in '{source_file_path}'.")

        runner_logger.info(f"Loading input payload from: {input_payload_path_str}")
        input_payload_file = Path(input_payload_path_str)
        if not input_payload_file.is_file():
            raise RunnerError(f"Input payload file not found: {input_payload_path_str}")
        with open(input_payload_file, 'r') as f:
            inputs = json.load(f)
        
        task_kwargs = {}
        if isinstance(inputs, dict):
            task_kwargs = inputs
        elif isinstance(inputs, list):
            raise RunnerError("List inputs (positional args) not yet supported by PoC runner; use a JSON object for keyword args.")
        else:
            raise RunnerError("Input payload must be a JSON object (for kwargs).")
        runner_logger.info(f"Task inputs: {task_kwargs}")

        # STAGE 2: Task Execution
        runner_logger.info(f"Starting task execution for {collection_name}/{task_function_name} (Instance: {task_instance_id})")
        task_final_status = "TASK_EXECUTION_FAILURE" # Default if task execution block has issues
        try:
            with Session(name=f"task_runner_session_{task_instance_id}") as session:
                runner_logger.info(f"Created session: {session.id}")
                task_instance = task_object.instance()
                # TaskInstance now correctly passes its ID and task_definition.name to TaskControl
                runner_logger.info(f"Task instance {task_instance.id} (TC: {task_instance.task_control.instance_id}) created for {collection_name}/{task_function_name}.")
                
                future = task_instance.start(**task_kwargs)
                runner_logger.info(f"Task {task_instance.id} started. Waiting for result...")
                
                # TODO: Implement runner-level timeout based on manifest or invocation config.
                output = future.result() # This will re-raise task exceptions
                result_payload = {"result": output} # Ensure 'output' is JSON serializable by user's task design
                task_final_status = "COMPLETED"
                runner_logger.info(f"Task {task_instance.id} completed successfully.")

        except TaskError as te: # Includes TaskCancelledError
            runner_logger.error(f"Task {task_instance_id_for_logging} execution failed with TaskError: {te}", exc_info=True)
            error_payload = {"error_type": type(te).__name__, "message": str(te), "traceback": traceback.format_exc()}
            task_final_status = "FAILED" # Specific task failure
        except Exception as e: # Other exceptions during task runtime
            runner_logger.error(f"Task {task_instance_id_for_logging} execution failed with unexpected Exception: {e}", exc_info=True)
            error_payload = {"error_type": type(e).__name__, "message": str(e), "traceback": traceback.format_exc()}
            task_final_status = "FAILED" # Specific task failure due to unexpected error

        # STAGE 3: Prepare final output payload (based on STAGE 2 outcomes)
        if error_payload:
            final_payload_to_write = {"status": task_final_status, **error_payload}
        elif result_payload:
            final_payload_to_write = {"status": task_final_status, **result_payload}
        else: # Should ideally not be reached if task_final_status is COMPLETED/FAILED properly
            final_payload_to_write = {"status": task_final_status, "message": "Task execution block completed without explicit result or error payload."}
            if task_final_status not in ["COMPLETED", "FAILED"]: # If status is still TASK_EXECUTION_FAILURE
                 runner_logger.warning(f"Task {task_instance_id_for_logging} ended with status {task_final_status} and no specific result/error payload.")


    except RunnerError as re:
        runner_logger.critical(f"Runner script critical error (Instance: {task_instance_id_for_logging}): {re}", exc_info=False) # Keep traceback in logs but not always in output file
        error_payload = {"error_type": type(re).__name__, "message": str(re), "traceback": traceback.format_exc()} # Include traceback for runner errors
        task_final_status = "RUNNER_SETUP_FAILED"
        final_payload_to_write = {"status": task_final_status, **error_payload}
    except Exception as e: # Catch any other unexpected errors during setup
        runner_logger.critical(f"Unexpected critical error in runner (Instance: {task_instance_id_for_logging}): {e}", exc_info=True)
        error_payload = {"error_type": type(e).__name__, "message": str(e), "traceback": traceback.format_exc()}
        task_final_status = "RUNNER_UNEXPECTED_ERROR"
        final_payload_to_write = {"status": task_final_status, **error_payload}
    finally:
        # STAGE 4: Write output and exit (ALWAYS RUNS)
        if not final_payload_to_write: # Ensure there's always something to write
            final_payload_to_write = {
                "status": task_final_status if task_final_status else "UNKNOWN_FINAL_STATE",
                "message": "Runner reached final stage without a definitive payload.",
                "task_instance_id": task_instance_id_for_logging
            }
            if error_payload: # If an error_payload was set but somehow final_payload_to_write wasn't
                final_payload_to_write.update(error_payload)


        runner_logger.info(f"Writing final payload for instance {task_instance_id_for_logging} to {output_payload_path_str if output_payload_path_str else '<<NOT_SET>>'}: {final_payload_to_write}")
        
        if output_payload_path_str:
            try:
                with open(output_payload_path_str, 'w') as f:
                    json.dump(final_payload_to_write, f, indent=2)
                runner_logger.info(f"Successfully wrote output to {output_payload_path_str}")
            except Exception as write_err:
                runner_logger.error(f"FATAL: Could not write output payload to {output_payload_path_str} for instance {task_instance_id_for_logging}: {write_err}", exc_info=True)
                # Fallback: Try to print to stderr if file write fails
                print(json.dumps({"runner_file_write_error": str(write_err), **final_payload_to_write}), file=sys.stderr)
                if task_final_status not in ["RUNNER_SETUP_FAILED", "RUNNER_UNEXPECTED_ERROR", "FAILED"]:
                    task_final_status = "RUNNER_OUTPUT_FAILURE" # Specific status for output write failure
        else:
            runner_logger.error(f"NUMEROUS_OUTPUT_PAYLOAD_PATH was not set. Cannot write final JSON output to file for instance {task_instance_id_for_logging}. Dumping to stderr.")
            print(json.dumps({"runner_warning": "Output path not available", **final_payload_to_write}), file=sys.stderr)
            if task_final_status not in ["RUNNER_SETUP_FAILED", "RUNNER_UNEXPECTED_ERROR", "FAILED"]:
                 task_final_status = "RUNNER_OUTPUT_CONFIG_ERROR"


        exit_code = 0
        if task_final_status not in ["COMPLETED"]: # Any non-completed status is a failure for exit code
            exit_code = 1
        
        runner_logger.info(f"Numerous Task Runner Entrypoint for instance {task_instance_id_for_logging} finished with overall status: {task_final_status}. Exiting with code {exit_code}.")
        sys.exit(exit_code)

if __name__ == "__main__":
    # This allows `python -m numerous.tasks.runner_entrypoint`
    main() 