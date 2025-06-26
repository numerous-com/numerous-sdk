# Collections

With Numerous Collections, you can persist (or save) data for your application.
Collections acts as a schemaless database, where users can store, retrieve, and search data in a simple and flexible way.

1. Organize data in collections, indexed with user-specified keys.
2. Store documents with JSON data (optionally indexed with user-specified keys).
3. Store files (optionally indexed with user-specified keys).
4. Tag documents, files, and collections with key/value tags, and filter documents, files, and collections
   by these tags.

!!! note
This feature only supports apps that are deployed to Numerous.

!!! tip
Remember to add `numerous` as a dependency in your project; most likely to your `requirements.txt` file.

Import the [Numerous SDK](http://www.pypi.org/project/numerous) in your Python
code.

Now, you can add code to your app that is similar to the following:

```py
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

# Loop over nested collections in a collection
for nested_col in col_ref.collections():
    print(nested_col.key)

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

To access a collection in your application, you can use
`numerous.collections.collection`, which returns a `CollectionReference` with
the specified key.

Collections are automatically created the first time they are accessed.

```py
collection = numerous.collections.collection("my-collection")
```

!!! note
Collection keys are scoped to an organization, meaning that if multiple apps
use the same collection keys, they will access the same data.

    For nested collections, they are scoped to their parent collection.

See the [reference](/reference/numerous/collections/collection#numerous.collections.collection.collection) for more information.

### Creating and accessing nested collections

Collections can contain other collections. You can use the `collection` method
on a collection to get a nested collection with the specified key.

Nested collections are automatically created the first time they are accessed.

```py
parent = numerous.collections.collection("parent-collection")
nested = parent.collection("nested-collection")
```

### Accessing documents and files in collections

Documents and files can be accessed by their key or can otherwise be iterated
over.

See more on the user guides for [collection documents](collection_documents.md)
and [collection files](collection_files.md):

```py
col_ref = numerous.collections.collection("my-collection")

doc_ref = col_ref.document("my-document")
for document in col_ref.documents():
    print(document.get())

file_ref = col_ref.file("my-file")
for file in col_ref.files():
    print(file.get())
```

### Filtering collections by tags

You can list all nested collections in a collection and also filter by specific tags.

```py
col_ref = numerous.collections.collection("my-collection")

# Iterate over all nested collections in the collection
for nested_col in col_ref.collections():
    print(nested_col.key)

# Iterate over all nested collections in the collection with the given tag
for nested_col in col_ref.collections(tag_key="my-tag-key", tag_value="my-tag-value"):
    print(nested_col.key)
```

### Tagging collections

Collections can be tagged. Tags are used to filter collections and to store
metadata about the collections.

```py
col_ref = numerous.collections.collection("my-collection")

col_ref.tag("tag-key", "tag-value")

for key, value in col_ref.tags.items():
    print(key, value)
```

### Collections in subscription apps

Collections exist within the organization that deploys the app.

It is on our immediate roadmap to store data in the subscribing
organization instead to ensure correct ownership of data produced by
apps.

## API reference

See the [API reference](reference/numerous/collections/index.md) for details.
