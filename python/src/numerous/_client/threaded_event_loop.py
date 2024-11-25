"""Separate Thread Manager for Asyncio Operations."""

import asyncio
import threading
import typing
from concurrent.futures import Future


T = typing.TypeVar("T")


class ThreadedEventLoop:
    """Wrapper for an asyncio event loop running in a thread."""

    def __init__(self) -> None:
        self._loop = asyncio.new_event_loop()
        self._thread = threading.Thread(
            target=self._run_loop_forever,
            name="Event Loop Thread",
            daemon=True,
        )

    def start(self) -> None:
        """Start the thread and run the event loop."""
        if not self._thread.is_alive():
            self._thread.start()

    def stop(self) -> None:
        """Stop the event loop, and terminate the thread."""
        if self._thread.is_alive():
            self._loop.stop()

    def await_coro(self, coroutine: typing.Awaitable[T]) -> T:
        """Awaiting for coroutine to finish."""
        f = self.schedule(coroutine)
        return f.result()

    def schedule(self, coroutine: typing.Awaitable[T]) -> Future[T]:
        """Schedule a coroutine in the event loop."""
        return asyncio.run_coroutine_threadsafe(coroutine, self._loop)

    def _run_loop_forever(self) -> None:
        asyncio.set_event_loop(self._loop)
        self._loop.run_forever()
