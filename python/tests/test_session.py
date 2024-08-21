from typing import AsyncIterator
from unittest.mock import Mock

import pytest
from numerous import action, app, container, html, slider
from numerous.generated.graphql import Client
from numerous.generated.graphql.all_elements import (
    AllElements,
    AllElementsSession,
    AllElementsSessionAllButton,
    AllElementsSessionAllButtonGraphContext,
    AllElementsSessionAllElement,
    AllElementsSessionAllElementGraphContext,
    AllElementsSessionAllHTMLElement,
    AllElementsSessionAllHTMLElementGraphContext,
    AllElementsSessionAllNumberField,
    AllElementsSessionAllNumberFieldGraphContext,
    AllElementsSessionAllSliderElement,
    AllElementsSessionAllSliderElementGraphContext,
    AllElementsSessionAllTextField,
    AllElementsSessionAllTextFieldGraphContext,
)
from numerous.generated.graphql.fragments import GraphContextParent
from numerous.generated.graphql.input_types import ElementInput
from numerous.generated.graphql.update_element import (
    UpdateElement,
    UpdateElementElementUpdate,
)
from numerous.generated.graphql.updates import (
    Updates,
    UpdatesToolSessionEventToolSessionActionTriggered,
    UpdatesToolSessionEventToolSessionActionTriggeredElement,
    UpdatesToolSessionEventToolSessionElementUpdated,
    UpdatesToolSessionEventToolSessionElementUpdatedElementHTMLElement,
    UpdatesToolSessionEventToolSessionElementUpdatedElementHTMLElementGraphContext,
    UpdatesToolSessionEventToolSessionElementUpdatedElementNumberField,
    UpdatesToolSessionEventToolSessionElementUpdatedElementNumberFieldGraphContext,
    UpdatesToolSessionEventToolSessionElementUpdatedElementSliderElement,
    UpdatesToolSessionEventToolSessionElementUpdatedElementSliderElementGraphContext,
    UpdatesToolSessionEventToolSessionElementUpdatedElementTextField,
    UpdatesToolSessionEventToolSessionElementUpdatedElementTextFieldGraphContext,
)
from numerous.session import Session, SessionElementMissingError


DEFAULT_SESSION_ID = "test_session_id"
DEFAULT_CLIENT_ID = "test_client_id"

TEXT_ELEMENT_ID = "text_element_id"
TEXT_ELEMENT_NAME = "text_param"
TEXT_ELEMENT = AllElementsSessionAllTextField(
    __typename="TextField",
    id=TEXT_ELEMENT_ID,
    graphContext=AllElementsSessionAllTextFieldGraphContext(
        parent=None,
        affectedBy=[],
        affects=[],
    ),
    name=TEXT_ELEMENT_NAME,
    textValue="text value",
)

NUMBER_ELEMENT_ID = "number_element_id"
NUMBER_ELEMENT_NAME = "number_param"
NUMBER_ELEMENT = AllElementsSessionAllNumberField(
    __typename="NumberField",
    id=NUMBER_ELEMENT_ID,
    graphContext=AllElementsSessionAllNumberFieldGraphContext(
        parent=None,
        affectedBy=[],
        affects=[],
    ),
    name=NUMBER_ELEMENT_NAME,
    numberValue=100.0,
)

SLIDER_ELEMENT_ID = "slider_element_id"
SLIDER_ELEMENT_NAME = "slider_param"
SLIDER_ELEMENT_MIN_VALUE = -200.0
SLIDER_ELEMENT_MAX_VALUE = 300.0
SLIDER_ELEMENT = AllElementsSessionAllSliderElement(
    __typename="SliderElement",
    id=SLIDER_ELEMENT_ID,
    graphContext=AllElementsSessionAllSliderElementGraphContext(
        parent=None,
        affectedBy=[],
        affects=[],
    ),
    name=SLIDER_ELEMENT_NAME,
    sliderValue=100.0,
    minValue=SLIDER_ELEMENT_MIN_VALUE,
    maxValue=SLIDER_ELEMENT_MAX_VALUE,
)

