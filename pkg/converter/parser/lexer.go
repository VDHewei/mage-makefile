// Package parser provides a lexer, parser, and AST for GNU Makefiles.
package parser

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType represents the type of a lexical token.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenComment
	TokenTarget
	TokenPrerequisite
	TokenRecipe
	TokenVariableAssign
	TokenVariableRef
	TokenFunction
	TokenConditional
	TokenInclude
	TokenExport
	TokenDefine
	TokenEndef
	TokenVPath
	TokenOverride
	TokenNewline
	TokenText
	TokenShell
	TokenError
)

// Token represents a lexical token with its type, value, and position.
type Token struct {
	Type     TokenType
	Value    string
	Line     int
	Col      int
	FileName string
}

// String returns a human-readable representation of the token.
func (t Token) String() string {
	typeNames := map[TokenType]string{
		TokenEOF:            "EOF",
		TokenComment:        "Comment",
		TokenTarget:         "Target",
		TokenPrerequisite:   "Prerequisite",
		TokenRecipe:         "Recipe",
		TokenVariableAssign: "VariableAssign",
		TokenVariableRef:    "VariableRef",
		TokenFunction:       "Function",
		TokenConditional:    "Conditional",
		TokenInclude:        "Include",
		TokenExport:         "Export",
		TokenDefine:         "Define",
		TokenEndef:          "Endef",
		TokenVPath:          "VPath",
		TokenOverride:       "Override",
		TokenNewline:        "Newline",
		TokenText:           "Text",
		TokenShell:          "Shell",
		TokenError:          "Error",
	}
	return fmt.Sprintf("%s(%q)@%d:%d", typeNames[t.Type], t.Value, t.Line, t.Col)
}

// Lexer performs lexical analysis on Makefile content.
type Lexer struct {
	input    []rune
	pos      int
	line     int
	col      int
	fileName string
}

// NewLexer creates a new lexer for the given input.
func NewLexer(input, fileName string) *Lexer {
	return &Lexer{
		input:    []rune(input),
		pos:      0,
		line:     1,
		col:      1,
		fileName: fileName,
	}
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	if l.pos >= len(l.input) {
		return l.makeToken(TokenEOF, "")
	}

	ch := l.input[l.pos]

	// Handle newlines
	if ch == '\n' {
		tok := l.makeToken(TokenNewline, "\n")
		l.advance()
		return tok
	}

	// Handle comments
	if ch == '#' {
		return l.lexComment()
	}

	// Handle tab-indented recipe lines
	if ch == '\t' {
		return l.lexRecipe()
	}

	// Handle spaces (skip but after targets they separate prerequisites)
	// The parser handles this context; we just emit text tokens

	// Handle shell command prefix (!=)
	if ch == '!' && l.peek() == '=' {
		return l.lexShellAssign()
	}

	// Handle include directive
	if strings.HasPrefix(string(l.input[l.pos:]), "include ") ||
		strings.HasPrefix(string(l.input[l.pos:]), "-include ") ||
		strings.HasPrefix(string(l.input[l.pos:]), "sinclude ") {
		return l.lexInclude()
	}

	// Handle conditional directives
	if l.matchKeyword("ifeq") || l.matchKeyword("ifneq") ||
		l.matchKeyword("ifdef") || l.matchKeyword("ifndef") ||
		l.matchKeyword("else") || l.matchKeyword("endif") {
		return l.lexConditional()
	}

	// Handle define/endef
	if l.matchKeyword("define ") {
		return l.lexDefine()
	}
	if l.matchKeyword("endef") {
		return l.lexEndef()
	}

	// Handle export/unexport
	if l.matchKeyword("export ") || l.matchKeyword("unexport ") || l.matchKeyword("export") {
		return l.lexExport()
	}

	// Handle vpath
	if l.matchKeyword("vpath ") || l.matchKeyword("vpath") {
		return l.lexVPath()
	}

	// Handle override
	if l.matchKeyword("override ") {
		return l.lexOverride()
	}

	// Check for target (text followed by ':')
	// We look ahead to see if there's a ':' in the current line
	if l.isTargetLine() {
		return l.lexTarget()
	}

	// Check for variable assignment (text followed by =, :=, +=, ?=)
	if l.isVariableAssignment() {
		return l.lexVariableAssign()
	}

	// Check for variable reference $(...) or ${...}
	if ch == '$' {
		return l.lexVariableRef()
	}

	// Default: text token
	return l.lexText()
}

