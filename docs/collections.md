# Collections

With Numerous Collections, you can persist (or save) data for your application.
Collections acts as a schemaless database, where users can store, retrieve, and search data in a simple and flexible way.

1. Organize data in collections, indexed with user-specified keys.
2. Store documents with JSON data (optionally indexed with user-specified keys).
3. Store files (optionally indexed with user-specified keys).
4. Tag documents and files with key/value tags, and filter documents and files
   by these tags.


> **Note:**
> This feature only supports apps that are deployed to Numerous.


> **Tip:**
> Remember to add `numerous` as a dependency in your project; most likely to your `requirements.txt` file.


Import the [Numerous SDK](http://www.pypi.org/project/numerous) in your Python
code.

Now, you can add code to your app that is similar to the following:

```python
from numerous.collections import collection

# Refer to a collection by its key
col_ref = collection("my-collection")

# Refer to a collection nested inside another collection
nested = col_ref.collection("nested-collection")

# Refer to a document by its key
doc_ref = col_ref.document("my-document")

# Read a document's data
data = doc_ref.get()

# Update a document with new data
data["my-value"] = "new value"
doc_ref.set(data)

# Loop over documents in a collection
for doc in col_ref.documents():
    print(doc.key)
    print(doc.get())

# Delete a document
doc_ref.delete()

# Check if a document exists
if doc_ref.exists:
    print("document exists")
else:
    print("document does not exist")
```

## Using Collections

Numerous Collections acts as a schemaless document database and a folder of files.

### Creating and accessing collections

In order to access a collection in your application, you can use
`numerous.collections.collection`, which returns a `CollectionReference` with
the specified key.

Collections are automatically created the first time they are accessed.

```python
collection = numerous.collections.collection("my-collection")
```

<Callout type="info">
  Collection keys are scoped to an organization, meaning that if multiple apps
  use the same Collection keys, they will access the same data.

  For nested collections, they are scoped to their parent collection.
</Callout>

### Creating and accessing nested collections

Collections can contain other collections. You can use the `collection` method
on a collection to get a nested Collection with the specified key.

Nested collections are automatically created the first time they are accessed.

```python
parent = numerous.collections.collection("parent-collection")
nested = parent.collection("nested-collection")
```

### Accessing documents inside collections

Documents can be accessed by their key or can otherwise be iterated over.

In order to access a specific document, use the `document` method:

```python
col_ref = numerous.collections.collection("my-collection")
doc_ref = col_ref.document("my-document")
```

To iterate over the documents in a collection, use the `documents`
method. You can filter documents by tags by providing `tag_key` and `tag_value`
keyword arguments.

```python
col_ref = numerous.collections.collection("my-collection")

for document in col_ref.documents():
    print(document.get())

for document in col_ref.documents(tag_key="my-tag-key", tag_value="my-tag-value"):
    print(document.get())
```

### Collections in subscription apps

Collections exist within the organization that deploys the app.

It is on our immediate roadmap to store data in the subscribing
organization instead, in order to ensure correct ownership of data produced by
apps.

## Using Documents

Documents contain JSON serializable data.

### Referring to documents

In order to access a document, use the `document` method of a collection.

Calling the `document` method does not actually create a document, but rather
creates a reference to that document.

No remote state of a document is actually accessed before methods are called on
the document object.

```python
col_ref = numerous.collections.collection("my-collection")
doc_ref = collection.document("my-document")
```

### Loading document data

In order to load the data in a document, use the `get`command. This command returns any data stored
in the document, or `None` if the document does not exist.

```python
col_ref = numerous.collections.collection("my-collection")
doc_ref = collection.document("my-document")
data = doc_ref.get()

if data is None:
    print("this document has not been set yet")
```

### Setting document content

In order to save data, you must `set` the document's content. This will override
any existing data in that document. If you wish to modify the data, load it with
`get` first, modify the loaded data, and `set` it again.

```python
col_ref = numerous.collections.collection("my-collection")
doc_ref = collection.document("my-document")
doc_ref.set({"field1": "my field 1 value", "field2": 2})
```

### Serializing and deserializing custom types

The Numerous SDK serializes the data used to set the document contents with the
built-in `json.dumps`, which is limited in terms of what it can serialize. For
example, `datetime.datetime` is not serializable.

Currently, it is up to the application developer to handle this serialization
manually.

In the example below, we manually serialize and deserialize `MyData` and
`datetime.datetime` objects.

```python
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

## Using files

Files contain arbitrary data.

### Referring to files

In order to access a document, use the `document` method of a collection.

Calling the `file` method does not actually create a file, but rather
creates a reference to that file.

No remote state of a document is actually accessed before methods are called on
the document object.

```python
col_ref = numerous.collections.collection("my-collection")
file_ref = collection.file("my-file.txt")
```

### Saving files

Files can be saved with either of the methods `save` (for saving `str` or
`bytes` data directly) or `save_file` (for saving a an opened IO object, e.g. a
file from the file system).

```python
col_ref = numerous.collections.collection("my-collection")
file_ref = collection.file("my-file.txt")

file_ref.save("my string data")
file_ref.save(b"my bytes data")
with open("some-local-file.txt", "r") as local_file:
    file_ref.save_file(local_file)
```

### Reading and opening files

File content can be read with `read_text`, `read_bytes`, or a file-like object
can be opened with `open`.

```python
col_ref = numerous.collections.collection("my-collection")
file_ref = collection.file("my-file.txt")

my_text = file_ref.read_text()

my_bytes = file_ref.read_bytes()

with file_ref.open() as f:
    f.read()
```

### Deleting files

Files can be deleted with the `delete` method.

```python
col_ref = numerous.collections.collection("my-collection")
file_ref = collection.file("my-file.txt")

file_ref.delete()
```


## API Documentation

## ::: numerous.collections
    options:
        show_root_heading: true
