from __future__ import annotations

import json
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import TYPE_CHECKING, Any, BinaryIO

from numerous.collection.file_reference import FileReference
from numerous.generated.graphql.fragments import (
    CollectionDocumentReference,
    CollectionDocumentReferenceTags,
    CollectionFileReferenceTags,
    CollectionReference,
)
from numerous.jsonbase64 import base64_to_dict, dict_to_base64


if TYPE_CHECKING:
    from numerous.generated.graphql.input_types import TagInput


@dataclass
class FileSystemCollectionTag:
    key: str
    value: str

    @staticmethod
    def load(tag: dict[str, Any]) -> FileSystemCollectionTag:
        key = tag.get("key")
        if not isinstance(key, str):
            tname = type(key).__name__
            msg = f"FileSystemCollectionTag key must be str, found {tname}"
            raise TypeError(msg)

        value = tag.get("value")
        if not isinstance(value, str):
            tname = type(value).__name__
            msg = f"FileSystemCollectionTag value must be str, found {tname}"
            raise TypeError(msg)

        return FileSystemCollectionTag(key=key, value=value)

    def to_file_reference_tag(self) -> CollectionFileReferenceTags:
        return CollectionFileReferenceTags(
            key=self.key,
            value=self.value,
        )

    def to_document_reference_tag(self) -> CollectionDocumentReferenceTags:
        return CollectionDocumentReferenceTags(
            key=self.key,
            value=self.value,
        )


@dataclass
class FileSystemFileMetadata:
    file_id: str
    file_key: str
    tags: list[FileSystemCollectionTag]

    def save(self, path: Path) -> None:
        def convert_to_serializable(obj: Path) -> str:
            if isinstance(obj, Path):
                return str(obj)
            return ""

        with path.open("w") as f:
            json.dump(asdict(self), f, default=convert_to_serializable)

    @staticmethod
    def load(file_path: Path) -> FileSystemFileMetadata:
        with file_path.open("r") as f:
            file_content = json.load(f)

        if not isinstance(file_content, dict):
            tname = type(file_content).__name__
            msg = f"FileSystemCollection file must be a dict, found {tname}"
            raise TypeError(msg)

        file_id = file_content.get("file_id")
        if not isinstance(file_id, str):
            tname = type(file_content).__name__
            msg = f"FileSystemCollection file id must be a str, found {tname}"
            raise TypeError(msg)

        file_key = file_content.get("file_key")
        if not isinstance(file_key, str):
            tname = type(file_content).__name__
            msg = f"FileSystemCollection file id must be a str, found {tname}"
            raise TypeError(msg)

        tags = file_content.get("tags", [])
        if not isinstance(tags, list):
            tname = type(tags).__name__
            msg = f"FileSystemCollection tags must be a list, found {tname}"
            raise TypeError(msg)

        return FileSystemFileMetadata(
            file_id=file_id,
            file_key=file_key,
            tags=[FileSystemCollectionTag.load(tag) for tag in tags],
        )

    def reference_tags(self) -> list[CollectionFileReferenceTags]:
        return [
            CollectionFileReferenceTags(key=tag.key, value=tag.value)
            for tag in self.tags
        ]

    def tag_matches(self, tag_input: TagInput) -> bool:
        matching_tag = next(
            (
                tag
                for tag in self.tags
                if tag.key == tag_input.key and tag.value == tag_input.value
            ),
            None,
        )

        return matching_tag is not None


