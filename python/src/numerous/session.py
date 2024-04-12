"""App sessions manages app instances."""

import asyncio
import logging
import random
import string
import threading
import typing
from concurrent.futures import Future
from typing import Any, Callable, Generic, Optional, Type, Union

from plotly import graph_objects as go

from numerous._plotly import plotly_html
from numerous.apps import HTML, Slider
from numerous.data_model import (
    ActionDataModel,
    AppDataModel,
    ContainerDataModel,
    ElementDataModel,
    HTMLElementDataModel,
    NumberFieldDataModel,
    PlotlyElementDataModel,
    SliderElementDataModel,
    TextFieldDataModel,
    dump_data_model,
)
from numerous.generated.graphql.fragments import GraphContextParent
from numerous.generated.graphql.input_types import ElementInput
from numerous.updates import UpdateHandler
from numerous.utils import AppT

from .generated.graphql.all_elements import (
    AllElementsSession,
    AllElementsSessionAllButton,
    AllElementsSessionAllElement,
    AllElementsSessionAllHTMLElement,
    AllElementsSessionAllNumberField,
    AllElementsSessionAllSliderElement,
    AllElementsSessionAllTextField,
)
from .generated.graphql.client import Client
from .generated.graphql.updates import (
    UpdatesToolSessionEventToolSessionElementAdded,
    UpdatesToolSessionEventToolSessionElementRemoved,
    UpdatesToolSessionEventToolSessionElementUpdated,
)

alphabet = string.ascii_lowercase + string.digits
ToolSessionEvent = Union[
    UpdatesToolSessionEventToolSessionElementAdded,
    UpdatesToolSessionEventToolSessionElementRemoved,
    UpdatesToolSessionEventToolSessionElementUpdated,
]
ToolSessionElement = Union[
    AllElementsSessionAllElement,
    AllElementsSessionAllButton,
    AllElementsSessionAllNumberField,
    AllElementsSessionAllTextField,
    AllElementsSessionAllHTMLElement,
    AllElementsSessionAllSliderElement,
]


AllElementsSession.model_rebuild()

log = logging.getLogger(__name__)


def get_client_id() -> str:
    return "".join(random.choices(alphabet, k=8))  # noqa: S311


class SessionElementTypeMismatchError(Exception):
    def __init__(self, sess_elem: ToolSessionElement, elem: ElementDataModel) -> None:
        super().__init__(
            f"{type(elem).__name__!r} does not match {type(sess_elem).__name__!r}",
        )


class SessionElementMissingError(Exception):
    def __init__(self, elem: ElementDataModel) -> None:
        super().__init__(f"Tool session missing required element '{elem.name}'")


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

    def schedule(self, coroutine: typing.Awaitable[typing.Any]) -> Future[Any]:
        """Schedule a coroutine in the event loop."""
        return asyncio.run_coroutine_threadsafe(coroutine, self._loop)

    def _run_loop_forever(self) -> None:
        asyncio.set_event_loop(self._loop)
        self._loop.run_forever()


class Session(Generic[AppT]):
    def __init__(
        self,
        session_id: str,
        client_id: str,
        instance: AppT,
        gql: Client,
    ) -> None:
        self._session_id = session_id
        self._client_id = client_id
        self._instance = instance
        self._gql = gql
        self._update_handler = UpdateHandler(instance)

    @staticmethod
    async def initialize(
        session_id: str,
        gql: Client,
        cls: Type[AppT],
    ) -> "Session[AppT]":
        """
        Initialize the session.

        Creates an instance, and validates it according to the remote app session.
        """
        threaded_event_loop = ThreadedEventLoop()
        threaded_event_loop.start()
        result = await gql.all_elements(session_id)
        client_id = result.session.client_id
        print(  # noqa: T201
            f"Running in session {session_id!r} as client {client_id!r}",
        )
        data_model = dump_data_model(cls)
        _validate_app_session(data_model, result.session)
        _wrap_app_class_setattr(threaded_event_loop, session_id, client_id, cls, gql)
        kwargs = _get_kwargs(cls, result.session)
        instance = cls(**kwargs)
        _add_element_ids(instance, None, result.session)
        _add_data_models(instance, data_model)
        return Session(session_id, client_id, instance, gql)

    async def run(self) -> None:
        """Run the app."""
        async for update in self._gql.updates(self._session_id, self._client_id):
            self._update_handler.handle_update(update)


def _add_element_ids(
    instance: AppT,
    element_id: Optional[str],
    session: AllElementsSession,
) -> None:
    names_to_ids = {}
    ids_to_names = {}

    for elem in session.all:
        if _element_parent_is(elem.graph_context.parent, element_id):
            names_to_ids[elem.name] = elem.id
            ids_to_names[elem.id] = elem.name

    instance.__element_names_to_ids__ = names_to_ids  # type: ignore[attr-defined]
    instance.__element_ids_to_names__ = ids_to_names  # type: ignore[attr-defined]


