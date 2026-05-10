package parser

import "strings"

// gnuMakeFuncs is the set of known GNU Make built-in function names.
// When ExpandVariable encounters $(func_name ...), it skips expansion
// and preserves the text as-is.
var gnuMakeFuncs = map[string]bool{
	// Text functions
	"subst": true, "patsubst": true, "strip": true,
	"findstring": true, "filter": true, "filter-out": true,
	"sort": true, "word": true, "words": true, "wordlist": true,
	"firstword": true, "lastword": true,
	// File name functions
	"dir": true, "notdir": true, "suffix": true, "basename": true,
	"addsuffix": true, "addprefix": true, "join": true,
	"wildcard": true, "realpath": true, "abspath": true,
	// Shell execution
	"shell": true,
	// Control functions
	"foreach": true, "if": true, "and": true, "or": true,
	"call": true, "value": true, "eval": true,
	// Origin/info functions
	"origin": true, "flavor": true,
	// Error functions
	"error": true, "warning": true, "info": true,
	// File operations
	"file": true,
	// Guile (not supported, but skip to avoid mangling)
	"guile": true,
}

// extractFuncName returns the first word of a $(...) content.
// For "$(subst a,b,$(VAR))", content is "subst a,b,$(VAR)" and returns "subst".
func extractFuncName(content string) string {
	idx := strings.IndexAny(content, " \t")
	if idx == -1 {
		return content
	}
	return content[:idx]
}

// isGnuMakeFunc returns true if the name is a known GNU Make built-in function.
func isGnuMakeFunc(name string) bool {
	return gnuMakeFuncs[name]
}

// Node is the interface for all AST nodes.
type Node interface {
	nodeMarker()
}

// Makefile represents the entire parsed Makefile.
type Makefile struct {
	Targets    []*Target
	Variables  []*Variable
	Includes   []*Include
	Exports    []*Export
	Conditionals []*Conditional
	Comments   []string
}

// Target represents a Makefile target with its prerequisites and recipe.
type Target struct {
	Name          string
	Prerequisites []string
	Recipes       []string
	IsPhony       bool
	IsPattern     bool   // true if target contains %
	Line          int
}

func (*Target) nodeMarker() {}

// Variable represents a variable assignment.
type Variable struct {
	Name     string
	Operator string // =, :=, +=, ?=, !=
	Value    string
	IsShell  bool   // true if operator is !=
	Line     int
}

func (*Variable) nodeMarker() {}

// Include represents an include directive.
type Include struct {
	Paths     []string
	IsOptional bool // true for -include/sinclude
	Line      int
}

func (*Include) nodeMarker() {}

// Export represents an export/unexport directive.
type Export struct {
	Variable string
	IsExport bool   // true for export, false for unexport
	Line     int
}

func (*Export) nodeMarker() {}

// Conditional represents an ifeq/ifneq/ifdef/ifndef block.
type Conditional struct {
	Kind     string // ifeq, ifneq, ifdef, ifndef
	Cond     string // the condition expression
	TrueBody []string
	FalseBody []string // else branch
	Line     int
}

func (*Conditional) nodeMarker() {}

// ExpandVariable resolves $(VAR) references in a string using the variable map.
// It guards against recursive self-references (e.g., PATH=$(HOME)/bin:$(PATH)).
// GNU Make built-in function calls like $(subst ...), $(dir ...) are evaluated
// when supported, otherwise preserved as-is.
func ExpandVariable(s string, vars map[string]string) string {
	result := s
	seen := make(map[string]bool)
	pos := 0 // tracks scan position — avoids infinite loops when skipping functions
	for i := 0; i < 10000; i++ {
		if pos >= len(result) {
			break
		}
		// Find next $( or ${
		dollar := strings.Index(result[pos:], "$(")
		if dollar == -1 {
			dollar = strings.Index(result[pos:], "${")
		}
		if dollar == -1 {
			break
		}
		dollar += pos // convert to absolute position

		isBrace := result[dollar+1] == '{'
		openChar := byte('(')
		closeChar := byte(')')
		if isBrace {
			openChar = '{'
			closeChar = '}'
		}

		// Depth-aware paren/brace matching
		close := -1
		depth := 1
		for j := dollar + 2; j < len(result); j++ {
			if result[j] == openChar {
				depth++
			} else if result[j] == closeChar {
				depth--
				if depth == 0 {
					close = j
					break
				}
			}
		}
		if close == -1 {
			break
		}

		// Extract content between $( and matching )
		content := result[dollar+2 : close]

		// Check if this is a GNU Make built-in function call
		if !isBrace {
			funcName := extractFuncName(content)
			if isGnuMakeFunc(funcName) {
				// Try to evaluate the function
				if evaluated, ok := evalGnuMakeFunc(funcName, content, vars); ok {
					result = result[:dollar] + evaluated + result[close+1:]
					pos = dollar // re-scan to catch new expansions in result
				} else {
					// Function not supported — advance past it
					pos = close + 1
				}
				continue
			}
		}

		// Guard against recursive self-references (e.g., PATH=$(HOME)/bin:$(PATH))
		if seen[content] {
			break
		}
		seen[content] = true

		// Look up variable and replace
		replacement := ""
		if val, ok := vars[content]; ok {
			replacement = val
		}
		result = result[:dollar] + replacement + result[close+1:]
		pos = dollar // re-scan to catch new expansions in replacement
	}
	return result
}

