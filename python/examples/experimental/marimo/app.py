# mypy: ignore-errors

import marimo

__generated_with = "0.3.12"
app = marimo.App()


@app.cell
def __():
    import marimo as mo

    return (mo,)


@app.cell
def __():
    from numerous.experimental.marimo import Field
    from numerous.experimental.model import BaseModel

    return BaseModel, Field


@app.cell
def __(BaseModel, Field):
    """
    Create your own model by extending the base Model class. This is like defining a pydantic base model. You can use the same arguments for the MoField (short for Marimo Field) as you would normally for the pydantic Field. The exception is that you either need to provide a type as the annotation keyword argument or the type can be inferred from the default value specified as the first argument or the keyword "default".

    The reason for having a dedicated MoField and Model classes and not using pydantic directly is to allow for code completion to work.

    The Model class could be changed to a decorator, but this would lead to a lot of monky-patching and also not give code completion for Numerous related functionality later on.
    """

    class SubModel(BaseModel):
        c = Field("c")

    return (SubModel,)


@app.cell
def __(BaseModel, Field, SubModel):
    """
    You can add models inside other models to create composite models easily.
    """

    class MyModel(BaseModel):
        a = Field(1, ge=0, le=5)
        b = Field("bla bla", min_length=0, max_length=100)
        sub_model = SubModel()

    return (MyModel,)


@app.cell
def __(MyModel):
    """Instanciate your model as with normal pydantic."""
    my_model = MyModel()
    return (my_model,)


@app.cell
def __(my_model):
    """
    Each field is part from being a way to define a validation scheme also a Marimo state object. You can use this to make ui controls that respond to changes in the state.
    For marimo to detect this state you need to assign it to a global variable.
    """
    state_a = my_model.a
    return (state_a,)


@app.cell
def __(my_model, state_a):
    """Create a slider from the MoField. If you reference the state defined above, marimo will update the slider value based on other UI components making changes to it."""
    state_a
    a_slider = my_model.a.slider()
    a_slider
    return (a_slider,)


@app.cell
def __(my_model, state_a):
    """We can also make a number field easily. Try to change it to see the slider also changing above."""
    state_a
    a_number = my_model.a.number()
    a_number
    return (a_number,)


@app.cell
def __(mo, my_model):
    """
    Now we are prepared to also sync state with the Numerous platform by since the model is notified of any changes to its data and any changes on the server side can be synchronized with the UI.

    Clicking the button below mimics how Numerous could update your UI with new data from the platform.
    """

    def on_click(event):
        my_model.a.value = 0

    button = mo.ui.button(label="Update a", on_click=on_click)
    button
    return button, on_click


@app.cell
def __(my_model):
    """Finally you can access a pydantic model directly if needed. Note updates to this is not reverse synchronized to my_model"""
    my_model.pydantic_model
    return


if __name__ == "__main__":
    app.run()
