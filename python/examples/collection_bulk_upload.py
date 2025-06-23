# ruff: noqa: T201, ERA001, D400, D415, E501, INP001
"""
Example: Basic Collection Bulk Upload.

This example demonstrates how to upload a local directory structure
to a Numerous collection using bulk_upload().

The bulk_upload function will:
- Upload all files and JSON documents from a local directory
- Create nested collections for subdirectories
- Apply ignore file filtering (using .numerous-ignore)
- Preserve the directory structure in the collection
"""

from pathlib import Path

from numerous.collections import CollectionReference, collection
from numerous.collections.bulk import bulk_upload


def create_sample_data() -> Path:
    """Create sample directory structure for demonstration."""
    base_path = Path("./sample-collection")
    collection_path = base_path / "my-collection"

    # Create directories
    collection_path.mkdir(parents=True, exist_ok=True)
    (collection_path / "docs").mkdir(exist_ok=True)
    (collection_path / "data").mkdir(exist_ok=True)

    # Create sample files
    (collection_path / "README.txt").write_text("This is a sample collection")
    (collection_path / "data.json").write_text(
        '{"note": "This is a regular JSON file, not a document"}'
    )

    # Create document files (with .collection-doc.json suffix)
    (collection_path / "config.collection-doc.json").write_text(
        '{"version": "1.0", "name": "My Collection"}'
    )
    (collection_path / "docs" / "guide.txt").write_text("User guide content")
    (collection_path / "docs" / "metadata.collection-doc.json").write_text(
        '{"title": "Documentation", "pages": 10}'
    )
    (collection_path / "data" / "dataset.csv").write_text(
        "name,value\\nItem1,100\\nItem2,200"
    )

    # Create ignore file
    ignore_content = """
# Temporary files
*.tmp
*.log

# System files
.DS_Store
Thumbs.db

# Build artifacts
build/
dist/

# Note: .collection-doc.json files are NOT ignored - they become documents
"""
    (collection_path / ".numerous-ignore").write_text(ignore_content.strip())

    # Create files that will be ignored
    (collection_path / "temp.tmp").write_text("temporary file")
    (collection_path / "debug.log").write_text("log file")

    print(f"✅ Sample data created in {collection_path}")
    return base_path


def main() -> None:
    """Upload a directory structure using bulk operations."""
    # Create sample data
    base_path = create_sample_data()

    # Initialize your collection (replace with your actual collection key)
    my_collection: CollectionReference = collection("my-collection-key")

    # Option 1: Upload from default location with default settings
    print(f"Uploading from {base_path} to collection '{my_collection.key}'...")
    print("Files with .collection-doc.json suffix will become documents")
    bulk_upload(my_collection, local_base_path=base_path)
    print("✅ Upload complete!")

    # Option 2: Upload with custom settings
    # If you have a custom ignore file or document suffix, you can specify them:
    # bulk_upload(my_collection,
    #            local_base_path=base_path,
    #            ignore_file_name=".myignore",
    #            document_suffix=".doc.json")

    print("\\nUploaded structure:")
    print("my-collection/")
    print("├── README.txt                      # → file 'README.txt'")
    print("├── data.json                       # → file 'data.json'")
    print("├── config.collection-doc.json      # → document 'config'")
    print("├── docs/                           # → nested collection")
    print("│   ├── guide.txt                   # → file 'guide.txt'")
    print("│   └── metadata.collection-doc.json # → document 'metadata'")
    print("├── data/                           # → nested collection")
    print("│   └── dataset.csv                 # → file 'dataset.csv'")
    print("└── .numerous-ignore                # → file '.numerous-ignore'")
    print("\\nIgnored files (not uploaded):")
    print("├── temp.tmp                        # Matched *.tmp pattern")
    print("└── debug.log                       # Matched *.log pattern")

    # Upload with default settings (looks for .json document files)
    print("Uploading ./my_local_data to collection...")
    bulk_upload(my_collection, Path("./my_local_data"))
    print("Upload completed!")

    # Example: Upload with custom document suffix
    print("\\nUploading files with .txt document suffix...")
    bulk_upload(my_collection, Path("./my_local_data_txt"), document_suffix=".txt")
    print("Upload with custom suffix completed!")

    # Example: Upload with no document suffix (treats files without extensions as documents)
    print("\\nUploading files with no document suffix...")
    bulk_upload(my_collection, Path("./my_local_data_no_suffix"), document_suffix="")
    print("Upload with no suffix completed!")


if __name__ == "__main__":
    main()
