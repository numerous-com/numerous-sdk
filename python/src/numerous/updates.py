"""Handling of updates sent to the tool session instance."""

import inspect
import logging
from types import MethodType
from typing import Any, Callable, Optional, Union

from numerous.generated.graphql import Updates
from numerous.generated.graphql.updates import (
    UpdatesToolSessionEventToolSessionActionTriggered,
    UpdatesToolSessionEventToolSessionElementUpdated,
    UpdatesToolSessionEventToolSessionElementUpdatedElementHTMLElement,
    UpdatesToolSessionEventToolSessionElementUpdatedElementNumberField,
    UpdatesToolSessionEventToolSessionElementUpdatedElementSliderElement,
    UpdatesToolSessionEventToolSessionElementUpdatedElementTextField,
)

log = logging.getLogger(__name__)


class ElementUpdateError(Exception):
    def __init__(self, element_name: str, instance_element_name: str) -> None:
        super().__init__(
            f"expected name {element_name} but instead found {instance_element_name}",
        )


class UpdateHandler:
    """Updates app instances according to events sent to the app session."""

    def __init__(self, instance: object) -> None:
        self._instance = instance
        self._update_handlers = _find_instance_update_handlers(instance)
        self._actions = self._find_actions(instance)

    def _find_actions(self, instance: object) -> dict[str, Callable[[], Any]]:
        methods = inspect.getmembers(instance, predicate=inspect.ismethod)
        return {
            name: method
            for name, method in methods
            if getattr(method, "__action__", True)
        }

    def handle_update(self, updates: Updates) -> None:
        """Handle an update for the tool session."""
        event = updates.tool_session_event
        if isinstance(event, UpdatesToolSessionEventToolSessionElementUpdated):
            self._handle_element_updated(event)
        elif isinstance(event, UpdatesToolSessionEventToolSessionActionTriggered):
            self._handle_action_triggered(event)
        else:
            log.info("unhandled event %s", event)

    def _handle_element_updated(
        self,
        event: UpdatesToolSessionEventToolSessionElementUpdated,
    ) -> None:
        element = event.element
        update_value = self._get_element_update_value(event)

        if update_value is not None:
            if self._naive_update_element(
                self._instance,
                element.id,
                element.name,
                update_value,
            ):
                log.debug(
                    "did not update element %s with value %s",
                    element,
                    update_value,
                )
        else:
            log.debug("unexpected update element %s", element)

        if element.id in self._update_handlers:
            log.debug("calling update handler for %s", element.name)
            self._update_handlers[element.id]()
        else:
            log.debug("no associated update handler for %s", element.name)

    def _get_element_update_value(
        self,
        event: UpdatesToolSessionEventToolSessionElementUpdated,
    ) -> Union[str, float, None]:
        element = event.element
        if isinstance(
            element,
            UpdatesToolSessionEventToolSessionElementUpdatedElementTextField,
        ):
            return element.text_value

        if isinstance(
            element,
            UpdatesToolSessionEventToolSessionElementUpdatedElementNumberField,
        ):
            return element.number_value

        if isinstance(
            element,
            UpdatesToolSessionEventToolSessionElementUpdatedElementHTMLElement,
        ):
            return element.html

        if isinstance(
            element,
            UpdatesToolSessionEventToolSessionElementUpdatedElementSliderElement,
        ):
            return element.slider_value

        return None

    def _handle_action_triggered(
        self,
        event: UpdatesToolSessionEventToolSessionActionTriggered,
    ) -> None:
        action = event.element
        if action.name in self._actions:
            self._actions[action.name]()
        else:
            log.warning("no action found for %r", action.name)

    def _naive_update_element(
        self,
        app_or_container: Any,  # noqa: ANN401
        element_id: str,
        element_name: str,
        value: Any,  # noqa: ANN401
    ) -> bool:
        element_ids_to_names: dict[str, str] = getattr(
            app_or_container,
            "__element_ids_to_names__",
            None,
        )  # type: ignore[assignment]
        if element_ids_to_names is None:
            return False

        if element_id in element_ids_to_names:
            instance_element_name = element_ids_to_names[element_id]
            if instance_element_name != element_name:
                raise ElementUpdateError(element_name, instance_element_name)
            app_or_container.__dict__[element_name] = value
            return True

        for child_name in element_ids_to_names.values():
            child = getattr(app_or_container, child_name)
            if self._naive_update_element(child, element_id, element_name, value):
                return True

        return False


def _find_instance_update_handlers(instance: object) -> dict[str, MethodType]:
    element_names_to_ids: Optional[dict[str, str]] = getattr(
        instance,
        "__element_names_to_ids__",
        None,
    )
    if element_names_to_ids is None:
        return {}

    methods = inspect.getmembers(instance, predicate=inspect.ismethod)
    element_names_to_handlers = {
        name.removesuffix("_updated"): method
        for name, method in methods
        if name.endswith("_updated")
    }

    element_ids_to_handlers = {
        element_names_to_ids[element_name]: handler
        for element_name, handler in element_names_to_handlers.items()
        if element_name in element_names_to_ids
    }

    for element_name in element_names_to_ids:
        child = getattr(instance, element_name, None)
        element_ids_to_handlers.update(_find_instance_update_handlers(child))

    return element_ids_to_handlers
