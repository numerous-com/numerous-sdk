# MCP & Live Agent UI Integration

Numerous apps can expose a [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server over Streamable HTTP and optionally publish structured context to the Numerous live Agent UI bridge.

## 1. Enable `[mcp]` in `numerous.toml`

```toml
name = "Sales Analytics"
description = "Sales dashboard with MCP tools"
port = 8501

[docker]
dockerfile = "Dockerfile"
context = "."

[mcp]
enabled = true
path = "/mcp" # optional, defaults to "/mcp"
```

When deployed with `numerous deploy`, the CLI prints both your app URL and your app's MCP URL (`https://<app-slug>.api.numerous.com/mcp`).

## 2. Read Verified User, Organization & UI Session Identity in MCP Tools

On every MCP (or HTTP API) request, the Numerous proxy strips any client-supplied `X-Numerous-*` headers and `numerous_user_info` cookies and injects verified identity headers derived from the caller's token.

Use `numerous.mcp` to read that verified identity inside your MCP tools or FastAPI endpoints:

```python
from mcp.server.fastmcp import Context, FastMCP
from numerous.mcp import caller_from_fastmcp_context, ui_sessions

mcp = FastMCP("sales-analytics", stateless_http=True, host="0.0.0.0", port=8501)


@mcp.tool()
def quarterly_revenue(quarter: str, ctx: Context) -> dict:
    """Return quarterly revenue scoped to the verified caller."""
    caller = caller_from_fastmcp_context(ctx)
    live_ui = ui_sessions.get(caller.ui_session_id)

    return {
        "quarter": quarter,
        "auth_type": caller.auth_type,
        "user_id": caller.user.id if caller.user else None,
        "organization_id": caller.organization_id,
        "live_ui_state": live_ui,
    }
```

- When the caller authenticates with a **Personal Access Token** (`num_...`) or **OIDC user token**, `caller.is_user` is `True`, `caller.user` is a `numerous.session.User` (so `caller.user.collection` works identically to web sessions), and `caller.ui_session_id` is set when linked to a verified active browser tab.
- When the caller authenticates with an **App API key** (`app_api_...`), `caller.is_api_key` is `True`, `caller.user` is `None`, and `caller.organization_id`, `caller.api_key_id`, and `caller.api_key_name` are populated.

## 3. Publish Live UI Context for Co-Browsing Agents

When a user clicks **Share with Agent** in the Numerous web shell, the proxy's UI bridge gives the user's personal Numerous MCP server access to the live DOM, highlights, and clicks with zero app code.

Apps can optionally enrich that fused snapshot with structured domain context (selected scenario, active filters, chart summary) using `numerous.agent`:

```python
import streamlit as st
from numerous.frameworks.streamlit import get_session, publish_agent_context
from numerous.mcp import ui_sessions

session = get_session()
selected_quarter = st.selectbox("Quarter", ["2026-Q1", "2026-Q2", "2026-Q3"])

# Publish structured context to the browser UI bridge (visible via ui_session_inspect)
publish_agent_context({"selectedQuarter": selected_quarter})

# Optionally store state keyed by the verified UI session ID for backend MCP tools
if session.ui_session_id:
    ui_sessions.update(session.ui_session_id, selected_quarter=selected_quarter)
```

For FastAPI, Dash, Panel, or custom HTML apps, embed `numerous.agent.agent_context_script(context)` in your page HTML, and use `numerous.frameworks.fastapi.get_caller(request)` to read verified caller identity in FastAPI routes.
