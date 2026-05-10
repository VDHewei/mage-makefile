package script

// EngineType represents the type of script engine.
type EngineType string

const (
	EngineGo  EngineType = "go"
	EngineLua EngineType = "lua"
	EngineJS  EngineType = "js"
)

// ScriptEngine defines the interface for executing scripts in different languages.
type ScriptEngine interface {
	Execute(script string, env map[string]string) (string, error)
	Validate(script string) error
	Name() string
	Type() EngineType
}
