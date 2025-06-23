# Bulk Upload/Download with Numerous Collections

The Numerous SDK provides utilities for performing bulk upload and download operations between a remote Numerous Collection and your local filesystem. These tools are designed to help you efficiently manage large sets of files and documents within your collections.

## Overview

Bulk operations allow you to:

-   **Download** an entire collection hierarchy (including all files, documents, and nested collections) from Numerous to your local machine.
-   **Upload** a local directory structure (containing files and documents) into a Numerous collection, automatically creating nested collections as needed.

This is particularly useful for:

-   Working with large datasets locally.
-   Batch-publishing new content or updates.
-   Backing up collection data.
-   Migrating data between local storage and Numerous.

Files are transferred as binary data, preserving their content. Documents are represented as JSON files locally (with configurable suffix) and are stored as structured data in Numerous.

**Note**: The current implementation uses `.collection-doc.json` as the default suffix for documents. This provides a clear distinction between files and documents and avoids naming conflicts. You can customize this suffix using the `document_suffix` parameter.

## Getting Started

Here's a quick example of how to use the bulk download and upload functions:

```python
from pathlib import Path
from numerous.collections import collection
from numerous.collections.bulk import bulk_download, bulk_upload

# 1. Get a reference to your target collection
my_collection = collection("my-main-data-collection")

# 2. Bulk Download
# This will download 'my-main-data-collection' into:
# ./collections/my-main-data-collection/
# Documents will be saved with .collection-doc.json suffix
print(f"Starting download for {my_collection.key}...")
bulk_download(my_collection)
print("Download complete.")

# To download to a custom local base directory with custom document suffix:
# custom_local_path = Path("/mnt/data/numerous_exports")
# bulk_download(my_collection, 
#               local_base_path=custom_local_path,
#               document_suffix=".doc.json")
# This would create: /mnt/data/numerous_exports/my-main-data-collection/
# Documents would be saved as: config.doc.json, settings.doc.json, etc.

# 3. Bulk Upload
# Assume you have a local directory structure like:
# ./collections_to_upload/my-main-data-collection/
#   ├── notes.txt                           # Regular file
#   ├── config.collection-doc.json          # Document (becomes "config" document)
#   ├── data.json                           # Regular JSON file
#   └── subfolder/
#       ├── report.docx                     # Regular file
#       └── settings.collection-doc.json    # Document (becomes "settings" document)

# This will upload the contents of './collections_to_upload/my-main-data-collection/'
# into the 'my-main-data-collection' in Numerous.
upload_source_dir = Path("./collections_to_upload") # Base path containing the collection key directory
print(f"Starting upload to {my_collection.key} from {upload_source_dir / my_collection.key}...")
bulk_upload(my_collection, local_base_path=upload_source_dir)
print("Upload complete.")

# To upload with custom settings:
# custom_upload_source = Path("./project_data")
# bulk_upload(my_collection,
#             local_base_path=custom_upload_source,
#             ignore_file_name=".customignore",
#             document_suffix=".doc.json")

## API Reference

### Function Signatures

```python
from numerous.collections import bulk_download, bulk_upload

# Download collections to local directory
bulk_download(
    collection: CollectionReference,
    local_path: str | Path,
    ignore_file: str | Path | None = None,
    document_suffix: str = ".json"
) -> None

# Upload local directory to collections
bulk_upload(
    collection: CollectionReference,
    local_path: str | Path,
    ignore_file: str | Path | None = None,
    document_suffix: str = ".json"
) -> None
```

**Parameters:**
- `collection`: The root collection to download from or upload to
- `local_path`: Local directory path for the bulk operation
- `ignore_file`: Optional path to ignore patterns file (defaults to `.numerous-ignore` in local_path)
- `document_suffix`: File extension for document files (defaults to ".json")

### Document Suffix Configuration

By default, document files are saved with a `.json` extension. You can customize this using the `document_suffix` parameter:

```python
# Use .txt extension for documents
bulk_download(my_collection, "./data", document_suffix=".txt")

# Use .doc extension for documents  
bulk_upload(my_collection, "./data", document_suffix=".doc")

