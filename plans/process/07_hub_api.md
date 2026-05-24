# Phase 7: Hub API Service / 第七阶段：Hub API 服务

## Status / 状态: Not Started / 未开始

## Overview / 概述

Phase 7 implements the Hub API service layer (`magego.hub.io`) for sharing, discovering, and versioning magefile snippets. This includes:
- REST API client with authentication
- Push/pull/search/upload/download endpoints
- Snippet metadata, tags, and versioning
- Code review and approval workflow

第七阶段实现 Hub API 服务层（magego.hub.io），用于分享、发现和版本管理 magefile 代码片段。包括：
- 带认证的 REST API 客户端
- push/pull/search/upload/download 端点
- 片段元数据、标签和版本管理
- 代码审核和审批流程

---

## Tasks / 任务

| Task | EN Description | CN Description | Status |
|------|---------------|---------------|--------|
| 7.1 | Create `pkg/hub/client.go` with `Client` struct and auth middleware | 创建 hub 包：Client 结构和认证中间件 | Pending |
| 7.2 | Implement `Push` — upload a magefile.go snippet | 实现 Push — 上传 magefile.go 片段 | Pending |
| 7.3 | Implement `Pull` — download a magefile.go snippet | 实现 Pull — 下载 magefile.go 片段 | Pending |
| 7.4 | Implement `Search` — search snippets by query and tags | 实现 Search — 按查询和标签搜索片段 | Pending |
| 7.5 | Implement `List` — list all snippets with pagination | 实现 List — 分页列出所有片段 | Pending |
| 7.6 | Implement `Version` — get snippet versions | 实现 Version — 获取片段版本 | Pending |
| 7.7 | Implement `Login` — authenticate with username/password or API key | 实现 Login — 用户名/密码或 API Key 认证 | Pending |
| 7.8 | Wire up CLI commands: `hub push/pull/search/list/version` | 连接 CLI 命令：hub push/pull/search/list/version | Pending |
| 7.9 | Add Hub client unit tests | 添加 Hub 客户端单元测试 | Pending |
| 7.10 | Add CLI integration tests | 添加 CLI 集成测试 | Pending |

---

## Key Design Decisions / 关键设计决策

| # | EN Decision | CN Decision | Rationale |
|---|---|---|---|
| 1 | Hub API is a NEW package `pkg/hub/` | Hub API 是新的独立包 | Keeps hub logic separate from converter/compiler |
| 2 | Uses `net/http` stdlib + `encoding/json` | 使用 stdlib HTTP + JSON | Zero external deps for the API client |
| 3 | Auth via Bearer token in `Authorization` header | 通过 Bearer token 认证 | Standard, simple, compatible with most APIs |
| 4 | Snippet format: JSON with code, metadata, tags | 片段格式：JSON 含代码、元数据、标签 | Portable, versionable, human-readable |
| 5 | Pagination via `?page=N&limit=M` query params | 分页通过查询参数 | Standard REST pattern |

---

## Test Strategy / 测试策略

**Unit tests:** HTTP client with httptest server, mock auth
**CLI tests:** `hub search`, `hub list`, `hub version` (dry-run mode)
**Integration:** End-to-end with mock Hub server

---

## Dependencies / 依赖

- Phase 5 (Config) — hub server URL and timeout config
- Phase 3 (CLI) — cobra patterns for hub subcommands
- Existing `net/http`, `encoding/json` stdlib
