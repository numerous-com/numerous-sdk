"""
The model module.

The module defines a `BaseModel` class that dynamically creates
a Pydantic model based on the declared fields and interfaces
in the subclass.

It leverages Pydantic's `create_model` function to handle validation and
type enforcement.

The `BaseModel` class provides methods to initialize the model,
retrieve the validated data,
and access the dynamically created Pydantic model object.
Additionally, the module defines a `Field` class that represents a field
in the model.

The `Field` class handles default values,
type annotations, and other properties of the field.

Classes:
- BaseModel: Class representing a generic model that integrates with
    Pydantic for data validation.
- Field: Class representing a field in the model.
"""

from typing import Any, Dict, Tuple, Union

from pydantic import BaseModel as PydanticBaseModel
from pydantic import Field as PydanticField
from pydantic import create_model


class _ModelInterface:
    """Interface for a model object."""

    _name: str

    @property
    def model_attrs(self) -> Tuple[Any, PydanticBaseModel]:
        """
        Return the model attributes.

        Returns
        -------
        tuple
            A tuple containing the model attributes.

        """
        raise NotImplementedError

    @property
    def value(self) -> PydanticBaseModel:
        """
        Return the model value.

        Returns
        -------
        PydanticBaseModel
            The model value.

        """
        raise NotImplementedError


class BaseModel(_ModelInterface):
    """
    Class representing a generic model that use with Pydantic for data validation.

    This class constructs a Pydantic model dynamically based on the declared fields and
    interfaces in the class. It leverages Pydantic's `create_model` function to handle
    validation and type enforcement.

    Attributes
    ----------
    pydantic_model_cls : PydanticBaseModel
        Dynamically created Pydantic model class based on the fields
        defined in the subclass.
    _fields : dict
        A dictionary mapping field names to Field or _ModelInterface instances,
        representing the attributes of the model.

    Methods
    -------
    __init__(**kwargs)
        Initializes the model by creating a Pydantic model and validating
        the input fields.
    value()
        Returns the Pydantic model instance containing validated data.
    model_attrs()
        Returns attributes of the dynamically created Pydantic model class.
    pydantic_model()
        Returns an instance of the dynamically created Pydantic model
        containing validated data.

    """

    def __init__(self, **kwargs: Dict[str, Any]) -> None:
        """
        Initialize a model object with the given fields.

        Args:
        ----
            **kwargs (dict): Keyword arguments representing the field values.

        """
        _attrs = {}
        for key, val in self.__class__.__dict__.items():
            # Find all the fields in the class
            if isinstance(val, Field):
                _attrs[key] = val.field_attrs
            elif isinstance(val, _ModelInterface):
                # If the field is a subclass of Model, get the model attributes
                _attrs[key] = val.model_attrs

        # Create a Pydantic model class with the fields
        self.pydantic_model_cls = create_model(self.__class__.__name__, **_attrs)  # type: ignore[call-overload]

        # Set them as values to Field objects
        for key, val in kwargs.items():
            getattr(self, key)._field_value = val  # noqa: SLF001

        self._fields = {}
        for key, val in self.__class__.__dict__.items():
            # Find all the fields in the class
            if isinstance(val, (Field, _ModelInterface)):
                val._name = key  # noqa: SLF001
                self._fields[key] = val

        # Check if any non-optional fields are missing a value
        for val in self._fields.values():
            # By accessing the value property, the value is checked for validity
            _ = val.value
            # Add self as the parent model.
            # This is used to validate the model when a field is changed
            val._parent_model = self  # noqa: SLF001

        # Trigger a validation of the model by accessing the value property
        _ = self.pydantic_model

    @property
    def value(self) -> PydanticBaseModel:
        """
        Return the PydanticBaseModel instance associated with this object.

        Returns
        -------
        PydanticBaseModel
            The PydanticBaseModel instance.

        """
        return self.pydantic_model

    @property
    def model_attrs(self) -> Tuple[Any, PydanticBaseModel]:
        """
        Return the data required to create a Pydantic model object.

        Returns
        -------
        tuple
            A tuple containing the type of the Pydantic model
            and the model object itself.

        """
        return (type(self.pydantic_model), self.pydantic_model)

    @property
    def pydantic_model(self) -> PydanticBaseModel:
        """
        Get a Pydantic model object representing the model.

        Returns
        -------
        BaseModel
            The Pydantic model object with the values from the current model.

        """
        _kwargs = {}

        # Get the values from the fields
        for key, val in self._fields.items():
            _kwargs[key] = val.value

        # Create a Pydantic model object with the values
        pydantic_model = self.pydantic_model_cls(**_kwargs)
        if isinstance(pydantic_model, PydanticBaseModel):
            return pydantic_model

        error_msg = "Invalid Pydantic model object."
        raise TypeError(error_msg)