# Use no extension for documents
bulk_download(my_collection, "./data", document_suffix="")
```

**Important Notes:**
- The suffix is used to identify document files during upload operations
- During download, documents are saved with the specified suffix
- File references are always saved with their original names/extensions
- The suffix must include the dot (e.g., ".json", ".txt") unless you want no extension

### `bulk_download`

```python
def bulk_download(
    collection_ref: CollectionReference,
    local_base_path: Path = Path("collections"),
    document_suffix: str = ".collection-doc.json",
) -> None:
```

Recursively downloads an entire collection hierarchy to the local filesystem.

-   **`collection_ref`**: The `CollectionReference` of the root collection to download.
-   **`local_base_path`**: The local base directory where the collection data will be saved. Defaults to `"collections"` in the current working directory. The actual download occurs into a subdirectory named after the root collection's key (e.g., `local_base_path/collection_key/...`).
-   **`document_suffix`**: Suffix to append to document filenames when saving locally. Defaults to `".collection-doc.json"`.

Downloads all files (as binary) and documents (as JSON files with the specified suffix) from the specified collection and all its nested sub-collections. The local directory structure mirrors the collection hierarchy. Existing local files are overwritten.

### `bulk_upload`

```python
def bulk_upload(
    collection_ref: CollectionReference,
    local_base_path: Path = Path("collections"),
    ignore_file_name: str = ".numerous-ignore",
    document_suffix: str = ".collection-doc.json",
) -> None:
```

Recursively uploads a local directory structure to a Numerous collection.

-   **`collection_ref`**: The `CollectionReference` of the root collection to upload into.
-   **`local_base_path`**: The local base directory from which to upload. The function expects the content to be uploaded to be inside a subdirectory named after the root collection's key (e.g., `local_base_path/collection_key/...`).
-   **`ignore_file_name`**: Name of the ignore file (e.g., `".numerous-ignore"`) located in the root of the directory being uploaded (i.e., inside `local_base_path/collection_key/`).
-   **`document_suffix`**: Suffix that identifies files to be uploaded as documents. These files will have the suffix removed from their key. Defaults to `".collection-doc.json"`.

Scans the specified local directory. Files matching patterns in the ignore file are skipped. Files with the document suffix are uploaded as Numerous documents (with the suffix removed from their key). Other files are uploaded as Numerous files. Nested directories in the local structure result in the creation of corresponding nested collections in Numerous. Existing remote files and documents are overwritten.

## Ignore File Guide (`.numerous-ignore`)

The `bulk_upload` function supports a `.numerous-ignore` file (or a custom-named file via `ignore_file_name`) to exclude files and directories from the upload process. This file uses gitignore-style syntax.

-   **Location**: The ignore file should be placed in the root of the specific collection directory being uploaded (e.g., `local_base_path/collection_key/.numerous-ignore`).
-   **Syntax**:
    -   Lines starting with `#` are treated as comments and are ignored.
    -   Blank lines are ignored.
    -   Patterns are typically [glob patterns](https://en.wikipedia.org/wiki/Glob_(programming)).
    -   `*` matches any number of characters (except path separators).
    -   `**` matches any number of characters including path separators (useful for matching directories recursively or files in any subdirectory).
    -   `?` matches any single character (except path separators).
    -   A pattern ending with a `/` specifically matches a directory.
    -   Patterns are matched relative to the location of the ignore file.
    -   Leading `!` negates a pattern; if a file matches a negated pattern and an earlier non-negated pattern, it will be included.

**Example `.numerous-ignore` file:**

```
# Ignore common temporary and system files
*.log
*.tmp
.DS_Store
Thumbs.db

# Ignore Python virtual environments and cache
.venv/
__pycache__/
*.pyc

# Ignore build artifacts
build/
dist/
*.egg-info/

# Ignore large media files by extension
*.mp4
*.zip
*.gz

# But make sure to include a specific important log file
!important_process.log

# Ignore all markdown files in a 'drafts' subdirectory anywhere
**/drafts/**/*.md
```

## Error Handling

-   Both `bulk_download` and `bulk_upload` will raise exceptions for critical errors such as network failures, permission issues, or if the root local path for upload doesn't exist.
-   During processing of individual files or documents:
    -   If an error occurs (e.g., a single file fails to download/upload, a JSON document is malformed), an error message will be logged, and the operation will attempt to continue with the next items.
-   Invalid Keys: If a local file or directory name is invalid for use as a Numerous collection or entity key during `bulk_upload` (e.g., contains forbidden characters), a warning will be logged, and that specific item will be skipped.

### Files and Documents: Separate Entity Types

Files and documents are **separate entity types** in Numerous Collections, so they **never conflict** with each other:

**Normal Operation:**
Files and documents with the same name coexist naturally without any conflicts:

**Remote Collection Scenario:**
- Collection contains: file `config` + document `config`
- `bulk_download()` creates:
  - `config` (regular file with original content)
  - `config.collection-doc.json` (document content as JSON)
- Both files coexist locally without overwriting each other

**Local Directory Scenario:**
- Local directory contains: `config.json` (regular file) + `config.collection-doc.json` (document file)
- `bulk_upload()` creates:
  - File named `config.json` in remote collection
  - Document named `config` in remote collection (suffix stripped)
- Both coexist as different entity types in the remote collection

**Processing Order Details:**

**During `bulk_download`:**
- Files are processed first, then documents
- If a collection contains both a file named `config` and a document named `config`, the download will create:
  1. `config` (the file, downloaded as-is)
  2. `config.collection-doc.json` (the document, serialized to JSON with suffix)
- No conflict occurs since documents get a distinct suffix

**During `bulk_upload`:**
- If your local directory contains both `config.json` (a regular file) and `config.collection-doc.json` (a document file), both will be uploaded:
  1. `config.json` → uploaded as a Numerous file with key `config.json`
  2. `config.collection-doc.json` → uploaded as a Numerous document with key `config` (suffix removed)
- Both operations succeed; the remote collection will contain both entities as different types

**Key Benefits:**
- **No conflicts ever**: Files and documents are separate entity types that coexist naturally
- **Simple workflow**: Document suffix is just for local identification - no conflict resolution needed
- **Safe operations**: Work with any combination of files and documents without worry
- **Predictable behavior**: Everything works seamlessly because entity types are separate

Consult the SDK's general error handling documentation for more details on common exception types like `GraphQLClientError`.

## Best Practices

-   **Test with a Sample Collection**: Before running bulk operations on critical data, test your setup and ignore patterns with a smaller, non-critical sample collection.
-   **Monitor Logs**: The bulk operations log information about their progress and any errors encountered. Check these logs to verify the process.
-   **Understand Overwrites**: Both download and upload operations will overwrite existing data (local files for download, remote Numerous entities for upload) without prompting. Ensure this is the desired behavior.
-   **Valid Keys**: Be mindful of valid characters for collection, file, and document keys when preparing local directories for upload. Avoid special characters like `/`, `\`, `:`, `*`, `?`, `"`, `<`, `>`, `|`, and control characters.
