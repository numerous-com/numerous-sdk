import time
import logging
import threading # For demonstrating backend cancellation
from typing import Optional

from numerous.tasks import task, TaskControl, Session, Future
from numerous.tasks.exceptions import TaskCancelledError
from numerous.tasks.backends import get_backend, ExecutionBackend

# Configure basic logging to see output from the example script itself
# and from the numerous.tasks.control logger
logging.basicConfig(
    level=logging.INFO, 
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
# Optionally, set a more specific level for task logs if desired
# logging.getLogger("numerous.tasks.control").setLevel(logging.DEBUG)

@task(name="long_running_logger")
def long_running_logger_task(tc: TaskControl, duration_seconds: int) -> str:
    tc.log(f"Starting task for {duration_seconds} seconds.", level="info")
    
    for i in range(duration_seconds):
        if tc.should_stop:
            tc.log("Stop request received, attempting to terminate gracefully.", level="warning")
            tc.update_status("Cancelled by request")
            raise TaskCancelledError(f"Task {tc.instance_id} was cancelled.")
        
        tc.log(f"Working... {i+1}/{duration_seconds} seconds elapsed.", level="debug") # More verbose
        tc.update_progress(((i + 1) / duration_seconds) * 100, f"Progress: {i+1}s / {duration_seconds}s")
        time.sleep(1)

    result_message = f"Task completed after {duration_seconds} seconds."
    tc.log(result_message, level="info")
    tc.update_status("Completed successfully")
    return result_message

def demonstrate_task_initiated_cancellation():
    logging.info("\n--- Demonstrating Task-Initiated Cancellation (via should_stop) ---")
    with Session(name="cancellation_test_session_1") as session:
        task_instance = long_running_logger_task.instance()
        tc_obj = task_instance.task_control # Get the control object to simulate external stop request
        
        logging.info(f"Starting task {task_instance.id}...")
        future = task_instance.start(duration_seconds=10)

        # Let it run for a few seconds, then request stop via its TaskControl
        time.sleep(3)
        logging.info(f"Requesting stop for task {task_instance.id} directly via TaskControl object...")
        tc_obj.request_stop() # Task itself will check tc.should_stop

        try:
            result = future.result(timeout=15)
            logging.info(f"Task {task_instance.id} result: {result} (Status: {future.status})")
        except TaskCancelledError as e:
            logging.warning(f"Task {task_instance.id} was cancelled as expected: {e} (Status: {future.status})")
        except Exception as e:
            logging.error(f"Task {task_instance.id} failed unexpectedly: {e} (Status: {future.status}, Error: {future.error})")

def demonstrate_backend_initiated_cancellation():
    logging.info("\n--- Demonstrating Backend-Initiated Cancellation (via backend.cancel_task_instance) ---")
    # Ensure we are using the local backend for this example
    backend: Optional[ExecutionBackend] = get_backend("local")
    if backend is None:
        logging.error("Local backend not found, skipping backend cancellation demo.")
        return

    with Session(name="cancellation_test_session_2") as session:
        task_instance = long_running_logger_task.instance()
        instance_id_to_cancel = task_instance.id

        logging.info(f"Starting task {instance_id_to_cancel}...")
        future = task_instance.start(duration_seconds=10)

        # Let it run for a few seconds, then request stop via the backend
        def cancel_via_backend():
            time.sleep(3)
            logging.info(f"Requesting stop for task {instance_id_to_cancel} via backend.cancel_task_instance()...")
            cancelled_by_backend = backend.cancel_task_instance(instance_id_to_cancel)
            logging.info(f"Backend cancel_task_instance returned: {cancelled_by_backend}")

        cancel_thread = threading.Thread(target=cancel_via_backend)
        cancel_thread.start()

        try:
            result = future.result(timeout=15)
            logging.info(f"Task {instance_id_to_cancel} result: {result} (Status: {future.status})")
        except TaskCancelledError as e:
            logging.warning(f"Task {instance_id_to_cancel} was cancelled as expected: {e} (Status: {future.status})")
        except Exception as e:
            logging.error(f"Task {instance_id_to_cancel} failed unexpectedly: {e} (Status: {future.status}, Error: {future.error})")
        finally:
            cancel_thread.join()

def main():
    demonstrate_task_initiated_cancellation()
    demonstrate_backend_initiated_cancellation()

if __name__ == "__main__":
    main() 