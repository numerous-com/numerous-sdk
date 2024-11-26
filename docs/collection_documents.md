
# Collection documents

Documents with JSON content can be stored in collections.

## Referring to documents

In order to access a document, use the `document` method of a collection.

Calling the `document` method does not actually create a document, but rather
creates a reference to that document.

No remote state of a document is actually accessed before methods are called on
the document object.

```py
col_ref = numerous.collections.collection("my-collection")
doc_ref = collection.document("my-document")
```

## Loading document data

In order to load the data in a document, use the `get`command. This command returns any data stored
in the document, or `None` if the document does not exist.

```py
col_ref = numerous.collections.collection("my-collection")
doc_ref = collection.document("my-document")
data = doc_ref.get()

if data is None:
    print("this document has not been set yet")
```

## Setting document content

In order to save data, you must `set` the document's content. This will override
any existing data in that document. If you wish to modify the data, load it with
`get` first, modify the loaded data, and `set` it again.

```py
col_ref = numerous.collections.collection("my-collection")
doc_ref = collection.document("my-document")
doc_ref.set({"field1": "my field 1 value", "field2": 2})
```

## Serializing and deserializing custom types

The Numerous SDK serializes the data used to set the document contents with the
built-in `json.dumps`, which is limited in terms of what it can serialize. For
example, `datetime.datetime` is not serializable.

Currently, it is up to the application developer to handle this serialization
manually.

In the example below, we manually serialize and deserialize `MyData` and
`datetime.datetime` objects.

```py
from dataclasses import dataclass, asdict
from datetime import datetime

@dataclass
class MyData:
    field1: str
    field2: int


my_datetime = datetime.now()
my_data = MyData(field1="my field 1 value", field2=2)

col_ref = numerous.collections.collection("my-collection")
doc_ref = collection.document("my-document")
doc_ref.set({
    "my-data": asdict(my_data),
    "my-datetime": my_datetime.isoformat(),
})

data = col_ref.document("my-document").get()
my_deserialized_datetime = datetime.fromisoformat(data["my-datetime"])
my_deserialized_data = MyData(**data["my-data"])
```
