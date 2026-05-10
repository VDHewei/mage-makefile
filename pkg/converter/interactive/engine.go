package interactive

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/VDHewei/mage-makefile/pkg/config"
	"github.com/VDHewei/mage-makefile/pkg/converter/generator"
	"github.com/VDHewei/mage-makefile/pkg/converter/parser"
	"github.com/VDHewei/mage-makefile/pkg/converter/transformer"
)

// InteractiveEngine 编排完整的交互式转换工作流。
type InteractiveEngine struct {
	categorizer *Categorizer
	previewer   *CodePreview
	genOpts     *generator.GeneratorOptions
	convertCfg  config.ConvertConfig
	reader      *bufio.Reader
}

// NewInteractiveEngine 创建新的交互引擎。
func NewInteractiveEngine(cfg *config.Config) *InteractiveEngine {
	opts := generator.DefaultGeneratorOptions()
	opts.ApplyConfig(&cfg.Convert)

	return &InteractiveEngine{
		categorizer: NewCategorizer(),
		previewer:   NewCodePreview(opts),
		genOpts:     opts,
		convertCfg:  cfg.Convert,
		reader:      bufio.NewReader(os.Stdin),
	}
}

// Run 执行完整的交互式转换流程。
func (e *InteractiveEngine) Run(makefileAST *parser.Makefile, ir *transformer.IR) error {
	// 第一步：显示欢迎信息和摘要
	e.showWelcome(makefileAST, ir)

	// 第二步：分类并显示目标
	categorized := e.categorizer.Categorize(ir)
	e.categorizer.DisplayCategories(categorized)

	// 第三步：目标选择
	selected, err := e.selectTargets(categorized)
	if err != nil {
		return fmt.Errorf("target selection: %w", err)
	}

	if len(selected) == 0 {
		fmt.Println("No targets selected. Exiting. / 未选择目标。退出。")
		return nil
	}

	// 第四步：预览并逐个接受/拒绝目标
	fmt.Println("\n--- Target Review / 目标审查 ---")
	finalSelected := e.reviewTargets(ir, selected)

	if len(finalSelected) == 0 {
		fmt.Println("No targets accepted. Exiting. / 未接受任何目标。退出。")
		return nil
	}

	// 第五步：将选择应用到 IR
	e.applySelection(ir, finalSelected)

	// 第六步：显示生成摘要
	fmt.Printf("\nGeneration summary / 生成摘要:\n")
	fmt.Printf("  Total targets / 总目标: %d\n", len(ir.Targets))
	fmt.Printf("  Selected targets / 已选择: %d\n", len(finalSelected))

	return nil
}

// showWelcome 显示欢迎横幅。
func (e *InteractiveEngine) showWelcome(mf *parser.Makefile, ir *transformer.IR) {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  mage-makefile Interactive Converter / 交互式转换器")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("\nParsed Makefile / 已解析 Makefile:\n")
	fmt.Printf("  Targets / 目标:      %d\n", len(ir.Targets))
	fmt.Printf("  Variables / 变量:    %d\n", len(ir.Variables))
	fmt.Printf("  Includes / 包含:     %d\n", len(ir.IncludePaths))
	fmt.Printf("  Conditionals / 条件: %d\n", len(mf.Conditionals))
	fmt.Println()
}

// selectTargets 呈现交互式目标选择界面。
func (e *InteractiveEngine) selectTargets(categorized map[TargetCategory][]TargetInfo) (map[string]bool, error) {
	selected := make(map[string]bool)

	// 收集所有目标，默认全选
	for _, cat := range []TargetCategory{CatStandard, CatUtility, CatMeta, CatPattern, CatConfiguration} {
		infos, ok := categorized[cat]
		if !ok {
			continue
		}
		for _, info := range infos {
			selected[info.Name] = true
		}
	}

	fmt.Println("\nSelect targets to convert / 选择要转换的目标:")
	fmt.Println("  Enter target name to toggle / 输入目标名称切换选择状态")
	fmt.Println("  Enter 'a' to select all / 输入 'a' 全选")
	fmt.Println("  Enter 'n' to select none / 输入 'n' 全不选")
	fmt.Println("  Enter 'd' when done / 输入 'd' 完成选择")
	fmt.Println()

	for {
		selectedCount := 0
		for _, s := range selected {
			if s {
				selectedCount++
			}
		}
		fmt.Printf("Currently selected / 当前已选择: %d/%d\n", selectedCount, len(selected))

		fmt.Print("> ")
		input, err := e.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "d", "done":
			fmt.Println()
			return selected, nil
		case "a", "all":
			for name := range selected {
				selected[name] = true
			}
			fmt.Println("All targets selected / 所有目标已选择")
		case "n", "none":
			for name := range selected {
				selected[name] = false
			}
			fmt.Println("No targets selected / 未选择任何目标")
		default:
			if input != "" {
				if _, exists := selected[input]; exists {
					selected[input] = !selected[input]
					status := "selected / 已选择"
					if !selected[input] {
						status = "deselected / 已取消"
					}
					fmt.Printf("  %s: %s\n", input, status)
				} else {
					// 尝试大小写不敏感匹配
					found := false
					for name := range selected {
						if strings.EqualFold(name, input) {
							selected[name] = !selected[name]
							status := "selected / 已选择"
							if !selected[name] {
								status = "deselected / 已取消"
							}
							fmt.Printf("  %s: %s\n", name, status)
							found = true
							break
						}
					}
					if !found {
						fmt.Printf("Unknown target / 未知目标: %s\n", input)
					}
				}
			}
		}
	}
}

// reviewTargets 逐个展示目标预览并询问用户是否接受。
func (e *InteractiveEngine) reviewTargets(ir *transformer.IR, selectedMap map[string]bool) map[string]bool {
	finalSelected := make(map[string]bool)

	for _, target := range ir.Targets {
		if !selectedMap[target.Name] {
			continue
		}

		fmt.Printf("\n%s\n", strings.Repeat("-", 50))
		preview := e.previewer.PreviewTarget(target)
		fmt.Println(preview)
		fmt.Println(strings.Repeat("-", 50))

		accepted, err := AcceptRejectPrompt(target.Name)
		if err != nil {
			fmt.Printf("Error reading input / 读取输入错误: %v\n", err)
			accepted = true
		}

		finalSelected[target.Name] = accepted
		status := "ACCEPTED / 已接受"
		if !accepted {
			status = "REJECTED / 已拒绝"
		}
		fmt.Printf("  => %s\n\n", status)
	}

	return finalSelected
}

// applySelection 过滤 IR，只保留被接受的目标。
func (e *InteractiveEngine) applySelection(ir *transformer.IR, selected map[string]bool) {
	var filtered []transformer.IRTarget
	for _, target := range ir.Targets {
		if selected[target.Name] {
			filtered = append(filtered, target)
		}
	}
	ir.Targets = filtered
}
