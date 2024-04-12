from numerous import app


@app
class ParameterApp:
    param1: str
    param2: float
    output: str

    def param1_updated(self):
        print("Got an update:", self.param1)
        self.output = f"output: {self.param1}"
