"""Generate safe client-side scripts for the Numerous Agent UI bridge."""

from __future__ import annotations

import json
from typing import TYPE_CHECKING, Any


if TYPE_CHECKING:
    from collections.abc import Mapping


def encode_agent_context_json(context: Mapping[str, Any]) -> str:
    """Serialize context data as JSON safe for embedding inside a <script> tag."""
    raw = json.dumps(dict(context), default=str)
    # Prevent closing </script> tags or HTML comments from breaking out of script text.
    return (
        raw.replace("</", "<\\/")
        .replace("<!--", "<\\!--")
        .replace("\u2028", "\\u2028")
        .replace("\u2029", "\\u2029")
    )


def agent_context_script(
    context: Mapping[str, Any] | None = None,
    *,
    actions: Mapping[str, str] | None = None,
) -> str:
    """
    Build a ``<script>`` tag publishing app context to the Numerous Agent UI bridge.

    The Numerous proxy injects ``/__numerous/agent-bridge.js`` into web app HTML
    responses. When a user clicks **Share with Agent**, the bridge reads
    ``window.__numerousAgentContext`` and exposes any registered actions to the
    user's personal Numerous MCP server via ``ui_session_inspect`` and
    ``ui_session_act``.
    """
    parts: list[str] = ["(function () {"]

    if context is not None:
        encoded = encode_agent_context_json(context)
        parts.append(f"  var ctx = {encoded};")
        parts.append("  window.__numerousAgentContext = function () { return ctx; };")
        parts.append(
            "  if (window.__numerousAgent && window.__numerousAgent.setContext) {"
            " window.__numerousAgent.setContext(function () { return ctx; }); }"
        )

    if actions:
        for name, js_body in actions.items():
            encoded_name = json.dumps(str(name)).replace("</", "<\\/")
            safe_body = str(js_body).replace("</", "<\\/")
            parts.append(
                "  if (window.__numerousAgent &&"
                " window.__numerousAgent.registerAction) {"
                f" window.__numerousAgent.registerAction({encoded_name},"
                f" {safe_body}); }}"
            )

    parts.append("})();")
    return "<script>\n" + "\n".join(parts) + "\n</script>"
