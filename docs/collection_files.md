# Collection files

Files can be stored in collections. They can contain arbitrary data.

## Referring to files

In order to access a document, use the `document` method of a collection.

Calling the `file` method does not actually create a file, but rather
creates a reference to that file.

No remote state of a document is actually accessed before methods are called on
the document object.

```py
col_ref = numerous.collections.collection("my-collection")
file_ref = col_ref.file("my-file.txt")
```

## Saving files

Files can be saved with either of the methods `save` (for saving `str` or
`bytes` data directly) or `save_file` (for saving a an opened IO object, e.g. a
file from the file system).

```py
col_ref = numerous.collections.collection("my-collection")
file_ref = col_ref.file("my-file.txt")

file_ref.save("my string data")
file_ref.save(b"my bytes data")
with open("some-local-file.txt", "r") as local_file:
    file_ref.save_file(local_file)
```

## Reading and opening files

File content can be read with `read_text`, `read_bytes`, or a file-like object
can be opened with `open`.

```py
col_ref = numerous.collections.collection("my-collection")
file_ref = col_ref.file("my-file.txt")

my_text = file_ref.read_text()

my_bytes = file_ref.read_bytes()

with file_ref.open() as f:
    f.read()
```

## Deleting files

Files can be deleted with the `delete` method.

```py
col_ref = numerous.collections.collection("my-collection")
file_ref = col_ref.file("my-file.txt")

file_ref.delete()
```
