# PostgreSQL RAG tool

Snapshot v1.6.0 composes Agentx v1.4.0, retrieval v1.2.0 and Servex v1.5.0
with database/postgres, vector/pgvector and search/postgres v1.0.0.
All dependencies are public versions; no sibling source checkout is required.

The thin entry is cmd/postgres-rag. Configuration lives in infrastructure/config;
services/knowledge/app.go owns assembly, explicit migrations and ingestion.
The retrieval adapter imports only Servex contracts, while this application
selects PostgreSQL backends and owns its SQL pool.

1. Create only a dedicated ax_demo_ schema and two namespaced indexes.
2. Encode complete rag.Chunk payloads with VectorRecord/TextDocument and write
   both indexes in one caller-owned transaction.
3. Query vector and text independently, fuse ranks with rag.Hybrid, and invoke
   rag.Tool.Call to return content and typed citations.
4. Enforce fixed exact source and string metadata filters. A deliberately
   inserted private source must not appear in tool output.

DemoEmbedder is deterministic fixture data. It does not measure semantic quality
or call a paid model. Replace document and query embeddings together for production.
Vector scores require cosine distance; text scores stay engine-specific.
The tool is read-only. Source prefix filters and multiple exact sources are unsupported.

Use an owned PostgreSQL 17 instance with vector installed:

```sh
export AGENTX_DEMO_POSTGRES_DSN='postgres://postgres@127.0.0.1:15439/agentx?sslmode=disable'
export AGENTX_DEMO_POSTGRES_SCHEMA=ax_demo_knowledge
export AGENTX_DEMO_SEARCH_ENGINE=native
go run ./cmd/postgres-rag
```

The command prints vector_hits/text_hits, content and citations. Repeated runs
upsert the same IDs. It leaves its demo schema for inspection; remove only that
schema in your owned fixture. It never installs extensions.

```sh
export AGENTX_EXAMPLES_IT_POSTGRES_DSN="$AGENTX_DEMO_POSTGRES_DSN"
task test INTEGRATION=true
```

The gate rejects missing configuration and skipped, failed, absent or unfinished
tests. It runs the real tool twice, verifies source/citation isolation and cleans
its generated schema. Ordinary task check does not require a database.
Hosted CI provisions its own pgvector service.

中文：本例在应用层组合 PostgreSQL 向量与全文实现，通过 AgentX RAG 工具返回引用；
数据写入、迁移和事务由应用负责。固定向量只用于确定性验收，不调用付费模型，也不代表
真实语义检索效果。必须使用独立 ax_demo_ schema，勿指向生产数据库。