def _element_parent_is(
    graph_parent: Optional[GraphContextParent],
    parent_element_id: Optional[str],
) -> bool:
    graph_parent_id = graph_parent.id if graph_parent else None
    return (
        graph_parent_id is None and parent_element_id is None
    ) or graph_parent_id == parent_element_id


def _add_data_models(
    instance: object,
    data_model: Union[AppDataModel, ContainerDataModel],
) -> None:
    names_to_data_models = {}
    for el in data_model.elements:
        obj = getattr(instance, el.name, None)
        if obj is None:
            continue

        names_to_data_models[el.name] = el

        if isinstance(el, ContainerDataModel) and getattr(obj, "__container__", False):
            _add_data_models(obj, el)

        instance.__element_data_models__ = names_to_data_models  # type: ignore[attr-defined]


def _send_sync_update(
    loop: ThreadedEventLoop,
    gql: Client,
    session_id: str,
    client_id: str,
    element_input: ElementInput,
) -> None:
    done = threading.Event()

    async def update() -> None:
        try:
            await gql.update_element(
                session_id=session_id,
                client_id=client_id,
                element=element_input,
            )
        finally:
            done.set()

    loop.schedule(update())

    done.wait()


def get_setattr(
    loop: ThreadedEventLoop,
    gql: Client,
    old_setattr: Callable[[object, str, Any], None],
    session_id: str,
    client_id: str,
) -> Callable[[object, str, Any], None]:
    def setattr_override(
        instance: AppT,
        name: str,
        value: Any,  # noqa: ANN401
    ) -> None:
        old_setattr(instance, name, value)
        element_id = _get_element_id(instance, name)
        if element_id is None:
            return

        data_model = _get_data_model(instance, name)

        update_element_input = _get_setattr_value_update_input(
            data_model,
            value,
            element_id,
        )
        if update_element_input:
            _send_sync_update(loop, gql, session_id, client_id, update_element_input)
        else:
            log.debug(
                "could not update element id=%s name=%s value=%s",
                element_id,
                name,
                value,
            )

    return setattr_override


def _get_element_id(instance: AppT, name: str) -> Optional[str]:
    element_names_to_ids: Optional[dict[str, str]] = getattr(
        instance,
        "__element_names_to_ids__",
        None,
    )

    if element_names_to_ids is None:
        return None

    return element_names_to_ids.get(name)


def _get_data_model(instance: AppT, name: str) -> Optional[ElementDataModel]:
    element_names_to_data_models: Optional[dict[str, ElementDataModel]] = getattr(
        instance,
        "__element_data_models__",
        None,
    )
    if element_names_to_data_models is None:
        return None

    return element_names_to_data_models.get(name)


def _get_setattr_value_update_input(  # noqa: PLR0911
    data_model: Optional[ElementDataModel],
    value: Any,  # noqa: ANN401
    element_id: str,
) -> Optional[ElementInput]:
    if isinstance(value, str):
        if isinstance(data_model, HTMLElementDataModel):
            return ElementInput(elementID=element_id, htmlValue=value)
        return ElementInput(elementID=element_id, textValue=value)

    if isinstance(value, go.Figure) and isinstance(data_model, PlotlyElementDataModel):
        return ElementInput(elementID=element_id, htmlValue=plotly_html(value))

    if isinstance(value, (float, int)):
        if isinstance(data_model, SliderElementDataModel):
            return ElementInput(elementID=element_id, sliderValue=float(value))
        return ElementInput(
            elementID=element_id,
            numberValue=float(value),
        )

    if isinstance(data_model, Slider):
        return ElementInput(elementID=element_id, sliderValue=float(value))

    if isinstance(value, HTML):
        return ElementInput(
            elementID=element_id,
            htmlValue=value.default,
        )

    return None


def _wrap_app_class_setattr(
    loop: ThreadedEventLoop,
    session_id: str,
    client_id: str,
    cls: Type[AppT],
    gql: Client,
) -> None:
    annotations = _get_annotations(cls)
    for annotation in annotations.values():
        if not getattr(annotation, "__container__", False):
            continue
        _wrap_container_class_setattr(loop, session_id, client_id, cls, gql)
    new_setattr = get_setattr(loop, gql, cls.__setattr__, session_id, client_id)
    cls.__setattr__ = new_setattr  # type: ignore[assignment]


def _get_annotations(cls: type) -> dict[str, Any]:
    return cls.__annotations__ if hasattr(cls, "__annotations__") else {}


def _wrap_container_class_setattr(
    loop: ThreadedEventLoop,
    session_id: str,
    client_id: str,
    cls: Union[Type[AppT]],
    gql: Client,
) -> None:
    annotations = _get_annotations(cls)
    for annotation in annotations.values():
        if not getattr(annotation, "__container__", False):
            continue
        _wrap_container_class_setattr(loop, session_id, client_id, annotation, gql)
    new_setattr = get_setattr(loop, gql, cls.__setattr__, session_id, client_id)
    cls.__setattr__ = new_setattr  # type: ignore[method-assign, assignment]


