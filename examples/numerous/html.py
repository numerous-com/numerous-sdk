from numerous import app, html

HTML_TEMPLATE = """
<div>
    <strong>{title}: A HTML story</strong>
    <p>
        There once was a mark-up language with which you could create anything you liked.
    </p>
    <p>{story}</p>
    <p>The end!</p>
</div>
"""
DEFAULT_TITLE = "Numerous and the markup"
DEFAULT_STORY = "And then, one day you found numerous to empower it even further!"


@app
class HTMLApp:
    title: str = DEFAULT_TITLE
    story: str = DEFAULT_STORY
    html_example: str = html(
        default=HTML_TEMPLATE.format(title=DEFAULT_TITLE, story=DEFAULT_STORY),
    )

    def set_html(self) -> None:
        self.html_example = HTML_TEMPLATE.format(title=self.title, story=self.story)

    def title_updated(self) -> None:
        self.set_html()

    def story_updated(self) -> None:
        self.set_html()