HTML_ELEMENT_ID = "html_element_id"
HTML_ELEMENT_NAME = "html_param"
HTML_ELEMENT = AllElementsSessionAllHTMLElement(
    __typename="HTMLElement",
    id=HTML_ELEMENT_ID,
    name=HTML_ELEMENT_NAME,
    html="",
    graphContext=AllElementsSessionAllHTMLElementGraphContext(
        parent=None,
        affectedBy=[],
        affects=[],
    ),
)

ACTION_ELEMENT_ID = "button_element_id"
ACTION_ELEMENT_NAME = "my_action"
ACTION_ELEMENT = AllElementsSessionAllButton(
    __typename="Button",
    id=ACTION_ELEMENT_ID,
    graphContext=AllElementsSessionAllButtonGraphContext(
        parent=None,
        affectedBy=[],
        affects=[],
    ),
    name=ACTION_ELEMENT_NAME,
    buttonValue="button value",
)

CONTAINER_ELEMENT_ID = "container_element_id"
CONTAINER_ELEMENT_NAME = "container"
CONTAINER_ELEMENT = AllElementsSessionAllElement(
    __typename="Container",
    id=CONTAINER_ELEMENT_ID,
    graphContext=AllElementsSessionAllElementGraphContext(
        parent=None,
        affectedBy=[],
        affects=[],
    ),
    name=CONTAINER_ELEMENT_NAME,
)

CONTAINER_CHILD_ELEMENT_ID = "container_child_element_id"
CONTAINER_CHILD_ELEMENT_NAME = "container_child"
CONTAINER_CHILD_ELEMENT = AllElementsSessionAllNumberField(
    __typename="NumberField",
    id=CONTAINER_CHILD_ELEMENT_ID,
    graphContext=AllElementsSessionAllNumberFieldGraphContext(
        parent=GraphContextParent(
            __typename="Container",
            id=CONTAINER_ELEMENT_ID,
        ),
        affectedBy=[],
        affects=[],
    ),
    name=CONTAINER_CHILD_ELEMENT_NAME,
    numberValue=100.0,
)


async def updates_mock_with_action_trigger(
    _session_id: str,
    _client_id: str,
) -> AsyncIterator[Updates]:
    yield Updates(
        toolSessionEvent=UpdatesToolSessionEventToolSessionActionTriggered(
            __typename="ToolSessionActionTriggered",
            element=UpdatesToolSessionEventToolSessionActionTriggeredElement(
                id=ACTION_ELEMENT_ID,
                name=ACTION_ELEMENT_NAME,
            ),
        ),
    )


@pytest.mark.asyncio
async def test_initialize_requests_element_information() -> None:
    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[TEXT_ELEMENT],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )

    @app
    class TestTool:
        text_param: str = "default"

    await Session.initialize(DEFAULT_SESSION_ID, gql, TestTool)

    gql.all_elements.assert_called_once_with(DEFAULT_SESSION_ID)


@pytest.mark.asyncio
async def test_initialize_fails_when_session_is_missing_parameter() -> None:
    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )

    @app
    class TestTool:
        text_param: str

    with pytest.raises(SessionElementMissingError) as exception_info:
        await Session.initialize(
            DEFAULT_SESSION_ID,
            gql,
            TestTool,
        )
    assert exception_info.value.args == (
        "Tool session missing required element 'text_param'",
    )


@pytest.mark.asyncio
async def test_initialize_fails_when_session_is_missing_action() -> None:
    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )

    @app
    class TestTool:
        @action
        def my_action(self) -> None:
            pass

    with pytest.raises(SessionElementMissingError) as exception_info:
        await Session.initialize(
            DEFAULT_SESSION_ID,
            gql,
            TestTool,
        )
    assert exception_info.value.args == (
        "Tool session missing required element 'my_action'",
    )


@pytest.mark.asyncio
async def test_initialize_fails_when_session_is_missing_container() -> None:
    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )

    @container
    class TestContainer:
        param: str

    @app
    class TestTool:
        container: TestContainer

    with pytest.raises(SessionElementMissingError) as exception_info:
        await Session.initialize(
            DEFAULT_SESSION_ID,
            gql,
            TestTool,
        )
    assert exception_info.value.args == (
        "Tool session missing required element 'container'",
    )


