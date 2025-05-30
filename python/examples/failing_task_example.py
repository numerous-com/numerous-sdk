import time
import logging

from numerous.tasks import task, TaskControl, Session, Future
from numerous.tasks.exceptions import TaskError # For custom error type

logging.basicConfig(
    level=logging.INFO, 
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)

class MyCustomTaskError(TaskError):
    """A custom error specific to our failing task."""
    pass

@task(name="potentially_failing_task")
def failing_task_example(tc: TaskControl, succeed: bool, message: str = "") -> str:
    tc.log(f"Task started. Will it succeed? {succeed}. Message: '{message}'", level="info")
    
    time.sleep(1) # Simulate some work

    if not succeed:
        error_msg = f"Task deliberately failed with message: {message}"
        tc.log(error_msg, level="error")
        tc.update_status("Failed as requested")
        # Raise a custom or built-in error
        raise MyCustomTaskError(error_msg)
    
    success_msg = f"Task completed successfully with message: {message}"
    tc.log(success_msg, level="info")
    tc.update_status("Succeeded")
    return success_msg

def main():
    logging.info("\n--- Demonstrating Task Failure Handling ---")

    with Session(name="failure_test_session") as session:
        # Scenario 1: Task succeeds
        logging.info("\nScenario 1: Task is expected to succeed.")
        task_instance_success = failing_task_example.instance()
        future_success: Future[str] = task_instance_success.start(succeed=True, message="Hello from success")
        
        try:
            result = future_success.result(timeout=5)
            logging.info(f"Successful task result: '{result}'")
            logging.info(f"Future status: {future_success.status}")
            if future_success.error is None:
                logging.info("Future error: None (as expected for success)")
            else:
                logging.error(f"Future error: {future_success.error} (UNEXPECTED!)")
        except Exception as e:
            logging.error(f"Error waiting for successful task: {type(e).__name__} - {e}")

        # Scenario 2: Task fails
        logging.info("\nScenario 2: Task is expected to fail.")
        task_instance_fail = failing_task_example.instance()
        future_fail: Future[str] = task_instance_fail.start(succeed=False, message="Triggering failure")

        try:
            # Trying to get result will raise the exception
            result = future_fail.result(timeout=5) 
            logging.error(f"Failing task returned a result unexpectedly: '{result}'") # Should not happen
        except MyCustomTaskError as e:
            logging.warning(f"Caught expected MyCustomTaskError: {e}")
            logging.info(f"Future status: {future_fail.status}")
            if future_fail.error is not None:
                logging.info(f"Future error: {type(future_fail.error).__name__} - {future_fail.error} (as expected)")
            else:
                logging.error("Future error is None (UNEXPECTED for failure!)")
        except Exception as e:
            logging.error(f"Caught unexpected error from failing task: {type(e).__name__} - {e}")
            logging.info(f"Future status: {future_fail.status}, Error: {future_fail.error}")

    logging.info("\nTask failure demonstration finished.")

if __name__ == "__main__":
    main() 