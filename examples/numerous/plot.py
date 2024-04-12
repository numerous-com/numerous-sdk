import math

from numerous import app, field, slider
from plotly import graph_objects as go


@app
class PlotlyExample:
    phase_shift: float = slider(default=1, label="Phase Shift")
    vertical_shift: float = slider(default=1, label="Vertical Shift")
    period: float = slider(default=1, min_value=1, max_value=100, label="Period")
    amplitude: float = slider(default=1, label="Amplitude")
    graph: go.Figure = field()

    def phase_shift_updated(self):
        print("Updated phase shift", self.phase_shift)
        self._graph()

    def vertical_shift_updated(self):
        print("Updated vertical shift", self.vertical_shift)
        self._graph()

    def period_updated(self):
        print("Updated period", self.period)
        self._graph()

    def amplitude_updated(self):
        print("Updated amplitude", self.amplitude)
        self._graph()

    def _graph(self):
        xs = [i / 10.0 for i in range(1000)]
        ys = [self._sin(x) for x in xs]
        self.graph = go.Figure(go.Scatter(x=xs, y=ys))

    def _sin(self, t: float):
        b = 2.0 * math.pi / self.period
        return (
            self.amplitude * math.sin(b * (t + self.phase_shift)) + self.vertical_shift
        )
