---
trigger: always_on
description: Consult the graphify knowledge graph at graphify-out/ for codebase and architecture questions.
---

## graphify (Primary Reference)

This project has a graphify knowledge graph at graphify-out/. It is the single source of truth for repository architecture, symbol maps, and dependencies across all chats.

Rules:
- MANDATORY FIRST STEP: For any codebase or architecture question, query `graphify-out/graph.json` first using `graphify query "<question>"` or MCP `query_graph`.
- Use `graphify path "<A>" "<B>"` / `shortest_path` for relationships and `graphify explain "<concept>"` / `get_node` for focused concepts.
- If graphify-out/wiki/index.md exists, navigate it instead of reading raw files.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- Always run `graphify update .` after modifying code files to keep the graph 100% current.
