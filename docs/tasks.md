# Tasks

With Numerous Tasks, you can define and run background computations from your
app. Tasks run asynchronously, report progress, and can be stopped on request.

!!! tip
    Remember to add `numerous` as a dependency in your project; most likely to
    your `requirements.txt` file.

## Defining a task

Use the `@task` decorator to turn any Python function into a task. When the
decorated function is called, it submits the work for asynchronous execution
and immediately returns a `TaskInstanceState` object.

```py
from numerous.tasks import task

@task
def compute(x: int) -> int:
    return x + 1
```

You can also give the task a custom name:

```py
@task(name="My Computation")
def compute(x: int) -> int:
    return x + 1
```

## Running a task

Call the decorated function like a regular Python function. The arguments are
forwarded to the underlying function when it executes.

```py
instance = compute(5)
```

The returned `instance` is a `TaskInstanceState` that you can use to monitor
progress and retrieve the result.

## Monitoring progress

Poll `instance.is_done()` and read `instance.progress` (a float between `0.0`
and `1.0`) to track execution.

```py
import time

while not instance.is_done():
    time.sleep(0.1)
    print(f"Progress: {instance.progress * 100:.1f}%")
```

## Waiting for the result

Use `wait_for_completion` to block until the task finishes and get its return
value.

```py
from numerous.tasks import wait_for_completion

result = wait_for_completion(instance)
print(f"Result: {result}")
```

## Reporting progress from inside a task

Inside a task function, call `get_task_controller` to obtain a
`TaskController`. Use it to report progress and structured output.

```py
import time
from numerous.tasks import task, get_task_controller

@task
def compute(x: int) -> int:
    controller = get_task_controller()

    num_steps = 10
    for i in range(num_steps):
        time.sleep(0.1)
        controller.set_progress(i / num_steps)

    result = x + 1
    controller.set_output({"result": result})
    return result
```

`set_progress` accepts a float between `0.0` and `1.0`.
`set_output` accepts a dictionary that is stored on the task instance.

## Stopping a task

### Requesting a stop from outside the task

Call `request_stop` with the `TaskInstanceState` to signal the task that it
should stop.

```py
from numerous.tasks.task import request_stop

request_stop(instance)
```

### Checking the stop signal from inside the task

Inside the task function, call `controller.should_stop()` and return early when
it is `True`.

```py
from numerous.tasks import task, get_task_controller

@task
def compute(x: int) -> int:
    controller = get_task_controller()

    num_steps = 10
    for i in range(num_steps):
        controller.set_progress(i / num_steps)

        if controller.should_stop():
            print("Task stopped by request")
            controller.set_output({"result": x})
            return x

    result = x + 1
    controller.set_output({"result": result})
    return result
```

## Listing task definitions and instances

Use `list_task_definitions` to see every task that has been registered in the
current process, and `list_task_instances` to inspect past and current
executions.

```py
from numerous.tasks import list_task_definitions, list_task_instances

for task_def in list_task_definitions():
    print(f"{task_def.name} (id: {task_def.id})")

for inst in list_task_instances():
    print(f"{inst.id}: {inst.status.value} (progress: {inst.progress * 100:.0f}%)")
```

You can also list instances for a specific task only:

```py
instances = compute.list_instances()
```

## Running tasks on the platform

When your app is deployed to the Numerous Platform, tasks are executed by the
platform executor. No code changes are required; the executor is selected
automatically based on environment variables set by the platform.

See [Using the SDK locally](using_the_sdk_locally.md) for information about
how to configure the executor when running outside the platform.