class Field:
    def __init__(
        self,
        default: Union[str, float, None] = None,
        annotation: Union[type, None] = None,
        **kwargs: Dict[str, Any],
    ) -> None:
        """
        Initialize a Field object.

        Parameters
        ----------
        default : str or float or None, optional
            The default value for the field, by default ...
        annotation : type or None, optional
            The type annotation for the field, by default None
        **kwargs : dict
            Additional properties for the field.

        """
        self._default = default
        self._props = kwargs
        self._default = default
        self._field_value = default
        self._name: Union[str, None] = None
        self._parent_model = None

        # Check if the annotation is provided if the default value is None
        if annotation is None and default is None:
            error_msg = "Annotation must be provided if value is None"
            raise ValueError(error_msg)

        # Set the annotation to the type of the default value if it is not provided
        self._annotation = annotation if annotation is not None else type(default)

    @property
    def name(self) -> str:
        """
        Returns the name of the field.

        Raises
        ------
        ValueError
            If the name has not been set.

        Returns
        -------
        str
            The name of the field.

        """
        if self._name is None:
            error_msg = "Name has not been set"
            raise ValueError(error_msg)
        return self._name

    @name.setter
    def name(self, name: str) -> None:
        """
        Set the name of the field.

        Parameters
        ----------
        name : str
            The name of the field.

        Raises
        ------
        ValueError
            If the name has already been set.

        """
        if self._name is not None:
            error_msg = "Name has already been set"
            raise ValueError(error_msg)
        self._name = name

    @property
    def field_attrs(self) -> Tuple[type, Any]:
        """
        Get the attributes of the field.

        Returns
        -------
        tuple
            A tuple containing the value and type annotation of the field,
            along with other properties.

        """
        return (self._annotation, self.field_info)

    @property
    def field_info(self) -> Any:  # noqa: ANN401
        """
        Get the field information.

        Returns
        -------
        tuple
            A tuple containing the Pydantic Field object with the value and properties.

        """
        # Create a Pydantic Field object with the value and properties
        self._field_info = PydanticField(self._default, **self._props)  # type: ignore[arg-type]
        return self._field_info

    @property
    def value(self) -> Union[str, float]:
        """
        Return the value of the object.

        Returns
        -------
        str or float
            The value of the object.

        """
        return self.get()

    @value.setter
    def value(self, value: Union[str, float]) -> None:
        """
        Set the value of the object.

        Parameters
        ----------
        value : str or float
            The new value to be set.

        """
        self.set(value)

    def get(self) -> Union[str, float]:
        """
        Get the value of the field.

        Returns
        -------
        str or float
            The value of the model.

        Raises
        ------
        ValueError
            If the value has not been set.

        """
        if self._field_value is None:
            error_msg = "Value has not been set"
            raise ValueError(error_msg)
        return self._field_value

    def set(self, value: Any) -> None:  # noqa: ANN401
        """
        Set the value of the field.

        Parameters
        ----------
        value : str or float
            The new value to be set.

        Raises
        ------
        Exception
            If the parent model validation fails.

        """
        old_value = self._field_value
        self._field_value = value
        if self._parent_model is None:
            error_msg = "Parent model has not been set"
            raise ValueError(error_msg)
        try:
            # Trigger the parent model validation
            _ = self._parent_model.value
        except:
            # If the validation fails, revert the value
            self._field_value = old_value
            # Raise the error
            raise