// evalGnuMakeFunc evaluates a known GNU Make built-in function.
// It first recursively expands variables inside the arguments (matching GNU Make's
// inside-out evaluation order), then evaluates the function.
// Returns the evaluated result and whether evaluation was successful.
func evalGnuMakeFunc(funcName, content string, vars map[string]string) (string, bool) {
	// Expand variables inside arguments first (GNU Make evaluates inside-out)
	expanded := ExpandVariable(content, vars)

	// Strip function name + whitespace to get arguments
	args := strings.TrimSpace(expanded[len(funcName):])

	switch funcName {
	case "subst":
		return evalSubst(args), true
	case "dir":
		return evalDir(args), true
	default:
		return "", false
	}
}

// evalSubst evaluates $(subst from,to,text).
// Arguments are comma-separated with depth awareness.
func evalSubst(args string) string {
	parts := splitArgs(args)
	if len(parts) < 3 {
		return ""
	}
	return strings.ReplaceAll(parts[2], parts[0], parts[1])
}

// evalDir evaluates $(dir names...).
// Returns the directory portion of each path (handles / and \).
// Handles Windows paths with spaces like "C:\Program Files\..." correctly
// by first splitting on newlines (from multi-line shell output), then detecting
// drive-prefixed paths within each line.
func evalDir(args string) string {
	// Split by newlines first to handle multi-line shell output
	// This preserves Windows paths with spaces across line boundaries.
	dirs := splitDirArgs(args)
	if len(dirs) == 0 {
		return ""
	}
	var result []string
	for _, p := range dirs {
		idx := strings.LastIndexAny(p, "/\\")
		if idx >= 0 {
			result = append(result, p[:idx+1])
		} else {
			result = append(result, "./")
		}
	}
	return strings.Join(result, " ")
}

// splitDirArgs splits $(dir ...) arguments into individual paths.
//	1. Split by newlines first (multi-line shell output)
//	2. Within each line, detect Windows drive prefixes to keep paths intact
//	3. Space-split non-Windows paths as normal
func splitDirArgs(s string) []string {
	var paths []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		paths = append(paths, splitLineIntoPaths(line)...)
	}
	return paths
}

// splitLineIntoPaths splits a single line of $(dir ...) arguments into paths.
// Paths starting with a Windows drive prefix (e.g., "C:\") are kept intact;
// other tokens are split by whitespace as normal GNU Make behavior.
func splitLineIntoPaths(line string) []string {
	tokens := strings.Fields(line)
	if len(tokens) == 0 {
		return nil
	}
	var paths []string
	buf := ""
	for _, tok := range tokens {
		if isWindowsDrivePrefix(tok) {
			if buf != "" {
				paths = append(paths, buf)
			}
			buf = tok
		} else if buf != "" {
			// Continuation of a Windows path broken by whitespace
			buf += " " + tok
		} else {
			paths = append(paths, tok)
		}
	}
	if buf != "" {
		paths = append(paths, buf)
	}
	return paths
}

// isWindowsDrivePrefix checks if s starts with a drive letter followed by ":\" or ":/".
func isWindowsDrivePrefix(s string) bool {
	if len(s) < 3 {
		return false
	}
	c := s[0]
	isLetter := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
	return isLetter && s[1] == ':' && (s[2] == '/' || s[2] == '\\')
}

// splitArgs splits a string by commas at depth 0, respecting nesting via () and {}.
func splitArgs(s string) []string {
	var args []string
	depth := 0
	start := 0
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '(' || s[i] == '{':
			depth++
		case s[i] == ')' || s[i] == '}':
			depth--
		case s[i] == ',' && depth == 0:
			args = append(args, strings.TrimSpace(s[start:i]))
			start = i + 1
		}
	}
	if start <= len(s) {
		args = append(args, strings.TrimSpace(s[start:]))
	}
	return args
}
