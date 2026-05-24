// Package transformer transforms a parsed Makefile AST into an intermediate
// representation (IR) suitable for magefile.go code generation.
package transformer

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/VDHewei/mage-makefile/pkg/converter/parser"
	"github.com/VDHewei/mage-makefile/pkg/script"
)

// IR represents the intermediate representation of a Makefile
// ready for Go code generation.
type IR struct {
	PackageName  string
	Variables    []IRVariable
	Targets      []IRTarget
	IncludePaths []string
	Platform     string
}

// IRVariable represents a resolved Makefile variable.
type IRVariable struct {
	Name  string
	Value string
	IsShell bool
}

// IRTarget represents a Makefile target ready for Go generation.
type IRTarget struct {
	Name          string
	FuncName      string
	Prerequisites []string
	Commands      []IRCommand
	IsPhony       bool
	IsDefault     bool
}

// IRCommand represents a single shell command after transformation.
type IRCommand struct {
	Original    string
	Transformed string
	Args        []string
	CanUseSh    bool
}

// Transformer transforms a Makefile AST into IR.
type Transformer struct {
	platform     *PlatformMapping
	scriptEngine script.ScriptEngine // 可选的脚本引擎，用于自定义 shell 命令转换
}

// NewTransformer creates a new Transformer with the default (current) platform.
func NewTransformer() *Transformer {
	return &Transformer{
		platform: NewPlatformMapping(""),
	}
}

// NewTransformerWithPlatform creates a new Transformer for a specific target OS.
func NewTransformerWithPlatform(targetOS string) *Transformer {
	return &Transformer{
		platform: NewPlatformMapping(targetOS),
	}
}

// NewTransformerWithEngine 创建带有指定脚本引擎和平台的 Transformer。
func NewTransformerWithEngine(targetOS string, engine script.ScriptEngine) *Transformer {
	return &Transformer{
		platform:      NewPlatformMapping(targetOS),
		scriptEngine:  engine,
	}
}

// Transform converts a Makefile AST into an IR.
func (t *Transformer) Transform(m *parser.Makefile) *IR {
	ir := &IR{
		PackageName: "main",
		Platform:    t.platform.TargetOS,
	}

	// Build variable map for resolution
	varMap := t.buildVarMap(m.Variables)

	// Transform variables
	for _, v := range m.Variables {
		// Step 1: Resolve $(shell ...) calls first (expands variables inside shell commands)
		shellResolved := t.resolveShellCalls(v.Value, varMap)
		// Step 2: Expand remaining $(VAR) references
		resolved := parser.ExpandVariable(shellResolved, varMap)
		// Step 3: Update varMap so subsequent references get the resolved value
		varMap[v.Name] = resolved
		ir.Variables = append(ir.Variables, IRVariable{
			Name:    v.Name,
			Value:   resolved,
			IsShell: v.IsShell,
		})
	}

	// Process conditionals - evaluate conditions and extract variable assignments
	// from the matched branch. This follows the same resolve-then-expand pipeline
	// used for top-level variables.
	for _, cond := range m.Conditionals {
		useTrue := evaluateCondition(cond, varMap)
		var bodyLines []string
		if useTrue {
			bodyLines = cond.TrueBody
		} else {
			bodyLines = cond.FalseBody
		}
		branchVars := extractVariableAssignments(bodyLines)
		for _, v := range branchVars {
			shellResolved := t.resolveShellCalls(v.Value, varMap)
			resolved := parser.ExpandVariable(shellResolved, varMap)
			varMap[v.Name] = resolved
			ir.Variables = append(ir.Variables, IRVariable{
				Name:    v.Name,
				Value:   resolved,
				IsShell: v.IsShell,
			})
		}
	}

	// Collect phony targets
	phonySet := t.collectPhonyTargets(m.Targets)

	// Transform targets
	defaultSet := false
	for _, target := range m.Targets {
		if target.Name == ".PHONY" {
			continue
		}

		isPhony := phonySet[target.Name]
		isDefault := !defaultSet
		defaultSet = true

		irTarget := IRTarget{
			Name:          target.Name,
			FuncName:      toFuncName(target.Name),
			Prerequisites: target.Prerequisites,
			IsPhony:       isPhony,
			IsDefault:     isDefault,
		}

		// Join backslash-continued recipe lines
		recipes := joinRecipeContinuations(target.Recipes)

		// Transform recipes
		for _, recipe := range recipes {
			// Resolve $(shell ...) calls first
			shellResolved := t.resolveShellCalls(recipe, varMap)
			// Resolve variables in the recipe
			resolved := parser.ExpandVariable(shellResolved, varMap)
			// Resolve automatic variables
			resolved = t.resolveAutoVars(resolved, target)

			cmd := t.transformCommand(resolved)
			irTarget.Commands = append(irTarget.Commands, cmd)
		}

		ir.Targets = append(ir.Targets, irTarget)
	}

	// Collect include paths
	for _, inc := range m.Includes {
		ir.IncludePaths = append(ir.IncludePaths, inc.Paths...)
	}

	return ir
}

