# ruff: noqa: T201, D400, D415
"""
Example: Basic Collection Bulk Download.

This example demonstrates how to download an entire collection
from Numerous to your local filesystem using bulk_download().

The bulk_download function will:
- Download all files and documents from the collection
- Create nested directories for sub-collections
- Save documents as JSON files with configurable suffix
- Preserve the collection structure locally
"""

from pathlib import Path

from numerous.collections import CollectionReference, collection
from numerous.collections.bulk import bulk_download


def main() -> None:
    """Download a collection using bulk operations."""
    # Initialize your collection (replace with your actual collection key)
    my_collection: CollectionReference = collection("my-collection-key")

    # Option 1: Download to default location (./collections/) with default suffix
    print("Downloading collection to ./collections/...")
    print("Documents will be saved with .collection-doc.json suffix")
    bulk_download(my_collection)
    print("✅ Download complete! Check the ./collections/ directory")

    # Option 2: Download to a custom location with default suffix
    custom_path = Path("./my-downloads")
    print(f"Downloading collection to {custom_path}...")
    bulk_download(my_collection, local_base_path=custom_path)
    print(f"✅ Download complete! Check the {custom_path} directory")

    # Option 3: Download with custom document suffix
    print("Downloading with custom document suffix (.doc.json)...")
    bulk_download(my_collection, document_suffix=".doc.json")
    print("✅ Download complete! Documents saved with .doc.json suffix")

    # Download with default settings (documents saved as .json files)
    print("Downloading collection to ./my_local_data...")
    bulk_download(my_collection, Path("./my_local_data"))
    print("Download completed!")

    # Example: Download with custom document suffix
    print("\nDownloading collection with .txt document suffix...")
    bulk_download(my_collection, Path("./my_local_data_txt"), document_suffix=".txt")
    print("Download with custom suffix completed!")

    # Example: Download with no document suffix
    print("\nDownloading collection with no document suffix...")
    bulk_download(my_collection, Path("./my_local_data_no_suffix"), document_suffix="")
    print("Download with no suffix completed!")

    # The downloaded structure will look like:
    # ./collections/my-collection-key/
    # ├── file1.txt                           # Regular file
    # ├── document1.collection-doc.json       # Document (from "document1" document)
    # ├── data.json                           # Regular JSON file
    # ├── config.collection-doc.json          # Document (from "config" document)
    # ├── subcollection/
    # │   ├── file2.txt
    # │   └── settings.collection-doc.json    # Document (from "settings" document)
    # └── another-subcollection/
    #     └── nested-file.txt


if __name__ == "__main__":
    # Since bulk operations are synchronous, we don't need asyncio here.
    # If your main script were async for other reasons, you could keep it.
    main()
