# Collection files

Files can be stored in collections. They can contain arbitrary data.

## Referring to files

Use the `file` method of a collection to access a file.

Calling the `file` method does not actually create a file, but rather
creates a reference to that file.

No remote state of a document is actually accessed before methods are called on
the document object.

```py
col_ref = numerous.collections.collection("my-collection")
file_ref = col_ref.file("my-file.txt")
```

## Listing and filtering files

You can list all files in a collection and also filter by specific tags.

```py
col_ref = numerous.collections.collection("my-collection")

# Iterate over all files in the collection
for file_ref in col_ref.files():
    print(file_ref.key, file_ref.get())

# Iterate over all files in the collection with the given tag
for file_ref in col_ref.files(tag_key="my-tag-key", tag_value="my-tag-value"):
    print(file_ref.key, file_ref.get())
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

## Saving files

Files can be saved with either of the two methods: `save` (for saving `str` or
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

## Deleting files

Files can be deleted with the `delete` method.

```py
col_ref = numerous.collections.collection("my-collection")
file_ref = col_ref.file("my-file.txt")

file_ref.delete()
```

## Tagging files

Files can be tagged. Tags are used to filter files and to store
metadata about the files.

```py
col_ref = numerous.collections.collection("my-collection")
file_ref = col_ref.file("my-file")

file_ref.tag("tag-key", "tag-value")

for key, value in file_ref.tags.items():
    print(key, value)
```
