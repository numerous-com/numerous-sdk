from numerous import action, app


@app
class ActionApp:
    count: float
    message: str

    @action
    def increment(self) -> None:
        self.count += 1
        self.message = f"Count now: {self.count}"
        print("Incrementing count:", self.count)
