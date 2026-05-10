// Package generator converts the intermediate representation (IR)
// of a Makefile into valid, compilable magefile.go code.
package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/VDHewei/mage-makefile/pkg/config"
	"github.com/VDHewei/mage-makefile/pkg/converter/transformer"
)

// GeneratorOptions 控制代码生成的行为。
type GeneratorOptions struct {
	// 输出样式：standard, verbose, minimal
	OutputStyle string
	// 是否添加 Makefile 源码注释
	AddComments bool
	// 是否包含原始 shell 命令作为注释
	AddOriginal bool
	// 是否按类别分组函数
	GroupByCategory bool
	// 用户自定义别名
	Aliases []AliasOverride
}

// AliasOverride 表示来自配置的自定义别名映射。
type AliasOverride struct {
	Alias  string
	Target string
}

// DefaultGeneratorOptions 返回默认生成选项。
func DefaultGeneratorOptions() *GeneratorOptions {
	return &GeneratorOptions{
		OutputStyle:     "standard",
		AddComments:     true,
		AddOriginal:     false,
		GroupByCategory: false,
	}
}

// ApplyConfig 将 ConvertConfig 设置应用到 GeneratorOptions。
func (opts *GeneratorOptions) ApplyConfig(cfg *config.ConvertConfig) {
	if cfg == nil {
		return
	}
	if cfg.OutputStyle != "" {
		opts.OutputStyle = cfg.OutputStyle
	}
	opts.AddComments = cfg.AddComments
	opts.AddOriginal = cfg.AddOriginal
	opts.GroupByCategory = cfg.GroupByCategory

	for _, a := range cfg.Aliases {
		opts.Aliases = append(opts.Aliases, AliasOverride{
			Alias:  a.Alias,
			Target: a.Target,
		})
	}
}

// Generator produces magefile.go code from a Makefile IR.
type Generator struct {
	ir   *transformer.IR
	opts *GeneratorOptions
}

// NewGenerator creates a new Generator from a Makefile IR.
func NewGenerator(ir *transformer.IR) *Generator {
	return &Generator{ir: ir, opts: DefaultGeneratorOptions()}
}

// NewGeneratorWithConfig 创建具有自定义选项的 Generator。
func NewGeneratorWithConfig(ir *transformer.IR, opts *GeneratorOptions) *Generator {
	if opts == nil {
		opts = DefaultGeneratorOptions()
	}
	return &Generator{ir: ir, opts: opts}
}

// Generate produces the complete magefile.go source code.
func (g *Generator) Generate() (string, error) {
	if g.ir == nil {
		return "", fmt.Errorf("IR is nil")
	}

	data := &TemplateData{
		PackageName: g.ir.PackageName,
	}

	// Collect imports
	data.Imports = g.collectImports()

	// Generate constants from variables
	data.Constants, data.Vars = g.generateVariables()

	// Generate target functions
	data.Functions = g.generateFunctions()

	// Generate aliases
	data.Aliases = g.generateAliases()

	return renderTemplate(data)
}

// collectImports determines which Go packages are needed.
func (g *Generator) collectImports() []string {
	needed := map[string]bool{
		"github.com/magefile/mage/sh": false,
	}

	for _, target := range g.ir.Targets {
		for _, cmd := range target.Commands {
			if cmd.CanUseSh && len(cmd.Args) > 0 {
				needed["github.com/magefile/mage/sh"] = true
			}
			if !cmd.CanUseSh && len(cmd.Args) > 0 {
				needed["os/exec"] = true
			}
		}
	}

	// If we need os/exec, we also need "os" for os.Stdout/os.Stderr
	if needed["os/exec"] {
		needed["os"] = true
	}

	// Always include mg for dependencies between targets
	hasDeps := false
	for _, target := range g.ir.Targets {
		if len(target.Prerequisites) > 0 {
			hasDeps = true
			break
		}
	}
	if hasDeps {
		needed["github.com/magefile/mage/mg"] = true
	}

	var imports []string
	for pkg, used := range needed {
		if used {
			imports = append(imports, pkg)
		}
	}
	sort.Strings(imports)
	return imports
}

// generateVariables converts IR variables into Go constants and vars.
func (g *Generator) generateVariables() ([]ConstDef, []VarDef) {
	var consts []ConstDef
	var vars []VarDef

	for _, v := range g.ir.Variables {
		if v.IsShell {
			vars = append(vars, VarDef{
				Name:  ToGoIdent(v.Name),
				Value: v.Value,
			})
		} else {
			consts = append(consts, ConstDef{
				Name:  ToGoIdent(v.Name),
				Value: v.Value,
			})
		}
	}

	return consts, vars
}

