# ruff: noqa: T201, BLE001, G201, G004, PLR2004, TRY003, EM101, F841, ANN204, INP001
"""
Advanced Collection Bulk Operations Example.

This comprehensive example demonstrates advanced usage patterns for bulk
download and upload operations with Numerous collections, including:

- Complex directory structures with multiple levels
- Custom local paths and ignore file configurations
- Error handling and recovery strategies
- Performance optimization techniques
- Best practices for production usage

This example covers real-world scenarios that developers commonly encounter
when working with large collections and complex file hierarchies.
"""

import json
import logging
import shutil
from pathlib import Path
from typing import Optional

from numerous.collections import CollectionReference, collection
from numerous.collections.bulk import bulk_download, bulk_upload


# Configure logging for the example
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    handlers=[logging.StreamHandler(), logging.FileHandler("bulk_operations.log")],
)
logger = logging.getLogger(__name__)

# Constants for demo
DEMO_BASE_PATH = Path("./advanced-demo")
DEMO_COLLECTION_KEY = "advanced-collection-demo"
MAX_TREE_ITEMS = 10
COLLECTION_NOT_INITIALIZED_ERROR = "Collection not initialized"


class BulkOperationsDemo:
    """Demonstration class for advanced bulk operations."""

    def __init__(self, base_path: Path) -> None:
        """
        Initialize the demo with a base working directory.

        Args:
            base_path: Base directory for all demo operations

        """
        self.base_path = base_path
        self.collection_obj: Optional[CollectionReference] = None

    def setup_collection(self, collection_key: str) -> None:
        """
        Initialize and configure the collection for demo operations.

        Args:
            collection_key: The key of the collection to work with

        """
        try:
            self.collection_obj = collection(collection_key)
            logger.info("Collection '%s' initialized successfully", collection_key)
        except Exception:
            logger.exception("Failed to initialize collection '%s'", collection_key)
            raise

    def create_complex_directory_structure(self) -> Path:
        """
        Create a comprehensive directory structure for testing.

        Returns:
            Path to the created directory structure

        """
        # Define the complex structure
        structure_path = self.base_path / "complex-project"

        # Create project directories
        directories = [
            "src/main/python/modules",
            "src/test/python/unit_tests",
            "src/test/python/integration_tests",
            "docs/api/reference",
            "docs/user/guides",
            "docs/developer/architecture",
            "data/raw/datasets",
            "data/processed/analysis",
            "data/external/imports",
            "config/environments/dev",
            "config/environments/staging",
            "config/environments/prod",
            "scripts/deployment",
            "scripts/maintenance",
            "logs/application",
            "logs/system",
            "temp/uploads",
            "temp/processing",
            "cache/downloads",
            "backup/daily",
        ]

        logger.info("Creating complex directory structure...")
        for dir_path in directories:
            full_path = structure_path / dir_path
            full_path.mkdir(parents=True, exist_ok=True)

        # Create various file types across the structure
        files_to_create = [
            # Source code files
            (
                "src/main/python/main.py",
                "# Main application entry point\nprint('Hello World')",
            ),
            (
                "src/main/python/modules/auth.py",
                "# Authentication module\nclass AuthManager: pass",
            ),
            (
                "src/main/python/modules/config.json",
                '{"app_name": "Demo App", "version": "1.0.0"}',
            ),
            # Test files
            (
                "src/test/python/unit_tests/test_auth.py",
                "# Unit tests for auth\nimport unittest",
            ),
            (
                "src/test/python/integration_tests/test_api.py",
                "# API integration tests\nimport requests",
            ),
            # Documentation files
            (
                "docs/api/reference/endpoints.md",
                "# API Endpoints\n## User Management\n### GET /users",
            ),
            (
                "docs/user/guides/quickstart.md",
                "# Quick Start Guide\n## Installation\n1. Clone repo",
            ),
            (
                "docs/developer/architecture/overview.md",
                "# Architecture Overview\n## Components",
            ),
            # Data files
            (
                "data/raw/datasets/users.csv",
                "id,name,email\n1,John,john@example.com\n2,Jane,jane@example.com",
            ),
            (
                "data/processed/analysis/summary.json",
                '{"total_users": 2, "active_users": 1}',
            ),
            # Configuration files
            (
                "config/environments/dev/database.yaml",
                "host: localhost\nport: 5432\ndatabase: myapp_dev",
            ),
            (
                "config/environments/prod/database.yaml",
                "host: prod-db\nport: 5432\ndatabase: myapp_prod",
            ),
            # Scripts
            (
                "scripts/deployment/deploy.sh",
                "#!/bin/bash\necho 'Deploying application...'",
            ),
            (
                "scripts/maintenance/cleanup.py",
                "# Cleanup script\nimport os\nprint('Cleaning up...')",
            ),
            # Log files (will be ignored)
            ("logs/application/app.log", "2024-01-01 INFO Application started"),
            ("logs/system/system.log", "2024-01-01 INFO System ready"),
            # Temporary files (will be ignored)
            ("temp/uploads/temp_file.tmp", "temporary content"),
            ("temp/processing/process.tmp", "processing data"),
            # Cache files (will be ignored)
            ("cache/downloads/cache.cache", "cached data"),
        ]

        logger.info(
            "Creating %d files across the directory structure...", len(files_to_create)
        )
        for file_path, content in files_to_create:
            full_file_path = structure_path / file_path
            full_file_path.write_text(content)

        return structure_path

    def create_comprehensive_ignore_file(self, target_dir: Path) -> None:
        """
        Create a comprehensive .numerous-ignore file.

        Args:
            target_dir: Directory where the ignore file should be created

        """
        ignore_content = """
# =============================================================================
# Comprehensive .numerous-ignore for Advanced Demo
# =============================================================================

# Temporary and cache files
*.tmp
*.cache
*.temp
temp/
cache/
.cache/

# Log files
*.log
logs/
.logs/

# System files
.DS_Store
Thumbs.db
desktop.ini
.Spotlight-V100
.Trashes

# Editor and IDE files
.vscode/
.idea/
*.swp
*.swo
*~
.sublime-project
.sublime-workspace

# Version control
.git/
.gitignore
.gitkeep
.svn/

# Build artifacts
build/
dist/
target/
*.egg-info/
__pycache__/
*.pyc
*.pyo
*.pyd
.pytest_cache/

# Package manager files
node_modules/
vendor/
.npm/
.yarn/

# Environment files
.env
.env.local
.env.*.local
venv/
.venv/
env/

# Database files
*.db
*.sqlite
*.sqlite3

# Large media files (examples)
*.mov
*.mp4
*.avi
*.mkv
*.iso

# Backup files
*.bak
*.backup
backup/
.backup/

# Documentation build
docs/_build/
docs/build/
site/

# Test coverage
.coverage
.nyc_output/
coverage/

# Specific to this demo - ignore processing temp files
temp/processing/
data/temp/
"""

        ignore_file_path = target_dir / ".numerous-ignore"
        ignore_file_path.write_text(ignore_content.strip())
        logger.info("Created comprehensive ignore file at %s", ignore_file_path)

    def demonstrate_bulk_download(self) -> None:
        """Demonstrate advanced bulk download operations."""
        if not self.collection_obj:
            raise ValueError(COLLECTION_NOT_INITIALIZED_ERROR)

        print("\n" + "=" * 60)
        print("🔽 BULK DOWNLOAD DEMONSTRATION")
        print("=" * 60)

        download_path = self.base_path / "downloads"

        try:
            # Download with custom path
            print(f"📥 Downloading collection to: {download_path}")
            logger.info("Starting bulk download to %s", download_path)

            bulk_download(self.collection_obj, local_base_path=download_path)

            print("✅ Download completed successfully!")
            logger.info("Bulk download completed successfully")

            # Display what was downloaded
            if download_path.exists():
                print("\n📁 Downloaded content structure:")
                self._display_directory_tree(download_path, max_depth=3)

        except Exception:
            print("❌ Download failed")
            logger.exception("Bulk download failed")
            raise

    def demonstrate_bulk_upload(self) -> None:
        """Demonstrate advanced bulk upload operations."""
        if not self.collection_obj:
            raise ValueError(COLLECTION_NOT_INITIALIZED_ERROR)

        print("\n" + "=" * 60)
        print("🔼 BULK UPLOAD DEMONSTRATION")
        print("=" * 60)

        # Create the complex structure
        source_path = self.create_complex_directory_structure()
        self.create_comprehensive_ignore_file(source_path)

        try:
            print(f"📤 Uploading from: {source_path}")
            print(f"📋 Using ignore file: {source_path / '.numerous-ignore'}")
            logger.info("Starting bulk upload from %s", source_path)

            # Upload with custom ignore file
            bulk_upload(
                self.collection_obj,
                local_base_path=source_path.parent,
                ignore_file_name=".numerous-ignore",
            )

            print("✅ Upload completed successfully!")
            logger.info("Bulk upload completed successfully")

        except Exception:
            print("❌ Upload failed")
            logger.exception("Bulk upload failed")
            raise

    def _display_directory_tree(
        self, path: Path, max_depth: int = 2, current_depth: int = 0
    ) -> None:
        """
        Display directory tree structure.

        Args:
            path: Root path to display
            max_depth: Maximum depth to traverse
            current_depth: Current traversal depth

        """
        if current_depth >= max_depth:
            return

        try:
            items = sorted(path.iterdir())
            for i, item in enumerate(items[:MAX_TREE_ITEMS]):
                is_last = i == len(items) - 1
                prefix = "└── " if is_last else "├── "
                indent = "    " * current_depth

                if item.is_dir():
                    print(f"{indent}{prefix}{item.name}/")
                    if current_depth < max_depth - 1:
                        self._display_directory_tree(item, max_depth, current_depth + 1)
                else:
                    print(f"{indent}{prefix}{item.name}")

            if len(items) > MAX_TREE_ITEMS:
                print(
                    f"{'    ' * current_depth}└── ... "
                    f"({len(items) - MAX_TREE_ITEMS} more items)"
                )

        except Exception:
            logger.exception("Error displaying directory tree for %s", path)

    def demonstrate_error_handling(self) -> None:
        """Demonstrate proper error handling patterns."""
        print("\n" + "=" * 60)
        print("⚠️  ERROR HANDLING DEMONSTRATION")
        print("=" * 60)

        # Simulate various error conditions
        test_cases = [
            ("Invalid collection key", "non-existent-collection"),
            ("Invalid local path", Path("/invalid/path/that/does/not/exist")),
        ]

        for description, test_case in test_cases:
            print(f"\n🧪 Testing: {description}")
            try:
                if isinstance(test_case, str):
                    # Test invalid collection
                    collection(test_case)
                elif isinstance(test_case, Path) and self.collection_obj:
                    # Test invalid path
                    bulk_download(self.collection_obj, local_base_path=test_case)

            except Exception as e:
                print(f"✅ Expected error caught: {type(e).__name__}")
                logger.info(
                    "Expected error in test '%s': %s", description, type(e).__name__
                )

    def cleanup_demo_files(self) -> None:
        """Clean up demo files and directories."""
        try:
            if self.base_path.exists():
                shutil.rmtree(self.base_path)
                logger.info("Demo cleanup completed")
        except Exception:
            logger.exception("Error during cleanup")

    def demonstrate_advanced_document_suffix_usage(self) -> None:
        """Demonstrate advanced document suffix usage patterns."""
        if not self.collection_obj:
            raise ValueError(COLLECTION_NOT_INITIALIZED_ERROR)

        print("\n=== Advanced Document Suffix Usage ===")

        # Scenario 1: Working with XML documents
        print("1. Downloading collection with XML document format...")
        try:
            bulk_download(
                self.collection_obj, Path("./data/xml_format"), document_suffix=".xml"
            )
            print("✅ XML format download completed")
        except Exception as e:
            print(f"❌ XML download failed: {e}")

        # Scenario 2: No file extensions (clean filenames)
        print("2. Downloading with clean filenames (no extensions)...")
        try:
            bulk_download(
                self.collection_obj, Path("./data/clean_names"), document_suffix=""
            )
            print("✅ Clean filename download completed")
        except Exception as e:
            print(f"❌ Clean filename download failed: {e}")

        # Scenario 3: Custom suffix for specialized workflows
        print("3. Using custom .data suffix for specialized workflow...")
        try:
            bulk_download(
                self.collection_obj, Path("./data/specialized"), document_suffix=".data"
            )
            print("✅ Specialized workflow download completed")
        except Exception as e:
            print(f"❌ Specialized download failed: {e}")

        # Scenario 4: Upload matching the download format
        print("4. Uploading back with matching document suffix...")
        try:
            # Create some test data files
            Path("./test_upload").mkdir(parents=True, exist_ok=True)

            # Create a document file with custom suffix
            with Path("./test_upload/sample.data").open("w") as f:
                json.dump({"test": "data", "format": "custom"}, f)

            # Create a regular file
            with Path("./test_upload/readme.txt").open("w") as f:
                f.write("This is a regular file")

            bulk_upload(
                self.collection_obj, Path("./test_upload"), document_suffix=".data"
            )
            print("✅ Custom suffix upload completed")
        except Exception as e:
            print(f"❌ Custom suffix upload failed: {e}")


