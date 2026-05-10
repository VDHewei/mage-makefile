package interactive

import (
	"bytes"
	"os"
	"testing"

	"github.com/VDHewei/mage-makefile/pkg/converter/generator"
	"github.com/VDHewei/mage-makefile/pkg/converter/transformer"
	"github.com/stretchr/testify/assert"
)

// makeTestIR 创建测试用的 IR。
func makeTestIR(targets ...transformer.IRTarget) *transformer.IR {
	return &transformer.IR{
		PackageName: "main",
		Targets:     targets,
	}
}

func TestCategorize_StandardTargets(t *testing.T) {
	ir := makeTestIR(
		transformer.IRTarget{Name: "build", FuncName: "Build"},
		transformer.IRTarget{Name: "test", FuncName: "Test"},
		transformer.IRTarget{Name: "deploy", FuncName: "Deploy"},
	)
	c := NewCategorizer()
	result := c.Categorize(ir)

	standard, ok := result[CatStandard]
	assert.True(t, ok)
	assert.Len(t, standard, 3)
	// 按字母排序：build, deploy, test
	assert.Equal(t, "build", standard[0].Name)
	assert.Equal(t, "deploy", standard[1].Name)
	assert.Equal(t, "test", standard[2].Name)
}

func TestCategorize_UtilityTargets(t *testing.T) {
	ir := makeTestIR(
		transformer.IRTarget{Name: "clean", FuncName: "Clean"},
		transformer.IRTarget{Name: "install", FuncName: "Install"},
		transformer.IRTarget{Name: "help", FuncName: "Help"},
	)
	c := NewCategorizer()
	result := c.Categorize(ir)

	utility, ok := result[CatUtility]
	assert.True(t, ok)
	assert.Len(t, utility, 3)
}

func TestCategorize_MetaTargets(t *testing.T) {
	ir := makeTestIR(
		transformer.IRTarget{Name: ".PHONY", FuncName: "PHONY"},
		transformer.IRTarget{Name: ".DEFAULT", FuncName: "DEFAULT"},
		transformer.IRTarget{Name: "build", FuncName: "Build"},
	)
	c := NewCategorizer()
	result := c.Categorize(ir)

	meta, ok := result[CatMeta]
	assert.True(t, ok)
	assert.Len(t, meta, 2)

	standard, ok := result[CatStandard]
	assert.True(t, ok)
	assert.Len(t, standard, 1)
	assert.Equal(t, "build", standard[0].Name)
}

func TestCategorize_PatternTargets(t *testing.T) {
	ir := makeTestIR(
		transformer.IRTarget{Name: "%.o: %.c", FuncName: "_O_C"},
	)
	c := NewCategorizer()
	result := c.Categorize(ir)

	pattern, ok := result[CatPattern]
	assert.True(t, ok)
	assert.Len(t, pattern, 1)
}

func TestCategorize_EmptyIR(t *testing.T) {
	ir := makeTestIR()
	c := NewCategorizer()
	result := c.Categorize(ir)
	assert.Empty(t, result)
}

func TestDisplayCategories_Output(t *testing.T) {
	ir := makeTestIR(
		transformer.IRTarget{Name: "build", FuncName: "Build"},
		transformer.IRTarget{Name: "clean", FuncName: "Clean"},
	)
	c := NewCategorizer()
	categorized := c.Categorize(ir)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	c.DisplayCategories(categorized)

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = oldStdout
	r.Close()

	output := buf.String()
	assert.Contains(t, output, "Build/Test/Deploy")
	assert.Contains(t, output, "Utility")
	assert.Contains(t, output, "[x] build")
	assert.Contains(t, output, "[x] clean")
}

func TestPreviewTarget_Basic(t *testing.T) {
	preview := NewCodePreview(nil)
	target := transformer.IRTarget{
		Name:     "build",
		FuncName: "Build",
		Commands: []transformer.IRCommand{
			{Original: "go build .", Transformed: "go build .", Args: []string{"go", "build", "."}, CanUseSh: true},
		},
	}
	result := preview.PreviewTarget(target)
	assert.Contains(t, result, "Target / 目标: build")
	assert.Contains(t, result, "func Build() error {")
	assert.Contains(t, result, "return nil")
	assert.NotContains(t, result, "mg.Deps")
}

func TestPreviewTarget_WithDeps(t *testing.T) {
	preview := NewCodePreview(nil)
	target := transformer.IRTarget{
		Name:          "all",
		FuncName:      "All",
		Prerequisites: []string{"build", "test"},
	}
	result := preview.PreviewTarget(target)
	assert.Contains(t, result, "Depends on / 依赖: build, test")
	assert.Contains(t, result, "mg.Deps(mg.F(Build))")
}

func TestPreviewTarget_AddOriginal(t *testing.T) {
	opts := generator.DefaultGeneratorOptions()
	opts.AddOriginal = true
	preview := NewCodePreview(opts)
	target := transformer.IRTarget{
		Name:     "build",
		FuncName: "Build",
		Commands: []transformer.IRCommand{
			{Original: "go build -o bin/app .", Args: []string{"go", "build", "-o", "bin/app", "."}},
		},
	}
	result := preview.PreviewTarget(target)
	assert.Contains(t, result, "Original: go build -o bin/app .")
}

func TestAcceptRejectPrompt_Default(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		w.Write([]byte("\n"))
		w.Close()
	}()

	accepted, err := AcceptRejectPrompt("test-target")
	os.Stdin = oldStdin
	r.Close()

	assert.NoError(t, err)
	assert.True(t, accepted)
}

func TestAcceptRejectPrompt_Yes(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		w.Write([]byte("y\n"))
		w.Close()
	}()

	accepted, err := AcceptRejectPrompt("test")
	os.Stdin = oldStdin
	r.Close()
	assert.NoError(t, err)
	assert.True(t, accepted)
}

func TestAcceptRejectPrompt_No(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		w.Write([]byte("n\n"))
		w.Close()
	}()

	accepted, err := AcceptRejectPrompt("test")
	os.Stdin = oldStdin
	r.Close()
	assert.NoError(t, err)
	assert.False(t, accepted)
}

func TestApplySelection(t *testing.T) {
	ir := makeTestIR(
		transformer.IRTarget{Name: "build", FuncName: "Build"},
		transformer.IRTarget{Name: "clean", FuncName: "Clean"},
		transformer.IRTarget{Name: "test", FuncName: "Test"},
	)

	var engine InteractiveEngine
	selected := map[string]bool{"build": true, "test": true}
	engine.applySelection(ir, selected)

	assert.Len(t, ir.Targets, 2)
	assert.Equal(t, "build", ir.Targets[0].Name)
	assert.Equal(t, "test", ir.Targets[1].Name)
}

func TestClassifyTarget_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		expected TargetCategory
	}{
		{"", CatStandard},
		{"Build", CatStandard},
		{"CLEAN", CatUtility},
		{"config-server", CatConfiguration},
		{"%.o", CatPattern},
		{".PRECIOUS", CatMeta},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyTarget(tt.name)
			assert.Equal(t, tt.expected, result)
		})
	}
}
