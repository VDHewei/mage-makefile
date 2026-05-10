// Package interactive provides interactive conversion workflow for mage-makefile.
// 包含目标分类、选择和预览功能，支持用户交互式转换 Makefile。
package interactive

import (
	"fmt"
	"sort"
	"strings"

	"github.com/VDHewei/mage-makefile/pkg/converter/transformer"
)

// TargetCategory 表示 Makefile 目标的分类。
type TargetCategory int

const (
	// CatStandard 标准目标：构建/测试/部署等
	CatStandard TargetCategory = iota
	// CatUtility 工具目标：clean/install/help 等
	CatUtility
	// CatMeta 元目标：.PHONY/.DEFAULT 等
	CatMeta
	// CatPattern 模式规则：%.o: %.c 等
	CatPattern
	// CatConfiguration 配置目标：config-xxx 等
	CatConfiguration
)

// TargetInfo 保存单个目标的信息，用于显示和用户选择。
type TargetInfo struct {
	Name         string
	FuncName     string
	Category     TargetCategory
	Commands     int      // 配方数量
	Dependencies []string // 依赖目标列表
	IsSelected   bool     // 用户选择状态
	Description  string   // 描述文本
}

// Label 返回分类的中英双语标签。
func (tc TargetCategory) Label() string {
	switch tc {
	case CatStandard:
		return "Build/Test/Deploy / 构建/测试/部署"
	case CatUtility:
		return "Utility / 工具"
	case CatMeta:
		return "Meta / 元目标"
	case CatPattern:
		return "Pattern Rules / 模式规则"
	case CatConfiguration:
		return "Configuration / 配置"
	default:
		return "Other / 其他"
	}
}

// Categorizer 负责对 Makefile 目标进行分类和组织，供用户交互选择。
type Categorizer struct{}

// NewCategorizer 创建新的分类器。
func NewCategorizer() *Categorizer {
	return &Categorizer{}
}

// Categorize 接收 IR 并返回按分类组织的目标信息。
func (c *Categorizer) Categorize(ir *transformer.IR) map[TargetCategory][]TargetInfo {
	result := make(map[TargetCategory][]TargetInfo)

	for _, target := range ir.Targets {
		info := TargetInfo{
			Name:         target.Name,
			FuncName:     target.FuncName,
			Commands:     len(target.Commands),
			Dependencies: target.Prerequisites,
			IsSelected:   true, // 默认全部选中
		}

		cat := classifyTarget(target.Name)
		result[cat] = append(result[cat], info)
	}

	// 在每个分类内按名称排序
	for _, infos := range result {
		sort.Slice(infos, func(i, j int) bool {
			return infos[i].Name < infos[j].Name
		})
	}

	return result
}

// DisplayCategories 将分类后的目标打印到 stdout。
func (c *Categorizer) DisplayCategories(categorized map[TargetCategory][]TargetInfo) {
	// 按固定顺序打印分类
	cats := []TargetCategory{CatStandard, CatUtility, CatMeta, CatPattern, CatConfiguration}
	for _, cat := range cats {
		infos, ok := categorized[cat]
		if !ok || len(infos) == 0 {
			continue
		}
		fmt.Printf("\n=== %s (%d) ===\n", cat.Label(), len(infos))
		for _, info := range infos {
			sel := "[ ]"
			if info.IsSelected {
				sel = "[x]"
			}
			fmt.Printf("  %s %s", sel, info.Name)
			if len(info.Dependencies) > 0 {
				fmt.Printf("  -> %s", strings.Join(info.Dependencies, ", "))
			}
			fmt.Printf("  (%d commands / 个命令)\n", info.Commands)
		}
	}
}

// classifyTarget 根据目标名称判断其分类。
func classifyTarget(name string) TargetCategory {
	lower := strings.ToLower(name)

	// 元目标检查
	metaTargets := map[string]bool{
		".phony":              true,
		".default":            true,
		".silent":             true,
		".ignore":             true,
		".export_all_variables": true,
		".low_resolution_time": true,
		".secondary":          true,
		".delete_on_error":    true,
		".intermediate":       true,
		".notparallel":        true,
		".oneshell":           true,
		".posix":              true,
		".precious":           true,
		".suffixes":           true,
		".d":                  true,
	}
	if metaTargets[lower] {
		return CatMeta
	}

	// 工具目标检查
	utilityTargets := map[string]bool{
		"clean": true, "distclean": true, "mrproper": true,
		"install": true, "uninstall": true, "setup": true,
		"init": true, "help": true, "info": true,
		"lint": true, "vet": true, "format": true, "fmt": true,
	}
	if utilityTargets[lower] {
		return CatUtility
	}

	// 模式规则（包含 %）
	if strings.Contains(name, "%") {
		return CatPattern
	}

	// 配置前缀检查
	configPrefixes := []string{"config-", "configure", "conf-"}
	for _, prefix := range configPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return CatConfiguration
		}
	}

	return CatStandard
}