@pytest.mark.asyncio
async def test_initialize_fails_when_session_is_missing_element_in_container() -> None:
    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )

    @container
    class TestContainer:
        param: str

    @app
    class TestTool:
        container: TestContainer

    with pytest.raises(SessionElementMissingError) as exception_info:
        await Session.initialize(
            DEFAULT_SESSION_ID,
            gql,
            TestTool,
        )
    assert exception_info.value.args == (
        "Tool session missing required element 'container'",
    )


@pytest.mark.asyncio
async def test_text_element_update_updates_value_in_instance() -> None:
    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[TEXT_ELEMENT],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )

    async def updates_mock(
        _session_id: str,
        _client_id: str,
    ) -> AsyncIterator[Updates]:
        yield Updates(
            toolSessionEvent=UpdatesToolSessionEventToolSessionElementUpdated(
                __typename="ToolSessionElementUpdated",
                element=UpdatesToolSessionEventToolSessionElementUpdatedElementTextField(
                    __typename="TextField",
                    id=TEXT_ELEMENT_ID,
                    name=TEXT_ELEMENT_NAME,
                    textValue="new text value",
                    graphContext=UpdatesToolSessionEventToolSessionElementUpdatedElementTextFieldGraphContext(
                        parent=None,
                        affectedBy=[],
                        affects=[],
                    ),
                ),
            ),
        )

    gql.updates = updates_mock
    value_after_update = None

    @app
    class TestTool:
        text_param: str

        def text_param_updated(self) -> None:
            nonlocal value_after_update
            value_after_update = self.text_param

    session = await Session.initialize(
        DEFAULT_SESSION_ID,
        gql,
        TestTool,
    )
    await session.run()

    assert value_after_update == "new text value"


@pytest.mark.asyncio
async def test_html_element_update_updates_value_in_instance() -> None:
    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[HTML_ELEMENT],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )

    async def updates_mock(
        _session_id: str,
        _client_id: str,
    ) -> AsyncIterator[Updates]:
        yield Updates(
            toolSessionEvent=UpdatesToolSessionEventToolSessionElementUpdated(
                __typename="ToolSessionElementUpdated",
                element=UpdatesToolSessionEventToolSessionElementUpdatedElementHTMLElement(
                    __typename="HTMLElement",
                    id=HTML_ELEMENT_ID,
                    name=HTML_ELEMENT_NAME,
                    html="<div>updated html!</div>",
                    graphContext=UpdatesToolSessionEventToolSessionElementUpdatedElementHTMLElementGraphContext(
                        parent=None,
                        affectedBy=[],
                        affects=[],
                    ),
                ),
            ),
        )

    gql.updates = updates_mock
    value_after_update = None

    @app
    class TestTool:
        html_param: str = html()

        def html_param_updated(self) -> None:
            nonlocal value_after_update
            value_after_update = self.html_param

    session = await Session.initialize(
        DEFAULT_SESSION_ID,
        gql,
        TestTool,
    )
    await session.run()

    assert value_after_update == "<div>updated html!</div>"


@pytest.mark.asyncio
async def test_number_element_update_updates_value_in_instance() -> None:
    expected_number_value = 123.0
    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[NUMBER_ELEMENT],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )

    async def updates_mock(
        _session_id: str,
        _client_id: str,
    ) -> AsyncIterator[Updates]:
        yield Updates(
            toolSessionEvent=UpdatesToolSessionEventToolSessionElementUpdated(
                __typename="ToolSessionElementUpdated",
                element=UpdatesToolSessionEventToolSessionElementUpdatedElementNumberField(
                    __typename="NumberField",
                    id=NUMBER_ELEMENT_ID,
                    name=NUMBER_ELEMENT_NAME,
                    numberValue=expected_number_value,
                    graphContext=UpdatesToolSessionEventToolSessionElementUpdatedElementNumberFieldGraphContext(
                        parent=None,
                        affectedBy=[],
                        affects=[],
                    ),
                ),
            ),
        )

    gql.updates = updates_mock
    value_after_update = None

    @app
    class TestTool:
        number_param: float

        def number_param_updated(self) -> None:
            nonlocal value_after_update
            value_after_update = self.number_param

    session = await Session.initialize(
        DEFAULT_SESSION_ID,
        gql,
        TestTool,
    )
    await session.run()

    assert value_after_update == expected_number_value