@dataclass
class FileSystemCollectionDocument:
    data: dict[str, Any]
    tags: list[FileSystemCollectionTag]

    def save(self, path: Path) -> None:
        with path.open("w") as f:
            json.dump(asdict(self), f)

    @staticmethod
    def load(path: Path) -> FileSystemCollectionDocument:
        with path.open("r") as f:
            file_content = json.load(f)

        if not isinstance(file_content, dict):
            tname = type(file_content).__name__
            msg = f"FileSystemCollection document must be a dict, found {tname}"
            raise TypeError(msg)

        tags = file_content.get("tags", [])
        if not isinstance(tags, list):
            tname = type(tags).__name__
            msg = f"FileSystemCollection tags must be a list, found {tname}"
            raise TypeError(msg)

        data = file_content.get("data", {})
        if not isinstance(data, dict):
            tname = type(data).__name__
            msg = f"FileSystemCollection data must be a dict, found {tname}"
            raise TypeError(msg)

        return FileSystemCollectionDocument(
            data=data, tags=[FileSystemCollectionTag.load(tag) for tag in tags]
        )

    def reference_tags(self) -> list[CollectionDocumentReferenceTags]:
        return [
            CollectionDocumentReferenceTags(
                key=tag.key,
                value=tag.value,
            )
            for tag in self.tags
        ]

    def tag_matches(self, tag_input: TagInput) -> bool:
        matching_tag = next(
            (
                tag
                for tag in self.tags
                if tag.key == tag_input.key and tag.value == tag_input.value
            ),
            None,
        )

        return matching_tag is not None


@dataclass
class FileIndexEntry:
    collection_id: str
    file_key: str
    _path: Path | None = None

    @staticmethod
    def load(path: Path) -> FileIndexEntry:
        return FileIndexEntry(**json.loads(path.read_text()), _path=path)

    def remove(self) -> None:
        if self._path is not None:
            self._path.unlink()

    def save(self, path: Path) -> None:
        if self._path:
            msg = "Cannot save file index entry that was already saved."
            raise RuntimeError(msg)
        self._path = path
        path.write_text(
            json.dumps({"collection_id": self.collection_id, "file_key": self.file_key})
        )


