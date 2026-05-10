package parser

import (
	"fmt"
	"strings"
)

// ParseError represents an error during Makefile parsing.
type ParseError struct {
	Line    int
	Col     int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d:%d: %s", e.Line, e.Col, e.Message)
}

// Parser converts Makefile tokens into an AST.
type Parser struct {
	input    string
	lexer    *Lexer
	tokens   []Token
	pos      int
	makefile *Makefile
	vars     map[string]string // variable lookup table
}

// NewParser creates a new Makefile parser.
func NewParser(input string) *Parser {
	return &Parser{
		input:    input,
		lexer:    NewLexer(input, "<input>"),
		makefile: &Makefile{},
		vars:     make(map[string]string),
	}
}

// Parse parses the Makefile and returns the AST.
func (p *Parser) Parse() (*Makefile, error) {
	// First pass: tokenize everything
	p.tokens = TokenizeAll(p.input)

	// Second pass: build AST
	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]

		switch tok.Type {
		case TokenEOF:
			p.pos++
			continue

		case TokenNewline:
			p.pos++
			continue

		case TokenComment:
			p.makefile.Comments = append(p.makefile.Comments, tok.Value)
			p.pos++
			continue

		case TokenTarget:
			target, err := p.parseTarget()
			if err != nil {
				return nil, err
			}
			p.makefile.Targets = append(p.makefile.Targets, target)

		case TokenVariableAssign:
			v, err := p.parseVariable()
			if err != nil {
				return nil, err
			}
			p.makefile.Variables = append(p.makefile.Variables, v)
			// Also store in lookup table for expansion
			p.vars[v.Name] = v.Value

		case TokenInclude:
			inc := p.parseInclude()
			p.makefile.Includes = append(p.makefile.Includes, inc)

		case TokenExport:
			exp := p.parseExport()
			p.makefile.Exports = append(p.makefile.Exports, exp)

		case TokenConditional:
			cond, err := p.parseConditional()
			if err != nil {
				return nil, err
			}
			p.makefile.Conditionals = append(p.makefile.Conditionals, cond)

		case TokenDefine:
			v, err := p.parseDefine()
			if err != nil {
				return nil, err
			}
			p.makefile.Variables = append(p.makefile.Variables, v)
			p.vars[v.Name] = v.Value

		default:
			// Skip other tokens at top level
			p.pos++
		}
	}

	return p.makefile, nil
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() {
	p.pos++
}

// parseTarget parses a target with its prerequisites and recipes.
func (p *Parser) parseTarget() (*Target, error) {
	targetTok := p.tokens[p.pos]
	// Split target line on ':' to get target name and prerequisites
	line := targetTok.Value
	parts := strings.SplitN(line, ":", 2)
	target := &Target{
		Name: strings.TrimSpace(parts[0]),
		Line: targetTok.Line,
	}
	if len(parts) > 1 {
		prereqText := strings.TrimSpace(parts[1])
		if prereqText != "" {
			target.Prerequisites = strings.Fields(prereqText)
		}
	}
	p.advance()

	// Skip newline
	if p.pos < len(p.tokens) && p.tokens[p.pos].Type == TokenNewline {
		p.advance()
	}

	// Collect recipes (tab-indented lines)
	for p.pos < len(p.tokens) && p.tokens[p.pos].Type == TokenRecipe {
		target.Recipes = append(target.Recipes, p.tokens[p.pos].Value)
		p.advance()
		// Skip newline after recipe
		if p.pos < len(p.tokens) && p.tokens[p.pos].Type == TokenNewline {
			p.advance()
		}
	}

	// Check if .PHONY
	if target.Name == ".PHONY" {
		target.IsPhony = true
	}

	return target, nil
}

// parseVariable parses a variable assignment.
func (p *Parser) parseVariable() (*Variable, error) {
	tok := p.tokens[p.pos]
	v := &Variable{
		Line: tok.Line,
	}

	val := tok.Value
	// Parse name, operator, value
	// Supported operators: =, :=, +=, ?=, !=
	operators := []string{":=", "+=", "?=", "!=", "="}
	for _, op := range operators {
		if idx := strings.Index(val, op); idx != -1 {
			v.Name = strings.TrimSpace(val[:idx])
			v.Operator = op
			v.Value = strings.TrimSpace(val[idx+len(op):])
			if op == "!=" {
				v.IsShell = true
			}
			break
		}
	}

	if v.Name == "" {
		return nil, &ParseError{
			Line:    tok.Line,
			Col:     tok.Col,
			Message: "invalid variable assignment: " + val,
		}
	}

	p.advance()
	return v, nil
}

