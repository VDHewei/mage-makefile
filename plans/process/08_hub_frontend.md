# Phase 8: Hub Frontend Integration / 第八阶段：Hub 前端集成

## Status / 状态: Not Started / 未开始

## Overview / 概述

Phase 8 implements the frontend/web layer for the Hub — a browsable web UI for discovering, sharing, and managing magefile snippets. This uses the Fiber HTTP framework (already installed) to serve a lightweight web interface.

第八阶段实现 Hub 的前端/Web 层 — 用于发现、分享和管理 magefile 片段的浏览式 Web UI。使用已安装的 Fiber HTTP 框架提供轻量级 Web 界面。

---

## Tasks / 任务

| Task | EN Description | CN Description | Status |
|------|---------------|---------------|--------|
| 8.1 | Create `pkg/hub/web/handler.go` — Fiber HTTP handlers | 创建 web 包：Fiber HTTP 处理器 | Pending |
| 8.2 | Create `pkg/hub/web/embed.go` — embed frontend assets | 创建 embed.go — 嵌入前端资源 | Pending |
| 8.3 | Implement snippet browser page | 实现片段浏览器页面 | Pending |
| 8.4 | Implement snippet detail page with code preview | 实现片段详情页，含代码预览 | Pending |
| 8.5 | Implement upload form with preview | 实现上传表单，含预览 | Pending |
| 8.6 | Implement login page | 实现登录页面 | Pending |
| 8.7 | Wire up hub serve command in CLI | 连接 CLI 的 hub serve 命令 | Pending |
| 8.8 | Add web integration tests | 添加 Web 集成测试 | Pending |

---

## Key Design Decisions / 关键设计决策

| # | EN Decision | CN Decision | Rationale |
|---|---|---|---|
| 1 | Fiber serves static + dynamic routes | Fiber 同时提供静态和动态路由 | Single server, no extra deps |
| 2 | Embedded SPA (Vue/React) via `//go:embed` | 嵌入单页应用 | Zero deploy overhead |
| 3 | API routes under `/api/`, UI under `/` | API 和 UI 路由分离 | Clean separation |

---

## Dependencies / 依赖

- Phase 7 (Hub API) — API service layer
- Fiber v2 (already in go.mod)
- Go embed (stdlib)
