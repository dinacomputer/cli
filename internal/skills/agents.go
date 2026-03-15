package skills

// Agent represents a supported AI agent/tool that can use agent skills.
type Agent struct {
	Name     string
	Slug     string
	// ProjectDir is the project-level skills directory (relative to project root).
	ProjectDir string
	// UserDir is the user-level skills directory (relative to home).
	UserDir string
}

// SupportedAgents lists all AI agents the Dina CLI skill can be installed for.
var SupportedAgents = []Agent{
	{
		Name:       "Claude Code",
		Slug:       "claude-code",
		ProjectDir: ".claude/skills",
		UserDir:    ".claude/skills",
	},
	{
		Name:       "Cursor",
		Slug:       "cursor",
		ProjectDir: ".cursor/skills",
		UserDir:    ".cursor/skills",
	},
	{
		Name:       "VS Code (Copilot)",
		Slug:       "vscode",
		ProjectDir: ".vscode/skills",
		UserDir:    ".vscode/skills",
	},
	{
		Name:       "Gemini CLI",
		Slug:       "gemini-cli",
		ProjectDir: ".gemini/skills",
		UserDir:    ".gemini/skills",
	},
	{
		Name:       "OpenAI Codex",
		Slug:       "codex",
		ProjectDir: ".codex/skills",
		UserDir:    ".codex/skills",
	},
	{
		Name:       "Any compatible agent (.agents/skills)",
		Slug:       "generic",
		ProjectDir: ".agents/skills",
		UserDir:    ".agents/skills",
	},
}