// parseInclude parses an include directive.
func (p *Parser) parseInclude() *Include {
	tok := p.tokens[p.pos]
	inc := &Include{
		Line: tok.Line,
	}

	val := tok.Value
	inc.IsOptional = strings.HasPrefix(val, "-include") || strings.HasPrefix(val, "sinclude")

	// Extract paths
	parts := strings.Fields(val)
	if len(parts) > 1 {
		inc.Paths = parts[1:]
	}

	p.advance()
	return inc
}

// parseExport parses an export/unexport directive.
func (p *Parser) parseExport() *Export {
	tok := p.tokens[p.pos]
	exp := &Export{
		Line:     tok.Line,
		IsExport: strings.HasPrefix(tok.Value, "export"),
	}

	// Extract variable name
	parts := strings.Fields(tok.Value)
	if len(parts) > 1 {
		exp.Variable = parts[1]
	}

	p.advance()
	return exp
}

// parseConditional parses an ifeq/ifneq/ifdef/ifndef block.
func (p *Parser) parseConditional() (*Conditional, error) {
	tok := p.tokens[p.pos]
	cond := &Conditional{
		Line: tok.Line,
	}

	val := tok.Value
	parts := strings.Fields(val)
	if len(parts) < 2 {
		return nil, &ParseError{
			Line:    tok.Line,
			Message: "invalid conditional: " + val,
		}
	}

	cond.Kind = parts[0]
	cond.Cond = strings.Join(parts[1:], " ")
	p.advance()

	// Collect true body until else or endif
	inElse := false
	for p.pos < len(p.tokens) {
		t := p.tokens[p.pos]
		if t.Type == TokenConditional {
			condVal := strings.TrimSpace(t.Value)
			if condVal == "else" {
				inElse = true
				p.advance()
				continue
			}
			if condVal == "endif" {
				p.advance()
				break
			}
		}

		line := t.Value
		if t.Type == TokenRecipe || t.Type == TokenText {
			line = t.Value
		}
		if inElse {
			cond.FalseBody = append(cond.FalseBody, line)
		} else {
			cond.TrueBody = append(cond.TrueBody, line)
		}
		p.advance()
	}

	return cond, nil
}

// parseDefine parses a define/endef multi-line variable.
// Reads raw lines between define and endef to avoid tokenization issues.
func (p *Parser) parseDefine() (*Variable, error) {
	tok := p.tokens[p.pos]
	name := strings.TrimSpace(tok.Value)
	p.advance()

	// Skip newline
	if p.pos < len(p.tokens) && p.tokens[p.pos].Type == TokenNewline {
		p.advance()
	}

	// Collect raw lines until endef
	var bodyLines []string
	for p.pos < len(p.tokens) {
		t := p.tokens[p.pos]
		if t.Type == TokenEndef {
			p.advance()
			break
		}
		// Each line in the define body is a separate token
		// (could be TokenText, TokenTarget, TokenRecipe, etc.)
		// We reconstruct the line from the original input
		bodyLines = append(bodyLines, t.Value)
		p.advance()
		// Skip newlines between lines
		if p.pos < len(p.tokens) && p.tokens[p.pos].Type == TokenNewline {
			p.advance()
		}
	}

	return &Variable{
		Name:     name,
		Operator: "=",
		Value:    strings.Join(bodyLines, "\n"),
		Line:     tok.Line,
	}, nil
}

// MakefileString returns the Makefile AST as a readable string (for debugging).
func MakefileString(m *Makefile) string {
	var sb strings.Builder

	sb.WriteString("Variables:\n")
	for _, v := range m.Variables {
		sb.WriteString(fmt.Sprintf("  %s %s %s\n", v.Name, v.Operator, v.Value))
	}

	sb.WriteString("Targets:\n")
	for _, t := range m.Targets {
		phony := ""
		if t.IsPhony {
			phony = " (.PHONY)"
		}
		sb.WriteString(fmt.Sprintf("  %s:%s\n", t.Name, phony))
		for _, p := range t.Prerequisites {
			sb.WriteString(fmt.Sprintf("    prereq: %s\n", p))
		}
		for _, r := range t.Recipes {
			sb.WriteString(fmt.Sprintf("    recipe: %s\n", r))
		}
	}

	sb.WriteString("Includes:\n")
	for _, i := range m.Includes {
		sb.WriteString(fmt.Sprintf("  %v (optional=%v)\n", i.Paths, i.IsOptional))
	}

	sb.WriteString("Conditionals:\n")
	for _, c := range m.Conditionals {
		sb.WriteString(fmt.Sprintf("  %s %s\n", c.Kind, c.Cond))
	}

	return sb.String()
}
