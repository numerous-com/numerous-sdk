import json
import os
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any, Optional

from numerous.generated.graphql.fragments import (
    CollectionDocumentReference,
    CollectionDocumentReferenceTags,
    CollectionFileReference,
    CollectionFileReferenceTags,
    CollectionReference,
)
from numerous.generated.graphql.input_types import TagInput
from numerous.jsonbase64 import base64_to_dict, dict_to_base64


@dataclass
class FileSystemCollectionDocumentTag:
    key: str
    value: str

    @staticmethod
    def load(tag: dict[str, Any]) -> "FileSystemCollectionDocumentTag":
        key = tag.get("key")
        if not isinstance(key, str):
            tname = type(key).__name__
            msg = f"FileSystemCollectionDocumentTag key must be str, found {tname}"
            raise TypeError(msg)

        value = tag.get("value")
        if not isinstance(value, str):
            tname = type(value).__name__
            msg = f"FileSystemCollectionDocumentTag value must be str, found {tname}"
            raise TypeError(msg)

        return FileSystemCollectionDocumentTag(key=key, value=value)

    def to_reference_tag(self) -> CollectionDocumentReferenceTags:
        return CollectionDocumentReferenceTags(
            key=self.key,
            value=self.value,
        )


@dataclass
class FileSystemCollectionFileTag:
    key: str
    value: str

    @staticmethod
    def load(tag: dict[str, Any]) -> "FileSystemCollectionFileTag":
        key = tag.get("key")
        if not isinstance(key, str):
            tname = type(key).__name__
            msg = f"FileSystemCollectionFileTag key must be str, found {tname}"
            raise TypeError(msg)

        value = tag.get("value")
        if not isinstance(value, str):
            tname = type(value).__name__
            msg = f"FileSystemCollectionFileTag value must be str, found {tname}"
            raise TypeError(msg)

        return FileSystemCollectionFileTag(key=key, value=value)

    def to_reference_tag(self) -> CollectionFileReferenceTags:
        return CollectionFileReferenceTags(
            key=self.key,
            value=self.value,
        )


