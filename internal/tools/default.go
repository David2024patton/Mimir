package tools

import "github.com/David2024patton/Mimir/internal/sdlc"

// Default builds the native tool registry (the built-in, lean toolset).
// Small-model mode (F31) uses exactly this focused set: bash, read/write/list, todo.
func Default(workDir string) *Registry {
	r := NewRegistry()
	r.Register(&ShellTool{WorkDir: workDir})
	r.Register(&ReadFileTool{})
	r.Register(&WriteFileTool{})
	r.Register(&EditBlockTool{})
	r.Register(&ListDirectoryTool{})
	r.Register(&TodoTool{Store: NewTodoStore()})
	r.Register(&QuestionTool{})
	r.Register(&SkillTool{})
	r.Register(&IngestRepoTool{WorkDir: workDir})
	r.Register(&SDLCTool{Manager: sdlc.NewManager(workDir)})
	return r
}

