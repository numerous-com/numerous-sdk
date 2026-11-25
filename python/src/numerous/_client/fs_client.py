from __future__ import annotations

import json
import os
import tempfile
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any, BinaryIO
from urllib.parse import quote, unquote

from numerous._utils.jsonbase64 import base64_to_dict, dict_to_base64
from numerous.collections._client import (
    CollectionDocumentIdentifier,
    CollectionFileIdentifier,
    CollectionIdentifier,
    Tag,
)
from numerous.collections.exceptions import ParentCollectionNotFoundError


def _atomic_write_bytes(path: Path, data: bytes) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, tmp_name = tempfile.mkstemp(dir=path.parent, prefix=".tmp-")
    tmp_path = Path(tmp_name)
    try:
        with os.fdopen(fd, "wb") as f:
            f.write(data)
            f.flush()
            os.fsync(f.fileno())
        tmp_path.replace(path)
    except BaseException:
        tmp_path.unlink(missing_ok=True)
        raise


def _atomic_write_text(path: Path, content: str) -> None:
    _atomic_write_bytes(path, content.encode("utf-8"))


def _parse_tag(tag: dict[str, Any]) -> Tag:
    key = tag.get("key")
    if not isinstance(key, str):
        tname = type(key).__name__
        msg = f"Tag key must be str, found {tname}"
        raise TypeError(msg)

    value = tag.get("value")
    if not isinstance(value, str):
        tname = type(value).__name__
        msg = f"Tag value must be str, found {tname}"
        raise TypeError(msg)

    return Tag(key=key, value=value)


@dataclass
class FileSystemCollectionMetadata:
    collection_id: str
    collection_key: str
    tags: list[Tag]

    def save(self, path: Path) -> None:
        _atomic_write_text(path, json.dumps(asdict(self)))

    @staticmethod
    def load(path: Path) -> FileSystemCollectionMetadata:
        with path.open("r") as f:
            file_content = json.load(f)

        if not isinstance(file_content, dict):
            tname = type(file_content).__name__
            msg = f"FileSystemCollection metadata must be a dict, found {tname}"
            raise TypeError(msg)

        collection_id = file_content.get("collection_id")
        if not isinstance(collection_id, str):
            tname = type(collection_id).__name__
            msg = f"FileSystemCollection id must be a str, found {tname}"
            raise TypeError(msg)

        collection_key = file_content.get("collection_key")
        if not isinstance(collection_key, str):
            tname = type(collection_key).__name__
            msg = f"FileSystemCollection key must be a str, found {tname}"
            raise TypeError(msg)

        tags = file_content.get("tags", [])
        if not isinstance(tags, list):
            tname = type(tags).__name__
            msg = f"FileSystemCollection tags must be a list, found {tname}"
            raise TypeError(msg)

        return FileSystemCollectionMetadata(
            collection_id=collection_id,
            collection_key=collection_key,
            tags=[_parse_tag(tag) for tag in tags],
        )

    def reference_tags(self) -> list[Tag]:
        return [Tag(key=tag.key, value=tag.value) for tag in self.tags]

    def tag_matches(self, tag_input: Tag) -> bool:
        matching_tag = next(
            (
                tag
                for tag in self.tags
                if tag_input.key == tag.key and tag_input.value == tag.value
            ),
            None,
        )

        return matching_tag is not None


@dataclass
class FileSystemFileMetadata:
    file_id: str
    file_key: str
    tags: list[Tag]

    def save(self, path: Path) -> None:
        def convert_to_serializable(obj: Path) -> str:
            if isinstance(obj, Path):
                return str(obj)
            return ""

        _atomic_write_text(
            path, json.dumps(asdict(self), default=convert_to_serializable)
        )

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
            tags=[_parse_tag(tag) for tag in tags],
        )

    def reference_tags(self) -> list[Tag]:
        return [Tag(key=tag.key, value=tag.value) for tag in self.tags]

    def tag_matches(self, tag_input: Tag) -> bool:
        matching_tag = next(
            (
                tag
                for tag in self.tags
                if tag_input.key == tag.key and tag_input.value == tag.value
            ),
            None,
        )

        return matching_tag is not None


