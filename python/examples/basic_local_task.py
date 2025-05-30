import time
import logging
from numerous.tasks import task, TaskControl, Session, Future
from numerous.tasks.exceptions import TaskCancelledError, MaxInstancesReachedError

# Configure basic logging to see output
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

@task(name="my_simple_adder", max_parallel=2)
def simple_adder(tc: TaskControl, x: int, y: int) -> int:
    """
    A simple task that adds two numbers and demonstrates TaskControl.
    """
    logging.info(f"Task '{tc.task_definition_name}' instance {tc.instance_id} started with x={x}, y={y}")
    
    total_steps = 10
    for i in range(total_steps):
        if tc.should_stop:
            tc.update_status("Stopping due to request.")
            logging.info(f"Task instance {tc.instance_id} stopping early.")
            raise TaskCancelledError(f"Task {tc.instance_id} cancelled by stop request")
        
        time.sleep(0.2) # Simulate work
        progress = ((i + 1) / total_steps) * 100
        tc.update_progress(progress, f"Step {i+1} of {total_steps} complete.")
        logging.info(f"Task instance {tc.instance_id} progress: {progress:.0f}%")

    result = x + y
    tc.update_status(f"Addition complete: {x} + {y} = {result}")
    logging.info(f"Task instance {tc.instance_id} finished. Result: {result}")
    return result

def main():
    logging.info("Starting basic task example...")

    # Sessions manage task instances and their context
    with Session(name="my_example_session") as session:
        logging.info(f"Active session: {session.name} (ID: {session.id})")

        # Create task instances from the task definition
        # The .instance() method requires an active session
        task_instance1 = simple_adder.instance()
        task_instance2 = simple_adder.instance()

        logging.info(f"Created task instance 1: {task_instance1.id} for task '{task_instance1.task_definition_name}'")
        logging.info(f"Created task instance 2: {task_instance2.id} for task '{task_instance2.task_definition_name}'")

        # Start the tasks
        # .start() returns a Future object immediately
        logging.info(f"Starting task instance 1 with inputs (5, 3)...")
        future1: Future[int] = task_instance1.start(5, 3)
        
        logging.info(f"Starting task instance 2 with inputs (10, 7)...")
        future2: Future[int] = task_instance2.start(10, 7)

        logging.info(f"Task 1 status after start: {future1.status}")
        logging.info(f"Task 2 status after start: {future2.status}")

        # Wait for results (blocking)
        # The .result() method will block until the task is complete.
        # It will also raise any exceptions that occurred within the task.
        try:
            logging.info("Waiting for result from task instance 1...")
            result1 = future1.result(timeout=10) # Wait up to 10 seconds
            logging.info(f"Result from task instance 1: {result1} (Status: {future1.status})")
        except TimeoutError:
            logging.error("Task instance 1 timed out!")
        except Exception as e:
            logging.error(f"Task instance 1 failed: {e} (Status: {future1.status}, Error: {future1.error})")

        try:
            logging.info("Waiting for result from task instance 2...")
            result2 = future2.result() # Wait indefinitely
            logging.info(f"Result from task instance 2: {result2} (Status: {future2.status})")
        except Exception as e:
            logging.error(f"Task instance 2 failed: {e} (Status: {future2.status}, Error: {future2.error})")

        # Example of calling a task with max_parallel=1 directly (synchronous-like)
        # First, let's define such a task
        @task(name="synchronous_task_example", max_parallel=1)
        def quick_task(tc: TaskControl, message: str) -> str:
            logging.info(f"Quick task '{tc.task_definition_name}' (Instance: {tc.instance_id}) started with message: '{message}'")
            tc.update_progress(50, "Halfway")
            time.sleep(0.1)
            tc.update_progress(100, "Done")
            return f"Processed: {message}"

        try:
            logging.info("Running 'quick_task' directly (synchronous behavior)...")
            # This call will block because max_parallel is 1 for quick_task
            direct_result = quick_task("hello direct") 
            logging.info(f"Result from direct call to 'quick_task': {direct_result}")
        except Exception as e:
            logging.error(f"Direct call to 'quick_task' failed: {e}")
            
        # Example of attempting to exceed max_parallel for simple_adder (max_parallel=2)
        try:
            logging.info("Attempting to start a third instance of simple_adder (should fail due to max_parallel=2)...")
            task_instance3 = simple_adder.instance()
            future3 = task_instance3.start(1,1) # This should raise MaxInstancesReachedError
            # We shouldn't reach here if the error is raised correctly
            result3 = future3.result()
            logging.info(f"Result from task instance 3: {result3}")
        except MaxInstancesReachedError as e: # Be specific about the expected error
             logging.error(f"Error starting third instance as expected: {type(e).__name__} - {e}")
        except Exception as e: # Catch other unexpected errors
             logging.error(f"Unexpected error starting third instance: {type(e).__name__} - {e}")


    logging.info("Basic task example finished.")

if __name__ == "__main__":
    main() 