@dataclass
class FileSystemCollectionFile:
    path: Path
    tags: list[FileSystemCollectionFileTag]

    def save(self, path: Path) -> None:
        def convert_to_serializable(obj: Path) -> str:
            if isinstance(obj, Path):
                return str(obj)
            return None

        with path.open("w") as f:
            json.dump(asdict(self), f, default=convert_to_serializable)

    @staticmethod
    def load(file_path: Path) -> "FileSystemCollectionFile":
        with file_path.open("r") as f:
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

        path = file_content.get("path", {})
        if not isinstance(path, str):
            tname = type(path).__name__
            msg = f"FileSystemCollection data must be a dict, found {tname}"
            raise TypeError(msg)
        path = Path(path)

        return FileSystemCollectionFile(
            path=path, tags=[FileSystemCollectionFileTag.load(tag) for tag in tags]
        )

    def reference_tags(self) -> list[CollectionFileReferenceTags]:
        return [
            CollectionFileReferenceTags(
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
class FileSystemCollectionDocument:
    data: dict[str, Any]
    tags: list[FileSystemCollectionDocumentTag]

    def save(self, path: Path) -> None:
        with path.open("w") as f:
            json.dump(asdict(self), f)

    @staticmethod
    def load(path: Path) -> "FileSystemCollectionDocument":
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
            data=data, tags=[FileSystemCollectionDocumentTag.load(tag) for tag in tags]
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


class FileSystemClient:
    def __init__(self, base_path: Path) -> None:
        self._base_path = base_path
        self._base_path.mkdir(exist_ok=True)

    def get_collection_reference(
        self, collection_key: str, parent_collection_id: Optional[str] = None
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
    ) -> Optional[CollectionDocumentReference]:
        path = self._base_path / collection_id / f"{document_key}.json"
        if not path.exists():
            return None

        doc = FileSystemCollectionDocument.load(path)

        doc_id = str(Path(collection_id) / document_key)
        return CollectionDocumentReference(
            id=doc_id,
            key=document_key,
            data=dict_to_base64(doc.data),
            tags=[tag.to_reference_tag() for tag in doc.tags],
        )

    def set_collection_document(
        self, collection_id: str, document_key: str, encoded_data: str
    ) -> Optional[CollectionDocumentReference]:
        doc_path = self._base_path / collection_id / f"{document_key}.json"
        data = base64_to_dict(encoded_data)
        if doc_path.exists():
            doc = FileSystemCollectionDocument.load(doc_path)
            doc.data = data
        else:
            doc = FileSystemCollectionDocument(data, [])
        doc.save(doc_path)

        doc_id = str(Path(collection_id) / document_key)
        return CollectionDocumentReference(
            id=doc_id,
            key=document_key,
            data=encoded_data,
            tags=[],
        )

    def delete_collection_document(
        self, document_id: str
    ) -> Optional[CollectionDocumentReference]:
        doc_path = self._base_path / (document_id + ".json")
        if not doc_path.exists():
            return None

        doc = FileSystemCollectionDocument.load(doc_path)

        doc_path.unlink()

        return CollectionDocumentReference(
            id=document_id,
            key=doc_path.stem,
            data=dict_to_base64(doc.data),
            tags=doc.reference_tags(),
        )

    def add_collection_document_tag(
        self, document_id: str, tag: TagInput
    ) -> Optional[CollectionDocumentReference]:
        doc_path = self._base_path / (document_id + ".json")
        if not doc_path.exists():
            return None

        doc = FileSystemCollectionDocument.load(doc_path)
        doc.tags.append(FileSystemCollectionDocumentTag(key=tag.key, value=tag.value))
        doc.save(doc_path)

        return CollectionDocumentReference(
            id=document_id,
            key=doc_path.stem,
            data=dict_to_base64(doc.data),
            tags=doc.reference_tags(),
        )

    def delete_collection_document_tag(
        self, document_id: str, tag_key: str
    ) -> Optional[CollectionDocumentReference]:
        doc_path = self._base_path / (document_id + ".json")
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
        tag_input: Optional[TagInput],
    ) -> tuple[Optional[list[Optional[CollectionDocumentReference]]], bool, str]:
        col_path = self._base_path / collection_id
        if not col_path.exists():
            return [], False, ""

        documents: list[Optional[CollectionDocumentReference]] = []
        for doc_path in col_path.iterdir():
            if doc_path.suffix != ".json":
                continue

            doc = FileSystemCollectionDocument.load(doc_path)

            if tag_input and not doc.tag_matches(tag_input):
                # skips files that do not match tag input, if it is given
                continue

            doc_id = str(doc_path.relative_to(self._base_path).with_suffix(""))
            documents.append(
                CollectionDocumentReference(
                    id=doc_id,
                    key=doc_path.stem,
                    data=dict_to_base64(doc.data),
                    tags=doc.reference_tags(),
                )
            )

        return sorted(documents, key=lambda d: d.id if d else ""), False, ""

    def get_collection_file(
        self, collection_id: str, file_key: str
    ) -> Optional[CollectionFileReference]:
        path = self._base_path / collection_id / f"{file_key}.json"
        if not path.exists():
            return None

        file = FileSystemCollectionFile.load(path)

        file_id = str(Path(collection_id) / file_key)
        return CollectionFileReference(
            id=file_id,
            key=file_key,
            downloadURL=os.fspath(file.path),
            uploadURL=os.fspath(file.path),
            tags=[tag.to_reference_tag() for tag in file.tags],
        )

    def delete_collection_file(self, file_id: str) -> Optional[CollectionFileReference]:
        file_path = self._base_path / (file_id + ".json")
        if not file_path.exists():
            return None

        file = FileSystemCollectionFile.load(file_path)

        file_path.unlink()
        file.path.unlink()

        return CollectionFileReference(
            id=file_id,
            key=file_path.stem,
            downloadURL=os.fspath(file.path),
            uploadURL=os.fspath(file.path),
            tags=file.reference_tags(),
        )

    def get_collection_files(
        self,
        collection_key: str,
        end_cursor: str,  # noqa: ARG002
        tag_input: Optional[TagInput],
    ) -> tuple[Optional[list[Optional[CollectionFileReference]]], bool, str]:
        col_path = self._base_path / collection_key
        if not col_path.exists():
            return [], False, ""

        files: list[Optional[CollectionFileReference]] = []
        for file_path in col_path.iterdir():
            if file_path.suffix != ".json":
                continue

            file = FileSystemCollectionFile.load(file_path)

            if tag_input and not file.tag_matches(tag_input):
                # skips files that do not match tag input, if it is given
                continue

            file_id = str(file_path.relative_to(self._base_path).with_suffix(""))
            files.append(
                CollectionFileReference(
                    id=file_id,
                    key=file_path.stem,
                    downloadURL=os.fspath(file.path),
                    uploadURL=os.fspath(file.path),
                    tags=file.reference_tags(),
                )
            )

        return files, False, ""

    def add_collection_file_tag(
        self, file_id: str, tag: TagInput
    ) -> Optional[CollectionFileReference]:

        file_path = self._base_path / (file_id + ".json")
        if not file_path.exists():
            return None

        file = FileSystemCollectionFile.load(file_path)
        file.tags.append(FileSystemCollectionFileTag(key=tag.key, value=tag.value))
        file.save(file_path)

        return CollectionFileReference(
            id=file_id,
            key=file_path.stem,
            downloadURL=os.fspath(file.path),
            uploadURL=os.fspath(file.path),
            tags=file.reference_tags(),
        )

    def delete_collection_file_tag(
        self, file_id: str, tag_key: str
    ) -> Optional[CollectionFileReference]:
        file_path = self._base_path / (file_id + ".json")
        if not file_path.exists():
            return None

        file = FileSystemCollectionFile.load(file_path)
        file.tags = [tag for tag in file.tags if tag.key != tag_key]
        file.save(file_path)

        return CollectionFileReference(
            id=file_id,
            key=file_path.stem,
            downloadURL=os.fspath(file.path),
            uploadURL=os.fspath(file.path),
            tags=file.reference_tags(),
        )

    def get_collection_collections(
        self,
        collection_key: str,
        end_cursor: str,  # noqa: ARG002
    ) -> tuple[Optional[list[CollectionReference]], bool, str]:
        col_path = self._base_path / collection_key
        if not col_path.exists():
            return [], False, ""

        collections: list[CollectionReference] = []
        for item in col_path.iterdir():
            if item.is_dir():
                col_id = str(item.relative_to(self._base_path))
                collections.append(CollectionReference(id=col_id, key=item.name))

        return sorted(collections, key=lambda c: c.id), False, ""
