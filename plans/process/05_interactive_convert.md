# Phase 5: Interactive Conversion & Config-Driven Conversion / 第五阶段：交互式转换与配置驱动转换

## Status / 状态: Completed / 已完成

## Overview / 概述

> **注意：** 本阶段所有代码注释使用中文，用户交互提示使用中英双语，与 Phase 4 风格一致。

Phase 5 implements two tightly coupled features:
1. **Interactive Conversion** (`makego convert --interactive`) — a full TUI workflow for selecting targets, previewing code, and accepting/rejecting individual conversions
2. **Config-Driven Conversion** — a `[convert]` section in `.mage_makefile.toml` for target filtering, command mapping overrides, output customization, and alias configuration

第五阶段实现两个紧密耦合的功能：1）交互式转换 — 完整的终端交互工作流，支持选择目标、预览代码、接受/拒绝单个转换；2）配置驱动转换 — 在配置文件中新增 `[convert]` 配置节，支持目标过滤、命令映射覆盖、输出定制和别名配置。

---

## Tasks / 任务

| Task | EN Description | CN Description | Status |
|------|---------------|---------------|--------|
| 5.1 | Add `ConvertConfig` struct to `pkg/config/config.go` with all new fields | 在 config.go 中添加 ConvertConfig 结构体及其所有字段 | Pending |
| 5.2 | Add `[convert]` section to `internal/config/base.toml` with defaults | 在 base.toml 中添加 [convert] 配置节及默认值 | Pending |
| 5.3 | Add `ConvertConfig` unit tests in config package | 为 ConvertConfig 添加单元测试 | Pending |
| 5.4 | Create `pkg/converter/interactive/` package: `Categorizer` — target selection & categorization | 创建 interactive 包：Categorizer — 目标选择和分类 | Pending |
| 5.5 | Create `pkg/converter/interactive/previewer.go`: `CodePreview` — preview generated Go code | 创建 previewer.go：CodePreview — 预览生成的 Go 代码 | Pending |
| 5.6 | Create `pkg/converter/interactive/engine.go`: `InteractiveEngine` — full TUI orchestration | 创建 engine.go：InteractiveEngine — 完整 TUI 编排 | Pending |
| 5.7 | Enhance `transformer.go`: add `TransformWithFilter()` and `NewTransformerWithConfig()` | 增强 transformer.go：添加目标过滤功能和配置感知构造器 | Pending |
| 5.8 | Enhance `generator.go`: add `GeneratorOptions` struct, `NewGeneratorWithConfig()`, export `ToGoIdent` | 增强 generator.go：添加 GeneratorOptions 和配置感知构造器 | Pending |
| 5.9 | Enhance `cmd/makego/main.go`: rewrite `runConvert()` with interactive mode + config integration | 重写 runConvert()：集成交互模式和配置驱动的转换流程 | Pending |
| 5.10 | Add `--list-targets` flag to convert command | 添加 --list-targets 标志 | Pending |
| 5.11 | Add interactive package unit tests | 添加 interactive 包单元测试 | Pending |
| 5.12 | Add CLI integration tests | 添加 CLI 集成测试 | Pending |
| 5.13 | Full test suite verification | 全量测试验证 | Pending |

---

## Execution Order / 执行顺序

| Step | EN Action | CN Action | Verify |
|------|-----------|-----------|--------|
| 1 | 5.1: Edit config.go | 添加 ConvertConfig 结构体 | go build ./pkg/config/ |
| 2 | 5.2: Edit base.toml | 添加 [convert] 配置节 | TOML 格式验证 |
| 3 | 5.3: Test config | 配置测试 | go test ./pkg/config/ |
| 4 | 5.4: Create interactive.go | 创建目标分类器 | go vet ./pkg/converter/interactive/ |
| 5 | 5.5: Create previewer.go | 创建代码预览 | Build check |
| 6 | 5.6: Create engine.go | 创建交互引擎 | Build check |
| 7 | 5.7: Enhance transformer.go | 添加过滤函数 | go test ./pkg/converter/transformer/ |
| 8 | 5.8: Enhance generator.go | 添加配置感知生成 | go test ./pkg/converter/generator/ |
| 9 | 5.9-5.10: Enhance main.go | 重写 runConvert() | go build ./cmd/makego/ |
| 10 | 5.11: Create interactive_test.go | 创建单元测试 | All tests pass |
| 11 | 5.12: Enhance main_test.go | 创建 CLI 测试 | All tests pass |
| 12 | 5.13: Full test suite | 全量测试 | go test ./... |