def main() -> None:
    """Run the advanced bulk operations demonstration."""
    print("🚀 Advanced Collection Bulk Operations Demo")
    print("=" * 60)

    # Initialize demo
    demo = BulkOperationsDemo(DEMO_BASE_PATH)

    try:
        # Setup collection (you'll need to replace with a real collection key)
        print("🔧 Initializing collection...")
        demo.setup_collection(DEMO_COLLECTION_KEY)

        # Run demonstrations
        demo.demonstrate_bulk_upload()
        demo.demonstrate_bulk_download()
        demo.demonstrate_error_handling()
        demo.demonstrate_advanced_document_suffix_usage()

        print("\n" + "=" * 60)
        print("🎉 DEMONSTRATION COMPLETE")
        print("=" * 60)
        print("Key takeaways:")
        print("• Use structured directory hierarchies for better organization")
        print("• Leverage ignore files to control what gets uploaded")
        print("• Implement proper error handling in production code")
        print("• Follow performance best practices for large datasets")
        print("• Monitor operations with appropriate logging")

    except Exception as e:
        print(f"\n❌ Demonstration failed: {e}")
        logger.exception("Main demonstration error")

    finally:
        # Cleanup
        print("\n🧹 Cleaning up demo files...")
        demo.cleanup_demo_files()
        print("✅ Cleanup complete!")


if __name__ == "__main__":
    # This example uses synchronous calls for simplicity.
    # If your application is asynchronous, you can use asyncio.run(main())
    # and adapt the synchronous calls accordingly.
    main()
