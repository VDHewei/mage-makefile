# Phase 3: CLI Integration / 第三阶段：CLI 集成

## Status / 状态: Completed / 已完成

## Overview / 概述

Phase 3 connects the existing backend packages to the CLI entry point (`cmd/makego/main.go`), fixes the build, wires up the `convert` pipeline correctly, and adds CLI-level tests.

第三阶段将现有的后端包连接到 CLI 入口点，修复构建问题，正确连接 `convert` 管道，并添加 CLI 级别的测试。

## Taks / 任务

| Task | EN Description | CN Description | Status |
|------|---------------|---------------|--------|
| 3.1 | Fix go.mod — add cobra dependency | 修复 go.mod — 添加 cobra 依赖 | Done / 完成 |
| 3.2 | Fix convert pipeline — add transformer step | 修复 convert 管道 — 添加 transformer 步骤 | Done / 完成 |
| 3.3 | Wire up detect command with CompatChecker | 连接 detect 命令与 CompatChecker | Done / 完成 |
| 3.4 | Verify CLI builds successfully | 验证 CLI 构建成功 | Done / 完成 |
| 3.5 | Create CLI integration tests (convert pipeline E2E) | 创建 CLI 集成测试（convert 端到端管道） | Done / 完成 |
| 3.6 | Full test suite verification | 全量测试验证 | Done / 完成 |

## Files to Modify / 待修改文件

### 1. `go.mod`
- **Action**: Add `github.com/spf13/cobra` dependency
- **Command**: `go get github.com/spf13/cobra@latest`

### 2. `cmd/makego/main.go`
- **Action**: Fix 3 things

**2a. Add import for transformer package:**
```go
"github.com/VDHewei/mage-makefile/pkg/converter/transformer"
```

**2b. Fix `runConvert()` — insert transformer step (BUG FIX at line ~242):**
```go
// Old (buggy):
// gen := generator.NewGenerator(makefileAST)

// New (fixed):
tr := transformer.NewTransformerWithPlatform(targetPlatform)
ir := tr.Transform(makefileAST)
gen := generator.NewGenerator(ir)
```

**2c. Wire up detect — replace Phase 4 placeholder with real CompatChecker:**
```go
// In runDetect(), replace placeholder at ~line 300:
// Old: fmt.Println("(Makefile compatibility check coming in Phase 4)")
// New:
data, err := os.ReadFile(target)
if err == nil {
    mf, parseErr := parser.NewParser(string(data)).Parse()
    if parseErr == nil {
        checker := runtime.NewCompatChecker()
        report := checker.CheckMakefileCompatibility(mf)
        fmt.Println(report.String())
    }
}
```

### 3. `cmd/makego/main_test.go` **(New File)**
- **Action**: Create CLI integration tests
- **Test functions**:
  - `TestCLI_ConvertBasic` — Convert a basic Makefile with a target and recipe
  - `TestCLI_ConvertWithVars` — Convert Makefile with variables
  - `TestCLI_ConvertWithDeps` — Convert Makefile with prerequisites
  - `TestCLI_Version` — Verify version command works

## Execution Order / 执行顺序

```
Step 1: go get github.com/spf13/cobra@latest → go mod tidy
Step 2: Edit main.go — add transformer import + fix runConvert
Step 3: Edit main.go — wire up detect command
Step 4: go build ./cmd/makego/ (verify build)
Step 5: Create cmd/makego/main_test.go
Step 6: go test ./cmd/... (verify CLI tests)
Step 7: go test ./... (full suite)
Step 8: Update this plan status to Completed
```

## Key Design Decisions / 关键设计决策

1. **Compile and hub commands remain stubs** — labeled Phase 6 / Phase 8 in existing code; no changes to those
2. **Transformer creation uses `targetPlatform`** — the flag already exists on convert command, now it's actually used
3. **Tests use `rootCmd.SetArgs()` + cobra's built-in test patterns** — avoids exec.Command overhead
4. **No parallel tests** — cobra uses global state and `os.Chdir` is process-wide

## Dependencies / 依赖

- Phase 1 (config system) — used by PersistentPreRunE → loads config before any command
- Phase 2 (parser) — used by convert + detect commands
- Transformer package — needs to be imported, already fully tested (12 tests pass)
- Runtime package — CompatChecker exists and is tested

## Risks / 风险

- Cobra global state in tests — each test must clean up args after itself
- `go get` may pick a newer cobra version — pin to v1.10+ (compatible with current usage)