---

## Key Design Decisions / 关键设计决策

| # | EN Decision | CN Decision | Rationale |
|---|-------------|-------------|-----------|
| 1 | Interactive engine is a NEW package `pkg/converter/interactive/` | 交互引擎是新的独立包 | Keeps CLI entry thin; interactive logic is testable in isolation |
| 2 | `Categorizer` uses heuristic keyword matching | 分类器使用启发式关键词匹配 | Deterministic, zero-dependency, fast |
| 3 | `GeneratorOptions` is a standalone struct | GeneratorOptions 是独立结构体 | Can be shared between previewer and batch generator |
| 4 | `TransformWithFilter()` filters AFTER full transform | 在完整转换后过滤 | Maintains all IR data for preview/dependency analysis |
| 5 | Config defaults: `AddComments: true`, `AddOriginal: false` | 配置默认值：注释开，原始命令关 | Makefile comments useful; raw shell commands noisy |
| 6 | `--list-targets` is a separate flag | --list-targets 是独立标志 | Useful for shell scripting and CI pipe inspection |
| 7 | Bilingual prompts throughout interactive mode | 交互模式全程双语提示 | Consistent with Phase 4 design |
| 8 | Export `ToGoIdent` from private `toGoIdent` | 将 toGoIdent 导出为 ToGoIdent | Needed by interactive previewer |
| 9 | `bufio.Reader` used in engine (not `fmt.Scanln`) | engine 中使用 bufio.Reader | Better handling of empty input and spaces |
| 10 | CLI flags take precedence over config values | CLI 标志优先级高于配置值 | Consistent with Phase 1 design |
| 11 | All code comments use Chinese | 所有代码注释使用中文 | User preference |

---

## Test Strategy / 测试策略

**interactive_test.go (NEW):** ~13 unit tests (categorization, preview, accept/reject, engine)
**config_test.go additions:** 3 tests (defaults, TOML round-trip, ApplyConfig)
**transformer_test.go additions:** 4 tests (TransformWithFilter variants)
**generator_test.go additions:** 5 tests (GeneratorOptions, comments, original, minimal, ToGoIdent)
**main_test.go additions:** 8 CLI tests (list-targets, interactive accept/reject, config filter/exclude/aliases, parse error, no targets)

### Progress Record / 进度记录

| Date | Task | Result |
|------|------|--------|
| 2026-05-10 | 5.1 config.go + ConvertConfig | Done - build passes |
| 2026-05-10 | 5.2 base.toml + [convert] section | Done - all config tests pass |
| 2026-05-10 | 5.4-5.6 interactive package | Done - build passes |
| 2026-05-10 | 5.8 generator enhancements | Done - all 12 generator tests pass |
| 2026-05-10 | 5.9-5.10 main.go rewrite + --list-targets | Done - all 10 CLI tests pass |
| 2026-05-10 | 5.11 interactive_test.go | Done - 14 tests pass |
| 2026-05-10 | 5.13 Full test suite | **ALL 9 packages PASS** |

## Final Test Results / 最终测试结果

```
ok  github.com/VDHewei/mage-makefile/cmd/makego         10.783s
ok  github.com/VDHewei/mage-makefile/pkg/compiler        1.048s
ok  github.com/VDHewei/mage-makefile/pkg/config          0.046s
ok  github.com/VDHewei/mage-makefile/pkg/converter/generator    0.043s
ok  github.com/VDHewei/mage-makefile/pkg/converter/interactive  0.044s
ok  github.com/VDHewei/mage-makefile/pkg/converter/parser       0.028s
ok  github.com/VDHewei/mage-makefile/pkg/converter/transformer  0.027s
ok  github.com/VDHewei/mage-makefile/pkg/runtime         12.199s
ok  github.com/VDHewei/mage-makefile/pkg/script          0.127s
```