@dataclass
class FileSystemCollectionDocument:
    data: dict[str, Any]
    tags: list[Tag]

    def save(self, path: Path) -> None:
        _atomic_write_text(path, json.dumps(asdict(self)))

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
            data=data, tags=[_parse_tag(tag) for tag in tags]
        )

    def reference_tags(self) -> list[Tag]:
        return [Tag(key=tag.key, value=tag.value) for tag in self.tags]

    def tag_matches(self, tag_input: Tag) -> bool:
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
            self._path.unlink(missing_ok=True)

    def save(self, path: Path) -> None:
        if self._path:
            msg = "Cannot save file index entry that was already saved."
            raise RuntimeError(msg)
        self._path = path
        _atomic_write_text(
            path,
            json.dumps(
                {"collection_id": self.collection_id, "file_key": self.file_key}
            ),
        )


class FileSystemClient:
    FILE_INDEX_DIR = "__file_index__"

    def __init__(self, base_path: Path) -> None:
        self._base_path = base_path
        self._base_path.mkdir(exist_ok=True)
        (self._base_path / self.FILE_INDEX_DIR).mkdir(exist_ok=True)

    def _collection_metadata_path(self, collection_id: str) -> Path:
        return self._base_path / collection_id / ".collection.meta.json"

    def _file_index_entry(self, file_id: str) -> FileIndexEntry:
        return FileIndexEntry.load(self._base_path / self.FILE_INDEX_DIR / file_id)

    def _file_metadata_path(self, collection_id: str, file_key: str) -> Path:
        return self._base_path / collection_id / f"{_escape(file_key)}.file.meta.json"

    def _file_data_path(self, collection_id: str, file_key: str) -> Path:
        return self._base_path / collection_id / f"{_escape(file_key)}.file.data"

    def _document_path(self, collection_id: str, document_key: str) -> Path:
        return self._base_path / collection_id / f"{_escape(document_key)}.doc.json"

    def _document_path_from_id(self, document_id: str) -> Path:
        return self._base_path / (document_id + ".doc.json")

    def _ensure_collection_metadata(
        self, collection_id: str, collection_key: str
    ) -> FileSystemCollectionMetadata:
        metadata_path = self._collection_metadata_path(collection_id)
        if metadata_path.exists():
            return FileSystemCollectionMetadata.load(metadata_path)

        metadata = FileSystemCollectionMetadata(
            collection_id=collection_id, collection_key=collection_key, tags=[]
        )
        metadata.save(metadata_path)
        return metadata

    def collection_reference(
        self, collection_key: str, parent_collection_id: str | None = None
    ) -> CollectionIdentifier:
        collection_relpath = (
            Path(parent_collection_id) / collection_key
            if parent_collection_id is not None
            else Path(collection_key)
        )
        collection_id = str(collection_relpath)
        collection_path = self._base_path / collection_relpath
        collection_path.mkdir(parents=True, exist_ok=True)
        return CollectionIdentifier(id=collection_id, key=collection_key)

    def collection_delete(self, collection_id: str) -> None:
        col_path = self._base_path / collection_id
        if not col_path.exists() or not col_path.is_dir():
            raise ParentCollectionNotFoundError(collection_id=collection_id)

        for item in col_path.iterdir():
            if item.name == ".collection.meta.json":
                continue
            if (
                item.is_dir()
                or item.name.endswith(".doc.json")
                or item.name.endswith(".file.meta.json")
            ):
                msg = (
                    f"Collection '{collection_id}' is not empty and cannot be deleted."
                )
                raise ValueError(msg)

        metadata_path = self._collection_metadata_path(collection_id)
        metadata_path.unlink(missing_ok=True)
        col_path.rmdir()

    def collection_collections(
        self,
        collection_key: str,
        end_cursor: str,  # noqa: ARG002
        tag: Tag | None,
    ) -> tuple[list[CollectionIdentifier], bool, str]:
        col_path = self._base_path / collection_key
        if not col_path.exists():
            return [], False, ""

        collections: list[CollectionIdentifier] = []
        for item in col_path.iterdir():
            if item.is_dir():
                col_id = str(item.relative_to(self._base_path))

                if tag is not None:
                    metadata_path = self._collection_metadata_path(col_id)
                    if metadata_path.exists():
                        metadata = FileSystemCollectionMetadata.load(metadata_path)
                        if not metadata.tag_matches(tag):
                            continue
                    else:
                        continue

                collections.append(CollectionIdentifier(id=col_id, key=item.name))

        return sorted(collections, key=lambda c: c.id), False, ""

    def collection_documents(
        self,
        collection_id: str,
        end_cursor: str,  # noqa: ARG002
        tag: Tag | None,
        *,
        include_data: bool = False,
        include_tags: bool = False,
    ) -> tuple[list[CollectionDocumentIdentifier], bool, str]:
        col_path = self._base_path / collection_id
        if not col_path.exists():
            return [], False, ""

        results: list[CollectionDocumentIdentifier] = []
        for doc_path in col_path.iterdir():
            if not doc_path.name.endswith(".doc.json"):
                continue

            doc = FileSystemCollectionDocument.load(doc_path)

            if tag and not doc.tag_matches(tag):
                # skips files that do not match tag input, if it is given
                continue

            raw_stem = doc_path.name.removesuffix(".doc.json")
            doc_id = str(doc_path.relative_to(self._base_path).with_name(raw_stem))
            doc_key = _unescape(raw_stem)
            results.append(
                CollectionDocumentIdentifier(
                    id=doc_id,
                    key=doc_key,
                    data=dict_to_base64(doc.data) if include_data else None,
                    tags=(
                        {t.key: t.value for t in doc.tags} if include_tags else None
                    ),
                )
            )

        return sorted(results, key=lambda d: d.id if d else ""), False, ""

    def collection_files(
        self,
        collection_id: str,
        end_cursor: str,  # noqa: ARG002
        tag: Tag | None,
    ) -> tuple[list[CollectionFileIdentifier], bool, str]:
        col_path = self._base_path / collection_id
        if not col_path.exists():
            return [], False, ""

        files: list[CollectionFileIdentifier] = []
        for file_path in col_path.iterdir():
            if not file_path.name.endswith(".file.meta.json"):
                continue

            meta = FileSystemFileMetadata.load(file_path)

            if tag and not meta.tag_matches(tag):
                # skips files that do not match tag input, if it is given
                continue

            files.append(CollectionFileIdentifier(id=meta.file_id, key=meta.file_key))

        return files, False, ""

    def collection_tags(self, collection_id: str) -> list[Tag]:
        metadata_path = self._collection_metadata_path(collection_id)
        if not metadata_path.exists():
            return []

        metadata = FileSystemCollectionMetadata.load(metadata_path)
        return metadata.reference_tags()

    def collection_tag_add(self, collection_id: str, tag: Tag) -> None:
        collection_key = collection_id.split("/")[-1]
        metadata = self._ensure_collection_metadata(collection_id, collection_key)

        if not metadata.tag_matches(tag):
            metadata.tags.append(Tag(key=tag.key, value=tag.value))
            metadata.save(self._collection_metadata_path(collection_id))

    def collection_tag_delete(self, collection_id: str, tag_key: str) -> None:
        metadata_path = self._collection_metadata_path(collection_id)
        if not metadata_path.exists():
            return

        metadata = FileSystemCollectionMetadata.load(metadata_path)
        metadata.tags = [tag for tag in metadata.tags if tag.key != tag_key]
        metadata.save(metadata_path)

    def document_reference(
        self, collection_id: str, document_key: str
    ) -> CollectionDocumentIdentifier | None:
        path = self._document_path(collection_id, document_key)
        if not path.exists():
            return None

        doc_id = str(Path(collection_id) / _escape(document_key))
        return CollectionDocumentIdentifier(id=doc_id, key=document_key)

    def document_get(self, document_id: str) -> str | None:
        path = self._document_path_from_id(document_id)
        if not path.exists():
            return None

        doc = FileSystemCollectionDocument.load(path)
        return dict_to_base64(doc.data)

    def document_exists(self, document_id: str) -> bool:
        return self._document_path_from_id(document_id).exists()

    def document_tags(self, document_id: str) -> dict[str, str] | None:
        path = self._document_path_from_id(document_id)
        if not path.exists():
            return None

        doc = FileSystemCollectionDocument.load(path)
        return {tag.key: tag.value for tag in doc.tags}

    def document_set(
        self, collection_id: str, document_key: str, encoded_data: str
    ) -> None:
        path = self._document_path(collection_id, document_key)
        data = base64_to_dict(encoded_data)
        if path.exists():
            doc = FileSystemCollectionDocument.load(path)
            doc.data = data
        else:
            doc = FileSystemCollectionDocument(data, [])
        doc.save(path)

    def document_delete(self, document_id: str) -> None:
        doc_path = self._document_path_from_id(document_id)
        if not doc_path.exists():
            return

        doc_path.unlink()

    def document_tag_add(self, document_id: str, tag: Tag) -> None:
        doc_path = self._document_path_from_id(document_id)
        if not doc_path.exists():
            return

        doc = FileSystemCollectionDocument.load(doc_path)
        doc.tags.append(Tag(key=tag.key, value=tag.value))
        doc.save(doc_path)

    def document_tag_delete(self, document_id: str, tag_key: str) -> None:
        doc_path = self._document_path_from_id(document_id)
        if not doc_path.exists():
            return

        doc = FileSystemCollectionDocument.load(doc_path)
        doc.tags = [tag for tag in doc.tags if tag.key != tag_key]
        doc.save(doc_path)

    def file_reference(
        self, collection_id: str, file_key: str
    ) -> CollectionFileIdentifier | None:
        meta_path = self._file_metadata_path(collection_id, file_key)
        if meta_path.exists():
            meta = FileSystemFileMetadata.load(meta_path)
        else:
            file_id = _escape(collection_id + "_" + file_key)
            index_entry = FileIndexEntry(collection_id=collection_id, file_key=file_key)
            index_entry.save(self._base_path / self.FILE_INDEX_DIR / file_id)
            meta = FileSystemFileMetadata(file_id=file_id, file_key=file_key, tags=[])
            meta.save(self._file_metadata_path(collection_id, file_key))

        return CollectionFileIdentifier(id=meta.file_id, key=file_key)

    def file_tags(self, file_id: str) -> dict[str, str] | None:
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

    def file_delete(self, file_id: str) -> None:
        index_entry = self._file_index_entry(file_id)
        meta_path = self._file_metadata_path(
            index_entry.collection_id, index_entry.file_key
        )
        data_path = self._file_data_path(
            index_entry.collection_id, index_entry.file_key
        )

        if not meta_path.exists():
            index_entry.remove()
            return

        meta_path.unlink(missing_ok=True)
        data_path.unlink(missing_ok=True)
        index_entry.remove()

    def file_tag_add(self, file_id: str, tag: Tag) -> None:
        index_entry = self._file_index_entry(file_id)
        meta_path = self._file_metadata_path(
            index_entry.collection_id, index_entry.file_key
        )
        if not meta_path.exists():
            return

        meta = FileSystemFileMetadata.load(meta_path)
        if not meta.tag_matches(tag):
            meta.tags.append(Tag(key=tag.key, value=tag.value))
        meta.save(meta_path)

    def file_delete_tag(self, file_id: str, tag_key: str) -> None:
        index_entry = self._file_index_entry(file_id)
        meta_path = self._file_metadata_path(
            index_entry.collection_id, index_entry.file_key
        )
        if not meta_path.exists():
            return

        meta = FileSystemFileMetadata.load(meta_path)
        meta.tags = [tag for tag in meta.tags if tag.key != tag_key]
        meta.save(meta_path)

    def file_read_text(self, file_id: str) -> str:
        index_entry = self._file_index_entry(file_id)
        data_path = self._file_data_path(
            index_entry.collection_id, index_entry.file_key
        )
        return data_path.read_text()

    def file_read_bytes(self, file_id: str) -> bytes:
        index_entry = self._file_index_entry(file_id)
        data_path = self._file_data_path(
            index_entry.collection_id, index_entry.file_key
        )
        return data_path.read_bytes()

    def file_download_url(self, file_id: str) -> str | None:
        try:
            index_entry = self._file_index_entry(file_id)
        except FileNotFoundError:
            return None
        data_path = self._file_data_path(
            index_entry.collection_id, index_entry.file_key
        )
        if not data_path.exists():
            return None
        return data_path.resolve().as_uri()

    def file_read_range(
        self, file_id: str, start: int, end: int | None = None
    ) -> bytes:
        if start < 0:
            msg = f"Range start must be >= 0, got {start}"
            raise ValueError(msg)
        if end is not None and end < start:
            msg = f"Range end ({end}) must be >= start ({start})"
            raise ValueError(msg)

        index_entry = self._file_index_entry(file_id)
        data_path = self._file_data_path(
            index_entry.collection_id, index_entry.file_key
        )
        with data_path.open("rb") as f:
            f.seek(start)
            if end is None:
                return f.read()
            return f.read(end - start + 1)

    def file_save(self, file_id: str, data: bytes | str) -> None:
        index_entry = self._file_index_entry(file_id)
        data_path = self._file_data_path(
            index_entry.collection_id, index_entry.file_key
        )
        if isinstance(data, bytes):
            _atomic_write_bytes(data_path, data)
        else:
            _atomic_write_text(data_path, data)

    def file_open(self, file_id: str) -> BinaryIO:
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
    return quote(key, safe="._-")


def _unescape(encoded: str) -> str:
    return unquote(encoded)