func (l *Lexer) makeToken(t TokenType, val string) Token {
	return Token{
		Type:     t,
		Value:    val,
		Line:     l.line,
		Col:      l.col,
		FileName: l.fileName,
	}
}

func (l *Lexer) advance() {
	if l.pos >= len(l.input) {
		return
	}
	ch := l.input[l.pos]
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	l.pos++
}

func (l *Lexer) peek() rune {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && l.input[l.pos] != '\n' &&
		(l.input[l.pos] == ' ' || l.input[l.pos] == '\r') {
		l.advance()
	}
}

func (l *Lexer) skipToEndOfLine() {
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.advance()
	}
}

func (l *Lexer) matchKeyword(keyword string) bool {
	rest := string(l.input[l.pos:])
	return strings.HasPrefix(rest, keyword)
}

func (l *Lexer) matchKeywordLine(keyword string) bool {
	rest := string(l.input[l.pos:])
	if !strings.HasPrefix(rest, keyword) {
		return false
	}
	// Must be at the start of a line (after optional whitespace)
	// Check: position must be at line start
	return l.col == 1 || (l.col > 1 && l.input[l.pos-1] == '\t')
}

func (l *Lexer) readUntil(pred func(rune) bool) string {
	start := l.pos
	for l.pos < len(l.input) && pred(l.input[l.pos]) {
		l.advance()
	}
	return string(l.input[start:l.pos])
}

func (l *Lexer) readLine() string {
	return l.readUntil(func(r rune) bool {
		return r != '\n'
	})
}

// isTargetLine checks if the current line looks like a target definition.
func (l *Lexer) isTargetLine() bool {
	if l.col != 1 {
		return false
	}
	// Look for pattern: text...: text...
	rest := string(l.input[l.pos:])
	// Find ':' before newline
	colonIdx := strings.IndexByte(rest, ':')
	newlineIdx := strings.IndexByte(rest, '\n')
	if colonIdx == -1 {
		return false
	}
	if newlineIdx != -1 && colonIdx > newlineIdx {
		return false
	}
	// Check if any variable assignment operator (=, :=, +=, ?=, !=)
	// appears BEFORE the colon. If so, this is a variable assignment, not a target.
	// This handles cases like: IMAGE ?= python:3.11-slim
	beforeColon := rest[:colonIdx]
	if hasAssignmentOp(beforeColon) {
		return false
	}
	// Exclude := (colon-equal) pattern like python:3.11-slim
	afterColon := rest[colonIdx+1:]
	if len(afterColon) > 0 && afterColon[0] == '=' {
		return false
	}
	return true
}

// hasAssignmentOp checks if text contains a makefile variable assignment operator.
func hasAssignmentOp(s string) bool {
	ops := []string{":=", "+=", "?=", "!="}
	for _, op := range ops {
		if strings.Contains(s, op) {
			return true
		}
	}
	// Plain '=' but not part of ':='
	if strings.Contains(s, " =") || strings.HasSuffix(s, "=") {
		return true
	}
	return false
}

