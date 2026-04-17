"""
Example usage of the functional task API.

Registering the task in numerous.toml:
[[tasks]]
  name = "Task Test"
   command = "numerous-executor task.py"  # File that contains the task function

Because we use the python task file, the platform will interpret this task
as a python task,
 and execute it by importing the file and look for the task with the name
 "Task Test".
If found it will execute the task in the Python interpreter.
"""

from __future__ import annotations

import time

from numerous.tasks import (
    get_task_controller,
    list_task_definitions,
    list_task_instances,
    task,
    wait_for_completion,
)
from numerous.tasks.task import request_stop


PROGRESS_THRESHOLD = 0.5


@task
def compute(x: int) -> int:
    """Perform example computation."""
    controller = get_task_controller()

    num_steps = 10
    for i in range(num_steps):
        time.sleep(0.1)

        # Update progress
        controller.set_progress(i / num_steps)

        # Check for stop signal
        if controller.should_stop():
            print("Task stopped by request")  # noqa: T201
            controller.set_output({"result": x})
            return x

    result = x + 1
    controller.set_output({"result": result})
    return result


def main() -> None:
    """Run the example."""
    # Run the task
    print("Starting task...")  # noqa: T201
    instance = compute(5)

    # Monitor progress
    while not instance.is_done():
        time.sleep(0.1)
        print(f"Progress: {instance.progress * 100:.1f}%")  # noqa: T201

        # Stop task if progress > threshold
        if instance.progress > PROGRESS_THRESHOLD:
            print("Requesting stop...")  # noqa: T201
            request_stop(instance)

    # Get result
    result = wait_for_completion(instance)
    print(f"Result: {result}")  # noqa: T201

    # List all tasks
    print("\nAll tasks:")  # noqa: T201
    for task_def in list_task_definitions():
        print(f"  - {task_def.name} (id: {task_def.id})")  # noqa: T201

    # List instances for this task
    print("\nTask instances:")  # noqa: T201
    for inst in list_task_instances():
        print(  # noqa: T201
            f"  - {inst.id}: {inst.status.value} (progress: {inst.progress * 100:.0f}%)"
        )


if __name__ == "__main__":
    main()