// resolveShellCalls detects $(shell ...) calls, expands variables inside the
// command text first (matching GNU Make's expand-then-execute order), executes
// the command, and substitutes the output.
func (t *Transformer) resolveShellCalls(s string, varMap map[string]string) string {
	result := s
	for i := 0; i < 100; i++ {
		idx := strings.Index(result, "$(shell ")
		if idx == -1 {
			idx = strings.Index(result, "$(shell\t")
		}
		if idx == -1 {
			break
		}

		// Find the matching closing paren, tracking nesting depth
		// for $(...) inside the shell command (e.g., $(shell which $(GO)))
		depth := 1
		closeIdx := -1
		for j := idx + 8; j < len(result); j++ { // 8 = len("$(shell ")
			if result[j] == '(' {
				depth++
			} else if result[j] == ')' {
				depth--
				if depth == 0 {
					closeIdx = j
					break
				}
			}
		}
		if closeIdx == -1 {
			break
		}

		// Extract the inner command text
		cmdText := result[idx+8 : closeIdx]

		// Step 1: Expand variables inside the command (expand-then-execute order)
		expandedCmd := parser.ExpandVariable(cmdText, varMap)

		// Step 2: Execute the expanded command
		output := execShellCommand(expandedCmd)

		// Step 3 (可选): 如果配置了脚本引擎，通过脚本引擎转换输出内容
		if t.scriptEngine != nil {
			transformed, err := t.scriptEngine.Execute(output, varMap)
			if err == nil && transformed != "" {
				output = transformed
			}
		}

		// Step 4: Replace $(shell ...) with output
		result = result[:idx] + output + result[closeIdx+1:]
	}
	return result
}

// evaluateCondition evaluates a Makefile conditional and returns which branch to take.
// Returns true for the TrueBody (if branch), false for the FalseBody (else branch).
func evaluateCondition(cond *parser.Conditional, varMap map[string]string) bool {
	switch cond.Kind {
	case "ifeq":
		return evalIfeq(cond.Cond, varMap)
	case "ifneq":
		return !evalIfeq(cond.Cond, varMap)
	case "ifdef":
		return evalIfdef(cond.Cond, varMap)
	case "ifndef":
		return !evalIfdef(cond.Cond, varMap)
	default:
		return false
	}
}

// evalIfeq evaluates an "ifeq (arg1, arg2)" condition.
// The cond argument is the raw condition text (e.g., "(GOHOSTOS, windows)").
// Both arguments are expanded via ExpandVariable before comparison.
func evalIfeq(cond string, varMap map[string]string) bool {
	s := strings.TrimSpace(cond)
	// Strip outer parens: "(arg1, arg2)" -> "arg1, arg2"
	if len(s) > 0 && s[0] == '(' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == ')' {
		s = s[:len(s)-1]
	}
	s = strings.TrimSpace(s)

	// Find the comma at depth 0 (not inside nested parens)
	depth := 0
	commaIdx := -1
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '(' || s[i] == '{':
			depth++
		case s[i] == ')' || s[i] == '}':
			depth--
		case s[i] == ',' && depth == 0:
			commaIdx = i
			break
		}
	}
	if commaIdx == -1 {
		return false
	}

	left := strings.TrimSpace(s[:commaIdx])
	right := strings.TrimSpace(s[commaIdx+1:])

	// Expand both sides before comparing (GNU Make expand-then-evaluate)
	leftExpand := parser.ExpandVariable(left, varMap)
	rightExpand := parser.ExpandVariable(right, varMap)
	return leftExpand == rightExpand
}