// isVariableAssignment checks if the current line is a variable assignment.
func (l *Lexer) isVariableAssignment() bool {
	rest := string(l.input[l.pos:])
	// Variable name must be valid (alphanumeric + _)
	// Look for =, :=, +=, ?=, != operators
	for i, ch := range rest {
		if ch == '\n' {
			return false
		}
		if ch == '=' || ch == ':' || ch == '+' || ch == '?' || ch == '!' {
			if ch == ':' && i+1 < len(rest) && rest[i+1] == '=' {
				return true // :=
			}
			if ch == '+' && i+1 < len(rest) && rest[i+1] == '=' {
				return true // +=
			}
			if ch == '?' && i+1 < len(rest) && rest[i+1] == '=' {
				return true // ?=
			}
			if ch == '!' && i+1 < len(rest) && rest[i+1] == '=' {
				return true // !=
			}
			if ch == '=' {
				return true // =
			}
			// If we see ':' and it's not := or : at start, it's a target, not assignment
			if ch == ':' {
				return false
			}
		}
	}
	return false
}

func (l *Lexer) lexComment() Token {
	l.advance() // skip #
	comment := l.readLine()
	return l.makeToken(TokenComment, strings.TrimSpace(comment))
}

func (l *Lexer) lexRecipe() Token {
	l.advance() // skip \t
	cmd := l.readLine()
	return l.makeToken(TokenRecipe, strings.TrimSpace(cmd))
}

func (l *Lexer) lexTarget() Token {
	line := l.readLine()
	// Return full line including prerequisites so the parser can split
	return l.makeToken(TokenTarget, line)
}

func (l *Lexer) lexVariableAssign() Token {
	line := l.readLine()
	return l.makeToken(TokenVariableAssign, strings.TrimSpace(line))
}

func (l *Lexer) lexVariableRef() Token {
	l.advance() // skip $
	if l.pos >= len(l.input) {
		return l.makeToken(TokenVariableRef, "$")
	}

	ch := l.input[l.pos]
	if ch == '(' || ch == '{' {
		open := ch
		close := byte(')')
		if open == '{' {
			close = '}'
		}
		l.advance() // skip ( or {
		ref := l.readUntil(func(r rune) bool {
			return r != rune(close) && r != '\n'
		})
		if l.pos < len(l.input) && l.input[l.pos] == rune(close) {
			l.advance() // skip ) or }
		}
		return l.makeToken(TokenVariableRef, string(open)+ref+string(close))
	}

	// Single character variable like $@, $<, $^
	ref := string(ch)
	l.advance()
	return l.makeToken(TokenVariableRef, ref)
}

func (l *Lexer) lexInclude() Token {
	val := l.readLine()
	return l.makeToken(TokenInclude, strings.TrimSpace(val))
}

func (l *Lexer) lexConditional() Token {
	val := l.readLine()
	return l.makeToken(TokenConditional, strings.TrimSpace(val))
}

func (l *Lexer) lexDefine() Token {
	l.advancePast("define ") // 7 chars
	name := l.readLine()
	return l.makeToken(TokenDefine, name)
}

func (l *Lexer) lexEndef() Token {
	l.advancePast("endef") // 5 chars
	l.skipToEndOfLine()
	return l.makeToken(TokenEndef, "endef")
}

func (l *Lexer) lexExport() Token {
	val := l.readLine()
	return l.makeToken(TokenExport, strings.TrimSpace(val))
}

func (l *Lexer) lexVPath() Token {
	val := l.readLine()
	return l.makeToken(TokenVPath, strings.TrimSpace(val))
}

func (l *Lexer) lexOverride() Token {
	val := l.readLine()
	return l.makeToken(TokenOverride, strings.TrimSpace(val))
}

func (l *Lexer) lexShellAssign() Token {
	val := l.readLine()
	return l.makeToken(TokenShell, strings.TrimSpace(val))
}

func (l *Lexer) lexText() Token {
	val := l.readLine()
	return l.makeToken(TokenText, strings.TrimSpace(val))
}

func (l *Lexer) advancePast(s string) {
	for i := 0; i < len(s) && l.pos < len(l.input); i++ {
		l.advance()
	}
}

// TokenizeAll returns all tokens from the input (useful for debugging/testing).
func TokenizeAll(input string) []Token {
	l := NewLexer(input, "<input>")
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return tokens
}

func isAlphaNum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.'
}
