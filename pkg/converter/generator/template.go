package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// TemplateData holds all data needed to render the magefile template.
type TemplateData struct {
	PackageName string
	Imports     []string
	Constants   []ConstDef
	Vars        []VarDef
	Functions   []FuncDef
	Aliases     []AliasDef
}

// ConstDef represents a Go const definition.
type ConstDef struct {
	Name  string
	Value string
}

// VarDef represents a Go var definition.
type VarDef struct {
	Name  string
	Value string
}

// FuncDef represents a mage target function.
type FuncDef struct {
	Name        string
	Description string
	Body        string
}

// AliasDef represents a mage alias mapping.
type AliasDef struct {
	Alias  string
	Target string
}

const magefileTemplate = `//go:build mage
// +build mage

package {{.PackageName}}

import (
{{range .Imports}}	"{{.}}"
{{end}})

{{range .Constants}}// {{.Name}} - defined from Makefile variable
const {{.Name}} = {{printf "%q" .Value}}
{{end}}
{{range .Vars}}// {{.Name}} - defined from Makefile variable
var {{.Name}} = {{printf "%q" .Value}}
{{end}}
{{range .Functions}}// {{.Name}} {{.Description}}
func {{.Name}}() error {
{{.Body}}
}
{{end}}
`

// renderTemplate renders the magefile template with the given data.
func renderTemplate(data *TemplateData) (string, error) {
	funcMap := template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}

	tmpl, err := template.New("magefile").Funcs(funcMap).Parse(magefileTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	result := buf.String()

	// Post-process: remove extra blank lines from empty sections
	result = cleanupOutput(result)

	return result, nil
}

// cleanupOutput removes consecutive blank lines from generated output.
func cleanupOutput(s string) string {
	lines := strings.Split(s, "\n")
	var cleaned []string
	prevBlank := false

	for _, line := range lines {
		isBlank := strings.TrimSpace(line) == ""
		if isBlank && prevBlank {
			continue
		}
		cleaned = append(cleaned, line)
		prevBlank = isBlank
	}

	return strings.Join(cleaned, "\n")
}
