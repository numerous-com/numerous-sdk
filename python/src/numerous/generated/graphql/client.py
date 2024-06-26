# Generated by ariadne-codegen
# Source: queries.gql

from typing import Any, AsyncIterator, Dict

from .all_elements import AllElements
from .async_base_client import AsyncBaseClient
from .input_types import ElementInput
from .update_element import UpdateElement
from .updates import Updates


def gql(q: str) -> str:
    return q


class Client(AsyncBaseClient):
    async def all_elements(self, tool_session_id: str, **kwargs: Any) -> AllElements:
        query = gql(
            """
            query AllElements($toolSessionId: ID!) {
              session: toolSession(id: $toolSessionId) {
                id
                clientID
                all: allElements {
                  __typename
                  id
                  name
                  graphContext {
                    ...GraphContext
                  }
                  ...ButtonValue
                  ...TextFieldValue
                  ...NumberFieldValue
                  ...HTMLValue
                  ...SliderValue
                }
              }
            }

            fragment ButtonValue on Button {
              buttonValue: value
            }

            fragment GraphContext on ElementGraphContext {
              parent {
                __typename
                id
              }
              affectedBy {
                id
              }
              affects {
                id
              }
            }

            fragment HTMLValue on HTMLElement {
              html
            }

            fragment NumberFieldValue on NumberField {
              numberValue: value
            }

            fragment SliderValue on SliderElement {
              sliderValue: value
              minValue
              maxValue
            }

            fragment TextFieldValue on TextField {
              textValue: value
            }
            """
        )
        variables: Dict[str, object] = {"toolSessionId": tool_session_id}
        response = await self.execute(
            query=query, operation_name="AllElements", variables=variables, **kwargs
        )
        data = self.get_data(response)
        return AllElements.model_validate(data)

    async def update_element(
        self, session_id: str, client_id: str, element: ElementInput, **kwargs: Any
    ) -> UpdateElement:
        query = gql(
            """
            mutation UpdateElement($sessionID: ID!, $clientID: ID!, $element: ElementInput!) {
              elementUpdate(toolSessionId: $sessionID, clientId: $clientID, element: $element) {
                __typename
              }
            }
            """
        )
        variables: Dict[str, object] = {
            "sessionID": session_id,
            "clientID": client_id,
            "element": element,
        }
        response = await self.execute(
            query=query, operation_name="UpdateElement", variables=variables, **kwargs
        )
        data = self.get_data(response)
        return UpdateElement.model_validate(data)

    async def updates(
        self, session_id: str, client_id: str, **kwargs: Any
    ) -> AsyncIterator[Updates]:
        query = gql(
            """
            subscription Updates($sessionId: ID!, $clientId: ID!) {
              toolSessionEvent(toolSessionId: $sessionId, clientId: $clientId) {
                __typename
                ... on ToolSessionElementUpdated {
                  element {
                    __typename
                    id
                    name
                    graphContext {
                      ...GraphContext
                    }
                    ...ButtonValue
                    ...TextFieldValue
                    ...NumberFieldValue
                    ...HTMLValue
                    ...SliderValue
                  }
                }
                ... on ToolSessionActionTriggered {
                  element {
                    id
                    name
                  }
                }
              }
            }

            fragment ButtonValue on Button {
              buttonValue: value
            }

            fragment GraphContext on ElementGraphContext {
              parent {
                __typename
                id
              }
              affectedBy {
                id
              }
              affects {
                id
              }
            }

            fragment HTMLValue on HTMLElement {
              html
            }

            fragment NumberFieldValue on NumberField {
              numberValue: value
            }

            fragment SliderValue on SliderElement {
              sliderValue: value
              minValue
              maxValue
            }

            fragment TextFieldValue on TextField {
              textValue: value
            }
            """
        )
        variables: Dict[str, object] = {"sessionId": session_id, "clientId": client_id}
        async for data in self.execute_ws(
            query=query, operation_name="Updates", variables=variables, **kwargs
        ):
            yield Updates.model_validate(data)