// evalIfdef evaluates an "ifdef VAR" condition.
// Returns true if the variable is defined AND its value is non-empty.
// GNU Make treats a variable defined with empty value (FOO =) as undefined for ifdef.
func evalIfdef(cond string, varMap map[string]string) bool {
	name := strings.TrimSpace(cond)
	val, ok := varMap[name]
	return ok && val != ""
}

// extractVariableAssignments scans raw Makefile lines and extracts variable assignments.
// It skips blank lines, comments, and non-assignment lines.
func extractVariableAssignments(lines []string) []*parser.Variable {
	var result []*parser.Variable
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		v := parseVarLine(trimmed)
		if v != nil {
			result = append(result, v)
		}
	}
	return result
}

// parseVarLine attempts to parse a single Makefile line as a variable assignment.
// Returns nil if the line is not a valid assignment.
func parseVarLine(line string) *parser.Variable {
	operators := []string{":=", "+=", "?=", "!=", "="}
	for _, op := range operators {
		if idx := strings.Index(line, op); idx != -1 {
			name := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+len(op):])
			if name != "" {
				return &parser.Variable{
					Name:     name,
					Operator: op,
					Value:    value,
					IsShell:  op == "!=",
				}
			}
		}
	}
	return nil
}

// buildVarMap creates a variable lookup map from Variable nodes.
func (t *Transformer) buildVarMap(vars []*parser.Variable) map[string]string {
	m := make(map[string]string)
	for _, v := range vars {
		m[v.Name] = v.Value
	}
	return m
}

// collectPhonyTargets collects target names marked as .PHONY prerequisites.
func (t *Transformer) collectPhonyTargets(targets []*parser.Target) map[string]bool {
	phony := make(map[string]bool)
	for _, target := range targets {
		if target.Name == ".PHONY" {
			for _, p := range target.Prerequisites {
				phony[p] = true
			}
		}
	}
	return phony
}

// joinRecipeContinuations merges Makefile recipe lines joined by trailing backslash.
// For example:
//
//	@echo "hello" \
//	    "world"
//
// becomes:
//
//	@echo "hello" "world"
func joinRecipeContinuations(recipes []string) []string {
	var result []string
	buf := ""
	for _, r := range recipes {
		trimmed := strings.TrimRight(r, " \t\r\n")
		if buf != "" {
			// Continuation of previous line
			if strings.HasSuffix(trimmed, "\\") {
				buf += " " + strings.TrimSuffix(trimmed, "\\")
			} else {
				buf += " " + r
				result = append(result, buf)
				buf = ""
			}
		} else {
			if strings.HasSuffix(trimmed, "\\") {
				buf = strings.TrimSuffix(trimmed, "\\")
			} else {
				result = append(result, r)
			}
		}
	}
	if buf != "" {
		result = append(result, buf)
	}
	return result
}

// resolveAutoVars replaces automatic Make variables ($@, $<, $^, etc.)
func (t *Transformer) resolveAutoVars(cmd string, target *parser.Target) string {
	result := cmd

	// $@ - target name
	result = strings.ReplaceAll(result, "$@", target.Name)

	// $< - first prerequisite
	if len(target.Prerequisites) > 0 {
		result = strings.ReplaceAll(result, "$<", target.Prerequisites[0])
	} else {
		result = strings.ReplaceAll(result, "$<", "")
	}

	// $^ - all prerequisites
	result = strings.ReplaceAll(result, "$^", strings.Join(target.Prerequisites, " "))

	// $* - stem (part that matched % in pattern rule)
	result = strings.ReplaceAll(result, "$*", "")

	// $? - all prerequisites newer than target (approximate with $^)
	result = strings.ReplaceAll(result, "$?", strings.Join(target.Prerequisites, " "))

	// $(@D) and $(@F) - directory and file parts of $@
	reAtD := regexp.MustCompile(`\$\(@D\)`)
	result = reAtD.ReplaceAllStringFunc(result, func(s string) string {
		if idx := strings.LastIndex(target.Name, "/"); idx >= 0 {
			return target.Name[:idx]
		}
		return "."
	})
	reAtF := regexp.MustCompile(`\$\(@F\)`)
	result = reAtF.ReplaceAllStringFunc(result, func(s string) string {
		if idx := strings.LastIndex(target.Name, "/"); idx >= 0 {
			return target.Name[idx+1:]
		}
		return target.Name
	})

	// Handle $(var:pattern=replacement) substitution references
	result = t.expandSubstRefs(result)

	return result
}