// generateFunctions creates function definitions for each target.
func (g *Generator) generateFunctions() []FuncDef {
	var funcs []FuncDef

	for _, target := range g.ir.Targets {
		if target.Name == ".PHONY" {
			continue
		}

		f := FuncDef{
			Name: target.FuncName,
		}

		// Build description from prerequisites
		if len(target.Prerequisites) > 0 {
			f.Description = fmt.Sprintf("runs %s (%s)",
				target.Name,
				strings.Join(target.Prerequisites, ", "))
		} else {
			f.Description = fmt.Sprintf("runs %s", target.Name)
		}

		// Generate function body
		f.Body = g.generateFunctionBody(target)

		funcs = append(funcs, f)
	}

	return funcs
}

// generateFunctionBody creates the body of a mage target function.
func (g *Generator) generateFunctionBody(target transformer.IRTarget) string {
	var body strings.Builder
	body.WriteString("\n")

	// Add dependency calls
	for _, dep := range target.Prerequisites {
		depFunc := ToGoIdent(dep)
		body.WriteString(fmt.Sprintf("\tmg.Deps(mg.F(%s))\n", depFunc))
	}

	// Check if any commands use os/exec (complex commands)
	// If so, declare cmd once at function scope to avoid := redeclaration
	hasComplexCmd := false
	for _, cmd := range target.Commands {
		if !cmd.CanUseSh && cmd.Transformed != "" {
			hasComplexCmd = true
			break
		}
	}
	if hasComplexCmd {
		body.WriteString("\tvar cmd *exec.Cmd\n")
	}

	// Generate command execution
	for _, cmd := range target.Commands {
		if cmd.Transformed == "" {
			continue
		}
		// 如果用户配置了 AddOriginal，添加原始命令注释
		if g.opts.AddOriginal && cmd.Original != "" {
			body.WriteString(fmt.Sprintf("\t// Original: %s\n", strings.TrimSpace(cmd.Original)))
		}
		body.WriteString(g.generateCommand(cmd))
	}

	body.WriteString("\treturn nil\n")
	return body.String()
}

// generateCommand generates Go code for a single shell command.
func (g *Generator) generateCommand(cmd transformer.IRCommand) string {
	var sb strings.Builder

	if cmd.CanUseSh && len(cmd.Args) > 0 {
		// Use sh.Run for simple commands
		sb.WriteString("\tif err := sh.Run(")
		for i, arg := range cmd.Args {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%q", arg))
		}
		sb.WriteString("); err != nil {\n")
		sb.WriteString("\t\treturn err\n")
		sb.WriteString("\t}\n")
	} else if cmd.Transformed != "" {
		// Use os/exec for complex commands
		sb.WriteString("\t// Complex command: " + cmd.Transformed + "\n")
		sb.WriteString(fmt.Sprintf("\tcmd = exec.Command(%q", cmd.Args[0]))
		for _, arg := range cmd.Args[1:] {
			sb.WriteString(fmt.Sprintf(", %q", arg))
		}
		sb.WriteString(")\n")
		sb.WriteString("\tcmd.Stdout = os.Stdout\n")
		sb.WriteString("\tcmd.Stderr = os.Stderr\n")
		sb.WriteString("\tif err := cmd.Run(); err != nil {\n")
		sb.WriteString("\t\treturn err\n")
		sb.WriteString("\t}\n")
	}

	return sb.String()
}

// generateAliases creates short aliases for common targets.
func (g *Generator) generateAliases() []AliasDef {
	var aliases []AliasDef
	seen := make(map[string]bool)

	// 添加配置中的自定义别名
	for _, a := range g.opts.Aliases {
		if !seen[a.Alias] {
			aliases = append(aliases, AliasDef{Alias: a.Alias, Target: a.Target})
			seen[a.Alias] = true
		}
	}

	for _, target := range g.ir.Targets {
		if target.Name == ".PHONY" {
			continue
		}

		// Generate aliases for known patterns
		switch target.Name {
		case "build":
			if !seen["b"] {
				aliases = append(aliases, AliasDef{Alias: "b", Target: "Build"})
				seen["b"] = true
			}
		case "clean":
			if !seen["c"] {
				aliases = append(aliases, AliasDef{Alias: "c", Target: "Clean"})
				seen["c"] = true
			}
		case "test":
			if !seen["t"] {
				aliases = append(aliases, AliasDef{Alias: "t", Target: "Test"})
				seen["t"] = true
			}
		}
	}

	return aliases
}

// ToGoIdent converts a Makefile variable/target name to a valid Go identifier.
// 将 Makefile 变量/目标名转换为合法的 Go 标识符。
func ToGoIdent(name string) string {
	// Replace non-identifier characters
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, name)

	// Must start with letter or underscore
	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		result = "_" + result
	}

	// Capitalize for export
	if len(result) > 0 {
		upper := strings.ToUpper(result[:1])
		result = upper + result[1:]
	}

	return result
}
