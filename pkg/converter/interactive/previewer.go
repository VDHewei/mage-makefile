package interactive

import (
	"fmt"
	"strings"

	"github.com/VDHewei/mage-makefile/pkg/converter/generator"
	"github.com/VDHewei/mage-makefile/pkg/converter/transformer"
)

// CodePreview 负责生成预览代码供用户审查。
type CodePreview struct {
	genOptions *generator.GeneratorOptions
}

// NewCodePreview 创建新的代码预览器。
func NewCodePreview(opts *generator.GeneratorOptions) *CodePreview {
	if opts == nil {
		opts = generator.DefaultGeneratorOptions()
	}
	return &CodePreview{genOptions: opts}
}

// PreviewTarget 生成单个目标的 Go 代码预览。
func (p *CodePreview) PreviewTarget(target transformer.IRTarget) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("// === Target / 目标: %s ===\n", target.Name))
	if target.IsPhony {
		sb.WriteString("// .PHONY\n")
	}
	if len(target.Prerequisites) > 0 {
		sb.WriteString(fmt.Sprintf("// Depends on / 依赖: %s\n",
			strings.Join(target.Prerequisites, ", ")))
	}

	// 生成函数签名
	sb.WriteString(fmt.Sprintf("func %s() error {\n", target.FuncName))

	// 添加依赖调用
	for _, dep := range target.Prerequisites {
		depFunc := generator.ToGoIdent(dep)
		sb.WriteString(fmt.Sprintf("\tmg.Deps(mg.F(%s))\n", depFunc))
	}

	// 添加命令
	for _, cmd := range target.Commands {
		if p.genOptions.AddOriginal && cmd.Original != "" {
			sb.WriteString(fmt.Sprintf("\t// Original: %s\n", cmd.Original))
		}
		if cmd.Transformed != "" {
			sb.WriteString(fmt.Sprintf("\t// %s\n", cmd.Transformed))
		}
	}

	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n")

	return sb.String()
}

// AcceptRejectPrompt 询问用户是否包含某个目标。
func AcceptRejectPrompt(targetName string) (bool, error) {
	for {
		fmt.Printf("\nInclude target '%s'? / 是否包含目标 '%s'? [Y/n]: ",
			targetName, targetName)
		var answer string
		_, err := fmt.Scanln(&answer)
		if err != nil {
			return true, nil // 默认接受
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer == "" || answer == "y" || answer == "yes" {
			return true, nil
		}
		if answer == "n" || answer == "no" {
			return false, nil
		}
		fmt.Println("Please enter y or n / 请输入 y 或 n")
	}
}
