from __future__ import annotations

import base64
import json
from datetime import datetime, timezone
from typing import TYPE_CHECKING
from unittest.mock import AsyncMock, MagicMock, Mock, patch

import pytest

from numerous._client.fs_client import FileSystemClient
from numerous._client.graphql.client import Client as GQLClient
from numerous._client.graphql_client import GraphQLClient
from numerous._utils.jsonbase64 import dict_to_base64
from numerous.collections.collection_reference import CollectionReference
from numerous.collections.exceptions import ParentCollectionNotFoundError
from numerous.tasks.context import _current_inputs, get_task_inputs
from numerous.tasks.controller import LocalTaskController, PlatformTaskController
from numerous.tasks.serialization import deserialize_task_inputs
from numerous.tasks.types import TaskInstanceState, TaskStatus, TaskWorkload


if TYPE_CHECKING:
    from pathlib import Path

EXPECTED_LOCAL_PROGRESS = 0.4


def test_get_task_inputs_reads_numerous_input_data_raw_json(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    token = _current_inputs.set(None)
    try:
        monkeypatch.setenv(
            "NUMEROUS_INPUT_DATA", json.dumps({"scenario": "base", "n": 42})
        )
        monkeypatch.delenv("TASK_DATA_INPUT", raising=False)

        assert get_task_inputs() == {"scenario": "base", "n": 42}
    finally:
        _current_inputs.reset(token)


def test_deserialize_task_inputs_supports_raw_json_and_base64() -> None:
    payload = {"run_id": "abc", "steps": [1, 2, 3]}
    raw_json = json.dumps(payload)
    b64_json = base64.b64encode(raw_json.encode("utf-8")).decode("utf-8")

    assert deserialize_task_inputs(raw_json) == payload
    assert deserialize_task_inputs(b64_json) == payload
    assert deserialize_task_inputs("plain-text") == {"_raw": "plain-text"}


def test_fs_client_hardening_nested_keys_collision_free_and_index_cleanup(
    tmp_path: Path,
) -> None:
    fs = FileSystemClient(tmp_path)
    col = CollectionReference(
        id=fs.collection_reference("project").id, key="project", _client=fs
    )

    # Nested keys with '/' do not collide with '__' or appear as child collections
    col.document("runs/1/status").set({"kind": "slash"})
    col.document("runs__1__status").set({"kind": "underscore"})

    assert list(col.collections()) == []
    docs = {d.key: d.get() for d in col.documents(include_data=True, include_tags=True)}
    assert docs == {
        "runs/1/status": {"kind": "slash"},
        "runs__1__status": {"kind": "underscore"},
    }

    # Files with '/' vs '__' do not collide and clean up index on delete
    f_slash = col.file("results/a.bin")
    f_under = col.file("results__a.bin")
    f_slash.save(b"0123456789abcdef")
    f_under.save(b"other-bytes")

    assert f_slash.read_range(4, 9) == b"456789"
    assert f_slash.read_range(12) == b"cdef"
    assert f_slash.presigned_url() is not None
    assert f_under.read_bytes() == b"other-bytes"

    index_file = tmp_path / FileSystemClient.FILE_INDEX_DIR / f_slash.id
    assert index_file.exists()
    f_slash.delete()
    assert not index_file.exists()
    assert not f_slash.exists

    # Empty collection delete rejects non-empty and succeeds once empty
    with pytest.raises(ValueError, match="not empty"):
        col.delete()

    col.document("runs/1/status").delete()
    col.document("runs__1__status").delete()
    f_under.delete()
    col.delete()
    assert not (tmp_path / "project").exists()

    with pytest.raises(ParentCollectionNotFoundError):
        col.delete()


def test_graphql_client_batch_documents_avoids_n_plus_1() -> None:
    gql = Mock(GQLClient)
    mock_response = MagicMock()
    gql.execute = AsyncMock(return_value=mock_response)
    gql.get_data = Mock(
        return_value={
            "collection": {
                "__typename": "Collection",
                "documents": {
                    "edges": [
                        {
                            "node": {
                                "id": "doc-1",
                                "key": "k1",
                                "data": dict_to_base64({"x": 1}),
                                "tags": [{"key": "env", "value": "prod"}],
                            }
                        }
                    ],
                    "pageInfo": {"hasNextPage": False, "endCursor": "doc-1"},
                },
            }
        }
    )

    client = GraphQLClient(gql, "org-1", "token")
    col = CollectionReference(id="col-1", key="root", _client=client)
    docs = list(col.documents(include_data=True, include_tags=True))

    assert len(docs) == 1
    assert docs[0].get() == {"x": 1}
    assert docs[0].tags == {"env": "prod"}
    gql.collection_document.assert_not_called()


def test_graphql_client_file_range_and_presigned_url() -> None:
    gql = Mock(GQLClient)
    file_resp = MagicMock()
    file_resp.collection_file.download_url = "https://s3.example/presigned"
    gql.collection_file = AsyncMock(return_value=file_resp)

    client = GraphQLClient(gql, "org-1", "token")
    assert client.file_download_url("file-1") == "https://s3.example/presigned"

    fake_http = MagicMock()
    fake_http.content = b"slice"
    with patch(
        "numerous._client.graphql_client.requests.get", return_value=fake_http
    ) as mock_get:
        chunk = client.file_read_range("file-1", 10, 14)

    assert chunk == b"slice"
    mock_get.assert_called_once()
    assert mock_get.call_args.kwargs["headers"] == {"Range": "bytes=10-14"}


def test_task_controller_progress_message_and_task_logs() -> None:
    state = TaskInstanceState(
        id="inst-1",
        task_id="sim",
        status=TaskStatus.RUNNING,
        progress=0.0,
        inputs={},
        workload=TaskWorkload.LOCAL,
        created_at=datetime.now(timezone.utc),
    )
    local_ctrl = LocalTaskController(state)
    local_ctrl.set_progress(EXPECTED_LOCAL_PROGRESS, message="Running EnergyPlus")
    assert state.progress == EXPECTED_LOCAL_PROGRESS
    assert state.progress_message == "Running EnergyPlus"

    gql = Mock(GQLClient)
    gql.task_instance_update_progress = AsyncMock()
    gql.execute = AsyncMock(return_value=MagicMock())
    gql.get_data = Mock(
        return_value={
            "taskInstance": {
                "workload": {
                    "logs": {
                        "edges": [
                            {"timestamp": "2026-01-01T12:00:01Z", "text": "step 1"},
                            {"timestamp": "2026-01-01T12:00:02Z", "text": "step 2"},
                        ],
                        "pageInfo": {
                            "endCursor": "2026-01-01T12:00:02Z",
                            "hasNextPage": False,
                        },
                    }
                }
            }
        }
    )

    client = GraphQLClient(gql, "org-1", "token")
    platform_ctrl = PlatformTaskController(state, client)
    platform_ctrl.set_progress(0.75, message="Post-processing")
    gql.task_instance_update_progress.assert_called_once_with(
        task_instance_id="inst-1",
        value=75.0,
        message="Post-processing",
        headers={"Authorization": "Bearer token"},
    )

    page = client.task_instance_logs(
        "inst-1", last=50, after="2026-01-01T12:00:00Z"
    )
    assert [e.text for e in page.entries] == ["step 1", "step 2"]
    assert page.end_cursor == "2026-01-01T12:00:02Z"
