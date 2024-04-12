from numerous import app, container


def test_initialize_decorated_app() -> None:
    param_value = 5

    @app
    class App:
        param: int = param_value

    instance = App()

    assert instance
    assert instance.param == param_value
    assert getattr(instance, "__numerous_app__", False) is True


def test_initialize_decorated_container() -> None:
    param_value = 5

    @container
    class Container:
        param: int = param_value

    instance = Container()

    assert instance
    assert instance.param == param_value
    assert getattr(instance, "__container__", False) is True


def test_initialize_decorated_app_with_pretty_name() -> None:
    param_value = 5

    @app(title="my_pretty_test_app")
    class App:
        param: int = param_value

    instance = App()

    assert instance
    assert instance.param == param_value
    assert getattr(instance, "__title__", False) == "my_pretty_test_app"
    assert getattr(instance, "__numerous_app__", False) is True
