from numerous import action, app, container


@container
class MyNestedContainer:
    nested_child: float


@container
class MyContainer:
    child: str
    nested: MyNestedContainer

    def child_updated(self) -> None:
        print("wow")


@app
class MyContainerApp:
    my_container: MyContainer
    output: str

    @action
    def set_output(self) -> None:
        self.output = f"First child is {self.my_container.child}, deep child is {self.my_container.nested.nested_child}"