def _validate_app_session(
    data_model: AppDataModel,
    session: AllElementsSession,
) -> None:
    session_elements = {sess_elem.name: sess_elem for sess_elem in session.all}
    for elem in data_model.elements:
        if elem.name not in session_elements:
            raise SessionElementMissingError(elem)
        sess_elem = session_elements[elem.name]
        _validate_element(elem, sess_elem)
    # TODO(jens): validate session elements exist in data model  # noqa: TD003, FIX002
    # TODO(jens): validate child elements  # noqa: TD003, FIX002


def _validate_element(elem: ElementDataModel, sess_elem: ToolSessionElement) -> None:
    valid_text_element = isinstance(elem, TextFieldDataModel) and isinstance(
        sess_elem,
        AllElementsSessionAllTextField,
    )
    valid_number_element = isinstance(elem, NumberFieldDataModel) and isinstance(
        sess_elem,
        AllElementsSessionAllNumberField,
    )
    valid_action_element = isinstance(elem, ActionDataModel) and isinstance(
        sess_elem,
        AllElementsSessionAllButton,
    )
    valid_html_element = isinstance(
        elem,
        (HTMLElementDataModel, PlotlyElementDataModel),
    ) and isinstance(
        sess_elem,
        AllElementsSessionAllHTMLElement,
    )
    valid_slider_element = isinstance(elem, SliderElementDataModel) and isinstance(
        sess_elem,
        AllElementsSessionAllSliderElement,
    )
    valid_container_element = (
        isinstance(elem, ContainerDataModel)
        and isinstance(sess_elem, AllElementsSessionAllElement)
        and sess_elem.typename__ == "Container"
    )
    if not (
        valid_text_element
        or valid_number_element
        or valid_action_element
        or valid_container_element
        or valid_html_element
        or valid_slider_element
    ):
        raise SessionElementTypeMismatchError(sess_elem, elem)


def _get_kwargs(cls: Type[AppT], session: AllElementsSession) -> dict[str, Any]:
    kwargs: dict[str, Any] = {}
    annotations = _get_annotations(cls)
    for element in session.all:
        if element.graph_context.parent is not None:
            continue
        if element.typename__ == "Button":
            continue
        kwargs[element.name] = _get_kwarg_value(
            annotations[element.name],
            session,
            element,
        )
    return kwargs


class SessionElementAnnotationMismatchError(Exception):
    def __init__(self, sess_elem: ToolSessionElement, annotation: type) -> None:
        sess_name = type(sess_elem).__name__
        ann_name = (
            annotation.__name__
            if type(annotation) is type
            else type(annotation).__name__
        )
        super().__init__(f"{sess_name!r} does not match annotation {ann_name!r}")


def _get_kwarg_value(
    annotation: type,
    session: AllElementsSession,
    element: ToolSessionElement,
) -> Any:  # noqa: ANN401
    if isinstance(element, AllElementsSessionAllNumberField) and annotation is float:
        return element.number_value

    if isinstance(element, AllElementsSessionAllTextField) and annotation is str:
        return element.text_value

    if isinstance(element, AllElementsSessionAllHTMLElement) and (
        annotation is str or annotation is go.Figure
    ):
        return element.html

    if isinstance(element, AllElementsSessionAllSliderElement) and annotation is float:
        return element.slider_value

    if (
        isinstance(element, AllElementsSessionAllElement)
        and element.typename__ == "Container"
        and getattr(annotation, "__container__", False)
    ):
        return _get_container_kwarg_value(
            annotation,
            session,
            element,
        )

    raise SessionElementAnnotationMismatchError(element, annotation)


class ContainerInitializeElementError(Exception):
    def __init__(self, container: AllElementsSessionAllElement) -> None:
        name = container.typename__
        super().__init__(f"cannot initialize non-container element {name!r}")


class ContainerInitializeAnnotationError(Exception):
    def __init__(self, container_cls: Type[AppT]) -> None:
        name = type(container_cls).__name__
        super().__init__(f"cannot initialize non-container annotation {name!r}")


def _get_container_kwarg_value(
    container_cls: Type[AppT],
    session: AllElementsSession,
    container: AllElementsSessionAllElement,
) -> Any:  # noqa: ANN401
    if container.typename__ != "Container":
        raise ContainerInitializeElementError(container)

    if not getattr(container_cls, "__container__", False):
        raise ContainerInitializeAnnotationError(container_cls)

    annotations = (
        container_cls.__annotations__
        if hasattr(container_cls, "__annotations__")
        else {}
    )

    children = {
        e.name: e
        for e in session.all
        if e.graph_context.parent and e.graph_context.parent.id == container.id
    }

    kwargs = {
        name: _get_kwarg_value(annotation, session, children[name])
        for name, annotation in annotations.items()
        if name in children
    }

    instance = container_cls(**kwargs)
    _add_element_ids(instance, container.id, session)
    return instance
