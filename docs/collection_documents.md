# Collection documents

Documents with JSON content can be stored in collections.

## Referring to documents

Use the `document` method of a collection to access a document.

Calling the `document` method does not actually create a document, but rather
creates a reference to that document.

No remote state of a document is actually accessed before methods are called on
the document object.

```py
col_ref = numerous.collections.collection("my-collection")
doc_ref = col_ref.document("my-document")
```

## Listing and filtering documents

You can list all documents in a collection and also filter by specific tags.

```py
col_ref = numerous.collections.collection("my-collection")

# Iterate over all documents in the collection
for doc_ref in col_ref.documents():
    print(doc_ref.key, doc_ref.get())

# Iterate over all documents in the collection with the given tag
for doc_ref in col_ref.documents(tag_key="my-tag-key", tag_value="my-tag-value"):
    print(doc_ref.key, doc_ref.get())
```

## Loading document data

Use the `get` command to load the data in a document. This command returns any data stored
in the document or `None` if the document does not exist.

```py
col_ref = numerous.collections.collection("my-collection")
doc_ref = col_ref.document("my-document")
data = doc_ref.get()

if data is None:
    print("this document has not been set yet")
```

## Setting document content

To save data, you must `set` the document&apos;s content. This will override
any existing data in that document. If you wish to modify the data, load it with
`get` first, modify the loaded data, and `set` it again.

```py
col_ref = numerous.collections.collection("my-collection")
doc_ref = col_ref.document("my-document")
doc_ref.set({"field1": "my field 1 value", "field2": 2})
```

## Deleting documents

Documents can be deleted with the `delete` method.

```py
col_ref = numerous.collections.collection("my-collection")
doc = col_ref.document("my-document.txt")

doc.delete()
```

## Tagging documents

Documents can be tagged. Tags are used to filter documents, and to store
metadata about the documents.

```py
col_ref = numerous.collections.collection("my-collection")
doc_ref = col_ref.document("my-document")

doc_ref.tag("tag-key", "tag-value")

for key, value in doc_ref.tags.items():
    print(key, value)
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
doc_ref = col_ref.document("my-document")
doc_ref.set({
    "my-data": asdict(my_data),
    "my-datetime": my_datetime.isoformat(),
})

data = col_ref.document("my-document").get()
my_deserialized_datetime = datetime.fromisoformat(data["my-datetime"])
my_deserialized_data = MyData(**data["my-data"])
```
