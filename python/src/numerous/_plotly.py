import plotly.graph_objects as go


def plotly_html(figure: go.Figure) -> str:
    return figure.to_html(include_plotlyjs="cdn", full_html=False)  # type: ignore[no-any-return]