@pytest.mark.asyncio
async def test_slider_element_update_updates_value_in_instance() -> None:
    expected_slider_value = 123.0
    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[SLIDER_ELEMENT],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )

    async def updates_mock(
        _session_id: str,
        _client_id: str,
    ) -> AsyncIterator[Updates]:
        yield Updates(
            toolSessionEvent=UpdatesToolSessionEventToolSessionElementUpdated(
                __typename="ToolSessionElementUpdated",
                element=UpdatesToolSessionEventToolSessionElementUpdatedElementSliderElement(
                    __typename="SliderElement",
                    id=SLIDER_ELEMENT_ID,
                    name=SLIDER_ELEMENT_NAME,
                    graphContext=UpdatesToolSessionEventToolSessionElementUpdatedElementSliderElementGraphContext(
                        parent=None,
                        affectedBy=[],
                        affects=[],
                    ),
                    sliderValue=expected_slider_value,
                    minValue=SLIDER_ELEMENT_MIN_VALUE,
                    maxValue=SLIDER_ELEMENT_MAX_VALUE,
                ),
            ),
        )

    gql.updates = updates_mock
    value_after_update = None

    @app
    class TestTool:
        slider_param: float = slider(
            min_value=SLIDER_ELEMENT_MIN_VALUE,
            max_value=SLIDER_ELEMENT_MAX_VALUE,
        )

        def slider_param_updated(self) -> None:
            nonlocal value_after_update
            value_after_update = self.slider_param

    session = await Session.initialize(
        DEFAULT_SESSION_ID,
        gql,
        TestTool,
    )
    await session.run()

    assert value_after_update == expected_slider_value


@pytest.mark.asyncio
async def test_container_child_element_update_updates_value_in_instance() -> None:
    expected_number_value = 123.0
    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[CONTAINER_ELEMENT, CONTAINER_CHILD_ELEMENT],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )

    async def updates_mock(
        _session_id: str,
        _client_id: str,
    ) -> AsyncIterator[Updates]:
        yield Updates(
            toolSessionEvent=UpdatesToolSessionEventToolSessionElementUpdated(
                __typename="ToolSessionElementUpdated",
                element=UpdatesToolSessionEventToolSessionElementUpdatedElementNumberField(
                    __typename="NumberField",
                    id=CONTAINER_CHILD_ELEMENT_ID,
                    name=CONTAINER_CHILD_ELEMENT_NAME,
                    numberValue=expected_number_value,
                    graphContext=UpdatesToolSessionEventToolSessionElementUpdatedElementNumberFieldGraphContext(
                        parent=GraphContextParent(
                            __typename="Container",
                            id=CONTAINER_ELEMENT_ID,
                        ),
                        affectedBy=[],
                        affects=[],
                    ),
                ),
            ),
        )

    gql.updates = updates_mock
    value_after_update = None

    @container
    class TestContainer:
        container_child: float

        def container_child_updated(self) -> None:
            nonlocal value_after_update
            value_after_update = self.container_child

    @app
    class TestTool:
        container: TestContainer

    session = await Session.initialize(
        DEFAULT_SESSION_ID,
        gql,
        TestTool,
    )
    await session.run()

    assert value_after_update == expected_number_value


@pytest.mark.asyncio
async def test_element_updated_event_triggers_update_handler() -> None:
    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[TEXT_ELEMENT],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )

    async def updates_mock(_session_id: str, _client_id: str) -> AsyncIterator[Updates]:
        yield Updates(
            toolSessionEvent=UpdatesToolSessionEventToolSessionElementUpdated(
                __typename="ToolSessionElementUpdated",
                element=UpdatesToolSessionEventToolSessionElementUpdatedElementTextField(
                    __typename="TextField",
                    id=TEXT_ELEMENT_ID,
                    name=TEXT_ELEMENT_NAME,
                    textValue="new text value",
                    graphContext=UpdatesToolSessionEventToolSessionElementUpdatedElementTextFieldGraphContext(
                        parent=None,
                        affectedBy=[],
                        affects=[],
                    ),
                ),
            ),
        )

    gql.updates = updates_mock

    param_updated = False

    @app
    class TestTool:
        text_param: str = "default value"

        def text_param_updated(self) -> None:
            nonlocal param_updated
            param_updated = True

    session = await Session.initialize(
        DEFAULT_SESSION_ID,
        gql,
        TestTool,
    )
    await session.run()

    assert param_updated


