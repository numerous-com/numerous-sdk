from numerous import app, slider


@app
class SliderApp:
    result: float
    default_slider: float = slider()
    custom_slider: float = slider(
        default=10.0,
        label="My custom label",
        min_value=-20.0,
        max_value=20.0,
    )

    def _compute(self):
        self.result = self.default_slider * self.custom_slider

    def default_slider_updated(self):
        self._compute()
        print("default slider updated", self.default_slider, self.result)

    def custom_slider_updated(self):
        self._compute()
        print("custom slider updated", self.custom_slider, self.result)