class FileSystemClient:
    FILE_INDEX_DIR = "__file_index__"

    def __init__(self, base_path: Path) -> None:
        self._base_path = base_path
        self._base_path.mkdir(exist_ok=True)
        (self._base_path / self.FILE_INDEX_DIR).mkdir(exist_ok=True)

    def _file_index_entry(self, file_id: str) -> FileIndexEntry:
        return FileIndexEntry.load(self._base_path / self.FILE_INDEX_DIR / file_id)

    def _file_metadata_path(self, collection_id: str, file_key: str) -> Path:
        return self._base_path / collection_id / f"{_escape(file_key)}.file.meta.json"

    def _file_data_path(self, collection_id: str, file_key: str) -> Path:
        return self._base_path / collection_id / f"{_escape(file_key)}.file.data"

    def _document_path(self, collection_id: str, document_key: str) -> Path:
        return self._base_path / collection_id / f"{document_key}.doc.json"

    def _document_path_from_id(self, document_id: str) -> Path:
        return self._base_path / (document_id + ".doc.json")

    def get_collection_reference(
        self, collection_key: str, parent_collection_id: str | None = None
    ) -> CollectionReference:
        collection_relpath = (
            Path(parent_collection_id) / collection_key
            if parent_collection_id is not None
            else Path(collection_key)
        )
        collection_id = str(collection_relpath)
        collection_path = self._base_path / collection_relpath
        collection_path.mkdir(parents=True, exist_ok=True)
        return CollectionReference(id=collection_id, key=collection_key)

    def get_collection_document(
        self, collection_id: str, document_key: str
    ) -> CollectionDocumentReference | None:
        path = self._document_path(collection_id, document_key)
        if not path.exists():
            return None

        doc = FileSystemCollectionDocument.load(path)

        doc_id = str(Path(collection_id) / document_key)
        return CollectionDocumentReference(
            id=doc_id,
            key=document_key,
            data=dict_to_base64(doc.data),
            tags=[tag.to_document_reference_tag() for tag in doc.tags],
        )

    def set_collection_document(
        self, collection_id: str, document_key: str, encoded_data: str
    ) -> CollectionDocumentReference | None:
        path = self._document_path(collection_id, document_key)
        data = base64_to_dict(encoded_data)
        if path.exists():
            doc = FileSystemCollectionDocument.load(path)
            doc.data = data
        else:
            doc = FileSystemCollectionDocument(data, [])
        doc.save(path)

        doc_id = str(Path(collection_id) / document_key)
        return CollectionDocumentReference(
            id=doc_id,
            key=document_key,
            data=encoded_data,
            tags=[],
        )

    def delete_collection_document(
        self, document_id: str
    ) -> CollectionDocumentReference | None:
        doc_path = self._document_path_from_id(document_id)
        if not doc_path.exists():
            return None

        doc = FileSystemCollectionDocument.load(doc_path)

        doc_path.unlink()

        return CollectionDocumentReference(
            id=document_id,
            key=doc_path.name.removesuffix(".doc.json"),
            data=dict_to_base64(doc.data),
            tags=doc.reference_tags(),
        )

    def add_collection_document_tag(
        self, document_id: str, tag: TagInput
    ) -> CollectionDocumentReference | None:
        doc_path = self._document_path_from_id(document_id)
        if not doc_path.exists():
            return None

        doc = FileSystemCollectionDocument.load(doc_path)
        doc.tags.append(FileSystemCollectionTag(key=tag.key, value=tag.value))
        doc.save(doc_path)

        return CollectionDocumentReference(
            id=document_id,
            key=doc_path.stem,
            data=dict_to_base64(doc.data),
            tags=doc.reference_tags(),
        )

    def delete_collection_document_tag(
        self, document_id: str, tag_key: str
    ) -> CollectionDocumentReference | None:
        doc_path = self._document_path_from_id(document_id)
        if not doc_path.exists():
            return None

        doc = FileSystemCollectionDocument.load(doc_path)
        doc.tags = [tag for tag in doc.tags if tag.key != tag_key]
        doc.save(doc_path)

        return CollectionDocumentReference(
            id=document_id,
            key=doc_path.stem,
            data=dict_to_base64(doc.data),
            tags=doc.reference_tags(),
        )

    def get_collection_documents(
        self,
        collection_id: str,
        end_cursor: str,  # noqa: ARG002
        tag_input: TagInput | None,
    ) -> tuple[list[CollectionDocumentReference | None], bool, str]:
        col_path = self._base_path / collection_id
        if not col_path.exists():
            return [], False, ""

        documents: list[CollectionDocumentReference | None] = []
        for doc_path in col_path.iterdir():
            if not doc_path.name.endswith(".doc.json"):
                continue

            doc = FileSystemCollectionDocument.load(doc_path)

            if tag_input and not doc.tag_matches(tag_input):
                # skips files that do not match tag input, if it is given
                continue

            doc_id = str(
                doc_path.relative_to(self._base_path).with_name(
                    doc_path.name.removesuffix(".doc.json")
                )
            )
            documents.append(
                CollectionDocumentReference(
                    id=doc_id,
                    key=doc_path.name.removesuffix(".doc.json"),
                    data=dict_to_base64(doc.data),
                    tags=doc.reference_tags(),
                )
            )

        return sorted(documents, key=lambda d: d.id if d else ""), False, ""

    def create_collection_file_reference(
        self, collection_id: str, file_key: str
    ) -> FileReference | None:
        meta_path = self._file_metadata_path(collection_id, file_key)
        if meta_path.exists():
            meta = FileSystemFileMetadata.load(meta_path)
        else:
            file_id = _escape(collection_id + "_" + file_key)
            index_entry = FileIndexEntry(collection_id=collection_id, file_key=file_key)
            index_entry.save(self._base_path / self.FILE_INDEX_DIR / file_id)
            meta = FileSystemFileMetadata(file_id=file_id, file_key=file_key, tags=[])
            meta.save(self._file_metadata_path(collection_id, file_key))

        return FileReference(
            client=self,
            file_id=meta.file_id,
            key=file_key,
        )

    def collection_file_tags(self, file_id: str) -> dict[str, str] | None:
        try:
            index_entry = self._file_index_entry(file_id)
        except FileNotFoundError:
            return None

        meta_path = self._file_metadata_path(
            index_entry.collection_id, index_entry.file_key
        )

        if not meta_path.exists():
            return None

        meta = FileSystemFileMetadata.load(meta_path)
        return {tag.key: tag.value for tag in meta.tags}

    def delete_collection_file(self, file_id: str) -> None:
        index_entry = self._file_index_entry(file_id)
        meta_path = self._file_metadata_path(
            index_entry.collection_id, index_entry.file_key
        )
        data_path = self._file_data_path(
            index_entry.collection_id, index_entry.file_key
        )

        if not meta_path.exists():
            return

        meta_path.unlink()
        data_path.unlink()

    def get_collection_files(
        self,
        collection_id: str,
        end_cursor: str,  # noqa: ARG002
        tag_input: TagInput | None,
    ) -> tuple[list[FileReference], bool, str]:
        col_path = self._base_path / collection_id
        if not col_path.exists():
            return [], False, ""

        files: list[FileReference] = []
        for file_path in col_path.iterdir():
            if not file_path.name.endswith(".file.meta.json"):
                continue

            meta = FileSystemFileMetadata.load(file_path)

            if tag_input and not meta.tag_matches(tag_input):
                # skips files that do not match tag input, if it is given
                continue

            files.append(
                FileReference(client=self, file_id=meta.file_id, key=meta.file_key)
            )

        return files, False, ""

    def add_collection_file_tag(self, file_id: str, tag: TagInput) -> None:
        index_entry = self._file_index_entry(file_id)
        meta_path = self._file_metadata_path(
            index_entry.collection_id, index_entry.file_key
        )
        if not meta_path.exists():
            return

        meta = FileSystemFileMetadata.load(meta_path)
        if not meta.tag_matches(tag):
            meta.tags.append(FileSystemCollectionTag(key=tag.key, value=tag.value))
        meta.save(meta_path)

    def delete_collection_file_tag(self, file_id: str, tag_key: str) -> None:
        index_entry = self._file_index_entry(file_id)
        meta_path = self._file_metadata_path(
            index_entry.collection_id, index_entry.file_key
        )
        if not meta_path.exists():
            return

        meta = FileSystemFileMetadata.load(meta_path)
        meta.tags = [tag for tag in meta.tags if tag.key != tag_key]
        meta.save(meta_path)

    def get_collection_collections(
        self,
        collection_key: str,
        end_cursor: str,  # noqa: ARG002
    ) -> tuple[list[CollectionReference], bool, str]:
        col_path = self._base_path / collection_key
        if not col_path.exists():
            return [], False, ""

        collections: list[CollectionReference] = []
        for item in col_path.iterdir():
            if item.is_dir():
                col_id = str(item.relative_to(self._base_path))
                collections.append(CollectionReference(id=col_id, key=item.name))

        return sorted(collections, key=lambda c: c.id), False, ""

    def read_text(self, file_id: str) -> str:
        index_entry = self._file_index_entry(file_id)
        data_path = self._file_data_path(
            index_entry.collection_id, index_entry.file_key
        )
        return data_path.read_text()

    def read_bytes(self, file_id: str) -> bytes:
        index_entry = self._file_index_entry(file_id)
        data_path = self._file_data_path(
            index_entry.collection_id, index_entry.file_key
        )
        return data_path.read_bytes()

    def save_file(self, file_id: str, data: bytes | str) -> None:
        index_entry = self._file_index_entry(file_id)
        data_path = self._file_data_path(
            index_entry.collection_id, index_entry.file_key
        )
        if isinstance(data, bytes):
            data_path.write_bytes(data)
        else:
            data_path.write_text(data)

    def open_file(self, file_id: str) -> BinaryIO:
        index_entry = self._file_index_entry(file_id)
        data_path = self._file_data_path(
            index_entry.collection_id, index_entry.file_key
        )
        return data_path.open("rb")

    def file_exists(self, file_id: str) -> bool:
        try:
            index_entry = self._file_index_entry(file_id)
        except FileNotFoundError:
            return False

        return self._file_data_path(
            index_entry.collection_id, index_entry.file_key
        ).exists()


def _escape(key: str) -> str:
    return key.replace("/", "__")
