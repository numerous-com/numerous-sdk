from numerous.agent import agent_context_script, encode_agent_context_json


def test_encode_agent_context_json_escapes_script_breakout() -> None:
    encoded = encode_agent_context_json(
        {"title": "</script><script>alert(1)</script>", "note": "<!--comment-->"}
    )

    assert "</script>" not in encoded
    assert "<!--" not in encoded
    assert r"<\/script>" in encoded


def test_agent_context_script_includes_context_and_actions() -> None:
    html = agent_context_script(
        {"selectedQuarter": "2026-Q3", "rows": 42},
        actions={"selectQuarter": "function(args) { return args.quarter; }"},
    )

    assert html.startswith("<script>")
    assert html.endswith("</script>")
    assert "window.__numerousAgentContext" in html
    assert '"selectedQuarter": "2026-Q3"' in html
    assert "window.__numerousAgent.registerAction" in html
    assert '"selectQuarter"' in html
