# Persistence with Collections

With Numerous Collections, you can persist (or save) data for your application.
Collections acts as a schemaless database, where users can store, retrieve, and search data in a simple and flexible way.

1. Organize data in collections, indexed with user-specified keys.
2. Store documents with JSON data (optionally indexed with user-specified keys).
3. Store files (optionally indexed with user-specified keys).
4. Tag documents and files with key/value tags, and filter documents and files
   by these tags.

!!! info
    This feature only supports apps that are deployed to Numerous.

!!! info
    Remember to add `numerous` as a dependency in your project; most likely to
    your `requirements.txt` file.

## Basic usage

Import the [Numerous SDK](http://www.pypi.org/project/numerous) in your Python
code.

Now, you can add code to your app that is similar to the following:

```python
import numerous

# Refer to a collection by its key
collection = numerous.collection("my-collection")

# Refer to a collection nested inside another collection
nested = collection.collection("nested-collection")

# Refer to a document by its key
document = collection.document("my-document")

# Read a document's data
data = document.get()

# Update a document with new data
data["my-value"] = "new value"
document.set(data)

# Loop over documents in a collection
for doc in collection.documents():
    print(doc.key)
    print(doc.get())

# Delete a document
document.delete()

# Check if a document exists
if document.exists:
    print("document exists")
else:
    print("document does not exist")
```

::: numerous.collection.collection
::: numerous.collection.numerous_collection.NumerousCollection
::: numerous.collection.numerous_document.NumerousDocument
