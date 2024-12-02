from __future__ import annotations

import json
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any, BinaryIO

from numerous._utils.jsonbase64 import base64_to_dict, dict_to_base64
from numerous.collections._client import (
    CollectionDocumentIdentifier,
    CollectionFileIdentifier,
    CollectionIdentifier,
    Tag,
)


def _parse_tag(tag: dict[str, Any]) -> Tag:
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

    return Tag(key=key, value=value)


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

    def collection_documents(
        self,
        collection_id: str,
        end_cursor: str,  # noqa: ARG002
        tag_input: Tag | None,
    ) -> tuple[list[CollectionDocumentIdentifier], bool, str]:
        col_path = self._base_path / collection_id
        if not col_path.exists():
            return [], False, ""

        results: list[CollectionDocumentIdentifier] = []
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
            doc_key = doc_path.name.removesuffix(".doc.json")
            results.append(CollectionDocumentIdentifier(id=doc_id, key=doc_key))

        return sorted(results, key=lambda d: d.id if d else ""), False, ""

    def collection_files(
        self,
        collection_id: str,
        end_cursor: str,  # noqa: ARG002
        tag_input: Tag | None,
    ) -> tuple[list[CollectionFileIdentifier], bool, str]:
        col_path = self._base_path / collection_id
        if not col_path.exists():
            return [], False, ""

        files: list[CollectionFileIdentifier] = []
        for file_path in col_path.iterdir():
            if not file_path.name.endswith(".file.meta.json"):
                continue

            meta = FileSystemFileMetadata.load(file_path)

            if tag_input and not meta.tag_matches(tag_input):
                # skips files that do not match tag input, if it is given
                continue

            files.append(CollectionFileIdentifier(id=meta.file_id, key=meta.file_key))

        return files, False, ""

    def collection_collections(
        self,
        collection_key: str,
        end_cursor: str,  # noqa: ARG002
    ) -> tuple[list[CollectionIdentifier], bool, str]:
        col_path = self._base_path / collection_key
        if not col_path.exists():
            return [], False, ""

        collections: list[CollectionIdentifier] = []
        for item in col_path.iterdir():
            if item.is_dir():
                col_id = str(item.relative_to(self._base_path))
                collections.append(CollectionIdentifier(id=col_id, key=item.name))

        return sorted(collections, key=lambda c: c.id), False, ""

    def document_reference(
        self, collection_id: str, document_key: str
    ) -> CollectionDocumentIdentifier | None:
        path = self._document_path(collection_id, document_key)
        if not path.exists():
            return None

        doc_id = str(Path(collection_id) / document_key)
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
            return

        meta_path.unlink()
        data_path.unlink()

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

    def file_save(self, file_id: str, data: bytes | str) -> None:
        index_entry = self._file_index_entry(file_id)
        data_path = self._file_data_path(
            index_entry.collection_id, index_entry.file_key
        )
        if isinstance(data, bytes):
            data_path.write_bytes(data)
        else:
            data_path.write_text(data)

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
    return key.replace("/", "__")
