"""Define applications with a dataclass-like interface."""

from dataclasses import dataclass
from types import MappingProxyType
from typing import Any, Callable, Optional, Type, Union, overload

from typing_extensions import dataclass_transform

from numerous.utils import MISSING, AppT


class HTML:
    def __init__(self, default: str) -> None:
        self.default = default


def html(  # type: ignore[no-untyped-def] # noqa: ANN201, PLR0913
    *,
    default: str = MISSING,  # type: ignore[assignment]
    default_factory: Callable[[], str] = MISSING,  # type: ignore[assignment] # noqa: ARG001
    init: bool = True,  # noqa: ARG001
    repr: bool = True,  # noqa: ARG001, A002
    hash: Optional[bool] = None,  # noqa: ARG001, A002
    compare: bool = True,  # noqa: ARG001
    metadata: Optional[MappingProxyType[str, Any]] = None,  # noqa: ARG001
    kw_only: bool = MISSING,  # type: ignore[assignment] # noqa: ARG001
):
    return HTML(default)


DEFAULT_FLOAT_MIN = 0.0
DEFAULT_FLOAT_MAX = 100.0


class Slider:
    def __init__(  # noqa: PLR0913
        self,
        *,
        default: float = MISSING,  # type: ignore[assignment]
        default_factory: Callable[[], float] = MISSING,  # type: ignore[assignment] # noqa: ARG002
        init: bool = True,  # noqa: ARG002
        repr: bool = True,  # noqa: ARG002, A002
        hash: Optional[bool] = None,  # noqa: ARG002, A002
        compare: bool = True,  # noqa: ARG002
        metadata: Optional[MappingProxyType[str, Any]] = None,  # noqa: ARG002
        kw_only: bool = MISSING,  # type: ignore[assignment] # noqa: ARG002
        label: Optional[str] = None,
        min_value: float = DEFAULT_FLOAT_MIN,
        max_value: float = DEFAULT_FLOAT_MAX,
    ) -> None:
        self.default = default
        self.label = label
        self.min_value = min_value
        self.max_value = max_value


def slider(  # type: ignore[no-untyped-def] # noqa: ANN201, PLR0913
    *,
    default: float = MISSING,  # type: ignore[assignment]
    default_factory: Callable[[], float] = MISSING,  # type: ignore[assignment] # noqa: ARG001
    init: bool = True,  # noqa: ARG001
    repr: bool = True,  # noqa: ARG001, A002
    hash: Optional[bool] = None,  # noqa: ARG001, A002
    compare: bool = True,  # noqa: ARG001
    metadata: Optional[MappingProxyType[str, Any]] = None,  # noqa: ARG001
    kw_only: bool = MISSING,  # type: ignore[assignment] # noqa: ARG001
    label: Optional[str] = None,
    min_value: float = DEFAULT_FLOAT_MIN,
    max_value: float = DEFAULT_FLOAT_MAX,
):
    return Slider(
        default=default,
        label=label,
        min_value=min_value,
        max_value=max_value,
    )


class Field:
    def __init__(  # noqa: PLR0913
        self,
        *,
        default: Union[str, float] = MISSING,  # type: ignore[assignment]
        default_factory: Callable[[], Union[str, float]] = MISSING,  # type: ignore[assignment]
        init: bool = True,  # noqa: ARG002
        repr: bool = True,  # noqa: ARG002, A002
        hash: Optional[bool] = None,  # noqa: ARG002, A002
        compare: bool = True,  # noqa: ARG002
        metadata: Optional[MappingProxyType[str, Any]] = None,  # noqa: ARG002
        kw_only: bool = MISSING,  # type: ignore[assignment] # noqa: ARG002
        label: Optional[str] = None,
    ) -> None:
        if default is MISSING and default_factory is not MISSING:  # type: ignore[comparison-overlap]
            default = default_factory()
        self.default = default
        self.label = label


def field(  # type: ignore[no-untyped-def] # noqa: ANN201, PLR0913
    *,
    default: Union[str, float] = MISSING,  # type: ignore[assignment]
    default_factory: Callable[[], float] = MISSING,  # type: ignore[assignment]
    init: bool = True,  # noqa: ARG001
    repr: bool = True,  # noqa: ARG001, A002
    hash: Optional[bool] = None,  # noqa: ARG001, A002
    compare: bool = True,  # noqa: ARG001
    metadata: Optional[MappingProxyType[str, Any]] = None,  # noqa: ARG001
    kw_only: bool = MISSING,  # type: ignore[assignment] # noqa: ARG001
    label: Optional[str] = None,
):
    return Field(default=default, default_factory=default_factory, label=label)


@dataclass_transform()
def container(cls: Type[AppT]) -> Type[AppT]:
    """Define a container."""
    cls.__container__ = True  # type: ignore[attr-defined]
    return dataclass(cls)


def action(action: Callable[[AppT], Any]) -> Callable[[AppT], Any]:
    """Define an action."""
    action.__action__ = True  # type: ignore[attr-defined]
    return action


@overload
def app(cls: Type[AppT]) -> Type[AppT]: ...


@overload
def app(title: str = ...) -> Callable[[Type[AppT]], Type[AppT]]: ...


def app(
    *args: Any,
    **kwargs: Any,
) -> Union[Type[AppT], Callable[[Type[AppT]], Type[AppT]]]:
    invalid_error_message = "Invalid @app usage"
    if len(args) == 1 and not kwargs:
        return app_decorator()(args[0])
    if len(args) == 0 and "title" in kwargs:
        return app_decorator(**kwargs)
    raise ValueError(invalid_error_message)


def app_decorator(**kwargs: dict[str, Any]) -> Callable[[Type[AppT]], Type[AppT]]:
    @dataclass_transform(field_specifiers=(field, html, slider))
    def decorator(cls: Type[AppT]) -> Type[AppT]:
        cls.__numerous_app__ = True  # type: ignore[attr-defined]
        if title := kwargs.get("title"):
            cls.__title__ = title  # type: ignore[attr-defined]
        return dataclass(cls)

    return decorator
