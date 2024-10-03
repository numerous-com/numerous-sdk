import marimo

__generated_with = "0.8.14"
app = marimo.App(width="medium")


@app.cell
def __():
    from numerous.experimental.marimo import cookies

    cookies = cookies()
    cookies
    return cookies,


if __name__ == "__main__":
    app.run()