// expandSubstRefs handles $(VAR:pattern=replacement) substitutions.
func (t *Transformer) expandSubstRefs(s string) string {
	re := regexp.MustCompile(`\$\(([^:)]+):([^=]+)=([^)]+)\)`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}
		// Return a placeholder - actual expansion needs varMap context
		return "${" + parts[1] + ":" + parts[2] + "=" + parts[3] + "}"
	})
}

// transformCommand converts a shell command into an IRCommand.
func (t *Transformer) transformCommand(cmd string) IRCommand {
	irCmd := IRCommand{
		Original: cmd,
	}

	// Skip empty commands and comments
	trimmed := strings.TrimSpace(cmd)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		irCmd.Transformed = cmd
		return irCmd
	}

	// Strip @ prefix (silent mode in make)
	if strings.HasPrefix(trimmed, "@") {
		trimmed = trimmed[1:]
	}

	// Strip - prefix (ignore errors)
	ignoreErrors := strings.HasPrefix(trimmed, "-")
	if ignoreErrors {
		trimmed = trimmed[1:]
	}

	// Strip + prefix (always execute)
	trimmed = strings.TrimPrefix(trimmed, "+")

	// Apply platform-specific command mapping
	trimmed = t.platform.MapCommand(trimmed)

	// Tokenize command into args
	args := shellSplit(trimmed)

	irCmd.Transformed = trimmed
	irCmd.Args = args
	irCmd.CanUseSh = !ignoreErrors && !strings.Contains(trimmed, "&&") &&
		!strings.Contains(trimmed, "||") && !strings.Contains(trimmed, "|") &&
		!strings.Contains(trimmed, ";") && !strings.Contains(trimmed, ">") &&
		!strings.Contains(trimmed, ">>") && !strings.Contains(trimmed, "<") &&
		!hasLeadingEnvVars(args)

	return irCmd
}

// hasLeadingEnvVars checks if the first argument(s) look like KEY=VALUE env prefixes.
// Example: "FOX_SKIP_LLAMA=1 cargo fmt" has an env prefix.
func hasLeadingEnvVars(args []string) bool {
	seenVar := false
	for _, arg := range args {
		if looksLikeEnvVar(arg) {
			seenVar = true
			continue
		}
		// First non-env-var argument — return true only if preceded by env var(s)
		return seenVar
	}
	return false // only env vars, no command
}

// looksLikeEnvVar checks if a string matches the pattern KEY=VALUE where KEY is a
// valid shell variable name (starts with letter/underscore, contains alnum/underscore).
func looksLikeEnvVar(s string) bool {
	eq := strings.IndexByte(s, '=')
	if eq < 1 {
		return false
	}
	// Key must start with letter or underscore
	if !((s[0] >= 'A' && s[0] <= 'Z') || (s[0] >= 'a' && s[0] <= 'z') || s[0] == '_') {
		return false
	}
	for i := 1; i < eq; i++ {
		c := s[i]
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

// toFuncName converts a Makefile target name to a valid Go function name.
// Examples: build -> Build, serve-dev -> ServeDev, test.foo -> TestFoo
func toFuncName(name string) string {
	// Handle special characters
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, ".", " ")
	name = strings.ReplaceAll(name, "_", " ")

	parts := strings.Fields(name)
	for i, p := range parts {
		if len(p) > 0 {
			r := []rune(p)
			r[0] = unicode.ToUpper(r[0])
			parts[i] = string(r)
		}
	}

	return strings.Join(parts, "")
}

// shellSplit splits a shell command into arguments respecting quoting.
func shellSplit(cmd string) []string {
	var args []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	for _, ch := range cmd {
		if escaped {
			current.WriteRune(ch)
			escaped = false
			continue
		}

		switch {
		case ch == '\\' && !inSingle:
			escaped = true
		case ch == '\'' && !inDouble:
			inSingle = !inSingle
		case ch == '"' && !inSingle:
			inDouble = !inDouble
		case ch == ' ' && !inSingle && !inDouble:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