@pytest.mark.asyncio
async def test_action_triggered_event_calls_action_method() -> None:
    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[ACTION_ELEMENT],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )
    gql.updates = updates_mock_with_action_trigger

    action_called = False

    @app
    class TestTool:
        @action
        def my_action(self) -> None:
            nonlocal action_called
            action_called = True

    session = await Session.initialize(
        DEFAULT_SESSION_ID,
        gql,
        TestTool,
    )
    await session.run()

    assert action_called


@pytest.mark.asyncio
async def test_setting_parameter_value_calls_element_update_mutation() -> None:
    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[TEXT_ELEMENT, ACTION_ELEMENT],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )
    gql.update_element.return_value = UpdateElement(
        elementUpdate=UpdateElementElementUpdate(__typename="TextField"),
    )
    gql.updates = updates_mock_with_action_trigger

    @app
    class UpdateParameterTool:
        text_param: str = "text value"

        def my_action(self) -> None:
            self.text_param = "new text value"

    session = await Session.initialize(
        DEFAULT_SESSION_ID,
        gql,
        UpdateParameterTool,
    )
    await session.run()

    gql.update_element.assert_called_once_with(
        session_id=DEFAULT_SESSION_ID,
        client_id=DEFAULT_CLIENT_ID,
        element=ElementInput(elementID=TEXT_ELEMENT_ID, textValue="new text value"),
    )


@pytest.mark.asyncio
async def test_setting_html_element_value_calls_element_update_mutation() -> None:
    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[HTML_ELEMENT, ACTION_ELEMENT],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )
    gql.update_element.return_value = UpdateElement(
        elementUpdate=UpdateElementElementUpdate(__typename="HTMLElement"),
    )
    gql.updates = updates_mock_with_action_trigger

    @app
    class UpdateParameterTool:
        html_param: str = html()

        @action
        def my_action(self) -> None:
            self.html_param = "<div>new html content!</div>"

    session = await Session.initialize(
        DEFAULT_SESSION_ID,
        gql,
        UpdateParameterTool,
    )
    await session.run()

    gql.update_element.assert_called_once_with(
        session_id=DEFAULT_SESSION_ID,
        client_id=DEFAULT_CLIENT_ID,
        element=ElementInput(
            elementID=HTML_ELEMENT_ID,
            htmlValue="<div>new html content!</div>",
        ),
    )


@pytest.mark.asyncio
async def test_setting_slider_element_value_calls_element_update_mutation() -> None:
    expected_slider_update_value = 111.0

    gql = Mock(Client)
    gql.all_elements.return_value = AllElements(
        session=AllElementsSession(
            id=DEFAULT_SESSION_ID,
            all=[SLIDER_ELEMENT, ACTION_ELEMENT],
            clientID=DEFAULT_CLIENT_ID,
        ),
    )
    gql.update_element.return_value = UpdateElement(
        elementUpdate=UpdateElementElementUpdate(__typename="SliderElement"),
    )
    gql.updates = updates_mock_with_action_trigger

    @app
    class UpdateParameterTool:
        slider_param: float = slider(
            min_value=SLIDER_ELEMENT_MIN_VALUE,
            max_value=SLIDER_ELEMENT_MAX_VALUE,
        )

        @action
        def my_action(self) -> None:
            self.slider_param = expected_slider_update_value

    session = await Session.initialize(
        DEFAULT_SESSION_ID,
        gql,
        UpdateParameterTool,
    )
    await session.run()

    gql.update_element.assert_called_once_with(
        session_id=DEFAULT_SESSION_ID,
        client_id=DEFAULT_CLIENT_ID,
        element=ElementInput(
            elementID=SLIDER_ELEMENT_ID,
            sliderValue=expected_slider_update_value,
        ),
    )
